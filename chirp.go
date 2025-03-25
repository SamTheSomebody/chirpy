package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/samthesomebody/chirpy/internal/auth"
	"github.com/samthesomebody/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func mapToChirp(from database.Chirp) Chirp {
	return Chirp{
		ID:        from.ID,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
		Body:      from.Body,
		UserID:    from.UserID,
	}
}

func handlerAddChirp(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Error getting authorization header: %v\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	id, err := auth.ValidateJWT(token, apiCfg.TokenSecret)
	if err != nil {
		log.Printf("Error validating JWT: %v [%v]\n", err, token)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var chirp Chirp
	err = json.NewDecoder(req.Body).Decode(&chirp)
	if err != nil {
		log.Printf("Error decoding parameters: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(chirp.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long.")
		return
	}
	chirp.Body = replaceProfanities(chirp.Body)

	params := database.CreateChirpParams{Body: chirp.Body, UserID: id}
	chirpDB, err := apiCfg.DB.CreateChirp(req.Context(), params)
	if err != nil {
		log.Printf("Error posting chirp: %+v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	respondWithJSON(w, http.StatusCreated, mapToChirp(chirpDB))
}

func replaceProfanities(msg string) string {
	profanities := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(msg, " ")
	for i, word := range words {
		if slices.Contains(profanities, strings.ToLower(word)) {
			words[i] = "****"
			break
		}
	}
	return strings.Join(words, " ")
}

func handlerGetChirps(w http.ResponseWriter, req *http.Request) {
	author_id := req.URL.Query().Get("author_id")
	var chirpsDB []database.Chirp
	var err error
	if author_id == "" {
		chirpsDB, err = apiCfg.DB.GetChirps(req.Context())
	} else {
		id, err := uuid.Parse(author_id)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid user id")
			return
		}
		chirpsDB, err = apiCfg.DB.GetChirpsByUser(req.Context(), id)
	}
	if err != nil {
		log.Printf("Error retreiving chirps: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sort_desc := req.URL.Query().Get("sort") == "desc"

	chirps := []Chirp{}
	for _, chirp := range chirpsDB {
		chirps = append(chirps, mapToChirp(chirp))
	}

	if sort_desc {
		sort.Slice(
			chirps,
			func(i, j int) bool { return chirps[j].CreatedAt.Before(chirps[i].CreatedAt) },
		)
	}
	respondWithJSON(w, http.StatusOK, chirps)
}

func handlerGetChirp(w http.ResponseWriter, req *http.Request) {
	id, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		log.Printf("Couldn't parse parameter: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	chirp, err := apiCfg.DB.GetChirp(req.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Chirp not found.")
			return
		}
		log.Printf("Error retreiving chirp: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, mapToChirp(chirp))
}

func handlerDeleteChirp(w http.ResponseWriter, req *http.Request) {
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

	chirpID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		log.Printf("Couldn't parse parameter: %v\n [%v]", err, chirpID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	chirp, err := apiCfg.DB.GetChirp(req.Context(), chirpID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Chirp not found.")
			return
		}
		log.Printf("Error retreiving chirp: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Forbidden")
		return
	}

	err = apiCfg.DB.RemoveChirp(req.Context(), chirpID)
	if err != nil {
		log.Printf("Error removing chirp: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
