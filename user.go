package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/samthesomebody/chirpy/internal/auth"
	"github.com/samthesomebody/chirpy/internal/database"
)

type LoginDetails struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

func mapToUser(from database.User) User {
	return User{
		ID:          from.ID,
		CreatedAt:   from.CreatedAt,
		UpdatedAt:   from.UpdatedAt,
		Email:       from.Email,
		IsChirpyRed: from.IsChirpyRed,
	}
}

func handlerAddUser(w http.ResponseWriter, req *http.Request) {
	var details LoginDetails
	err := json.NewDecoder(req.Body).Decode(&details)
	if err != nil {
		log.Printf("Error decoding request body: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	password, err := auth.HashPassword(details.Password)
	if err != nil {
		log.Printf("Error hasing password: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	params := database.CreateUserParams{
		Email:          details.Email,
		HashedPassword: password,
	}
	userDB, err := apiCfg.DB.CreateUser(req.Context(), params)
	if err != nil {
		log.Fatal(err)
	}

	user := mapToUser(userDB)
	respondWithJSON(w, http.StatusCreated, user)
}

func handlerLoginUser(w http.ResponseWriter, req *http.Request) {
	var details LoginDetails
	err := json.NewDecoder(req.Body).Decode(&details)
	if err != nil {
		log.Printf("Error decoding request body: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if details.ExpiresInSeconds == 0 || details.ExpiresInSeconds > 3600 {
		details.ExpiresInSeconds = 3600
	}

	userDB, err := apiCfg.DB.GetUserByEmail(req.Context(), details.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
			return
		}
		log.Printf("Error retreiving user: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = auth.CheckPasswordHash(details.Password, userDB.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	token, err := auth.MakeJWT(userDB.ID, apiCfg.TokenSecret)
	if err != nil {
		log.Printf("Error generating JWT: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	refresh_token, _ := auth.MakeRefreshToken() // err is always nil
	params := database.CreateRefreshTokenParams{
		Token:     refresh_token,
		UserID:    userDB.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	}
	_, err = apiCfg.DB.CreateRefreshToken(req.Context(), params)
	if err != nil {
		log.Printf("Error generating refresh token database entry: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	user := mapToUser(userDB)
	user.Token = token
	user.RefreshToken = refresh_token
	respondWithJSON(w, http.StatusCreated, user)
}

func handlerRemoveUsers(w http.ResponseWriter, req *http.Request) {
	if apiCfg.Platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Invalid Permissions")
		return
	}

	err := apiCfg.DB.RemoveUsers(req.Context())
	if err != nil {
		log.Printf("Error removing user: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handlerRefreshJWT(w http.ResponseWriter, req *http.Request) {
	if req.Body != http.NoBody {
		log.Println("Error refreshing JWT: request has a body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Error finding refresh token in request header: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := apiCfg.DB.GetUserFromRefreshToken(req.Context(), token)
	if err != nil {
		log.Printf("Error finding refresh token database entry: %v\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return

	}

	jwt, err := auth.MakeJWT(userID, apiCfg.TokenSecret)
	if err != nil {
		log.Printf("Error generating JWT: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var user User
	user.Token = jwt
	body, err := json.Marshal(user)
	if err != nil {
		log.Printf("Error marshalling jwt token: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func handlerRevokeRefreshToken(w http.ResponseWriter, req *http.Request) {
	if req.Body != http.NoBody {
		log.Println("Error refreshing JWT: request has a body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Error finding refresh token in request header: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = apiCfg.DB.RevokeRefreshToken(req.Context(), token)
	if err != nil {
		log.Printf("Error adjusting refresh token db entry: %v\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handlerUpdateUser(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Error getting refresh token: %v\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(token, apiCfg.TokenSecret)
	if err != nil {
		log.Printf("Error validating JWT: %v\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var user database.UpdateUserLoginParams
	err = json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		log.Printf("Error decoding refresh token: %v/n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user.HashedPassword, err = auth.HashPassword(user.HashedPassword)
	if err != nil {
		log.Printf("Error hashing password: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user.ID = userID
	userDB, err := apiCfg.DB.UpdateUserLogin(req.Context(), user)
	if err != nil {
		log.Printf("Error updating user login details: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, 200, mapToUser(userDB))
}
