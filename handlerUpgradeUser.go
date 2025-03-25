package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/samthesomebody/chirpy/internal/auth"
)

func handlerPaymentWebhook(w http.ResponseWriter, req *http.Request) {
	apiKey, err := auth.GetApiKey(req.Header)
	if err != nil || apiKey != apiCfg.PolkaKey {
		respondWithError(w, http.StatusUnauthorized, "Forbidden")
		return
	}

	var body struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		}
	}

	err = json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Incorrect body parameters")
		return
	}

	if body.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	id, err := uuid.Parse(body.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Incorrect body parameters")
		return
	}

	_, err = apiCfg.DB.UpgradeUserToRed(req.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Printf("Error upgrading user to red in database: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully upgraded user [%v] to red \n", id)
	w.WriteHeader(http.StatusNoContent)
}
