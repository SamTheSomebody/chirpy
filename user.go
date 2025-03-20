package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/samthesomebody/chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func mapToUser(from database.User) User {
	return User{
		ID:        from.ID,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
		Email:     from.Email,
	}
}

func addUser(w http.ResponseWriter, req *http.Request) {
	var user database.User
	err := json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}

	user, err = apiCfg.DB.CreateUser(req.Context(), user.Email)
	if err != nil {
		log.Fatal(err)
	}

	respondWithJSON(w, 201, mapToUser(user))
}

func removeUsers(w http.ResponseWriter, req *http.Request) {
	if apiCfg.Platform != "dev" {
		respondWithError(w, 403, "Invalid Permissions")
		return
	}

	err := apiCfg.DB.RemoveUsers(req.Context())
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(200)
}
