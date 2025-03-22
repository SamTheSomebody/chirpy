package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
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
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func mapToUser(from database.User) User {
	return User{
		ID:        from.ID,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
		Email:     from.Email,
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
		log.Printf("Error decoding request body: %v", err)
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

	expiresIn, err := time.ParseDuration(strconv.Itoa(details.ExpiresInSeconds) + "s")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not parse expiration durations")
		return
	}
	token, err := auth.MakeJWT(userDB.ID, apiCfg.TokenSecret, expiresIn)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not generate jwt")
		return
	}
	w.WriteHeader(http.StatusOK)
	user := mapToUser(userDB)
	user.Token = token
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
