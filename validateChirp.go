package main

import (
  "encoding/json"
  "net/http"
  "strings"
  "log"
)

func validateChirp(w http.ResponseWriter, r *http.Request) {
  type chirp struct {
    Body string `json:"body"`
  }

  decoder := json.NewDecoder(r.Body)
  c := chirp{}
  err := decoder.Decode(&c)
  if err != nil {
    log.Printf("Error decoding parameters: %s", err)
    w.WriteHeader(500)
    return
  }

  w.Header().Set("Content-Type", "application/json")

  if len(c.Body) > 140 {
    respondWithError(w, 400, "Chirp is too long.")
    return
  }
  type cleanedChirp struct {
    CleanedBody string `json:"cleaned_body"`
  }
  respondWithJSON(w, 200, cleanedChirp{replaceProfanities(c.Body)})
}


func replaceProfanities(msg string) string {
  profanities := []string{"kerfuffle", "sharbert", "fornax"}
  words := strings.Split(msg, " ")
  for i, word := range(words) {
    for _, profanity := range(profanities) {
      if strings.ToLower(word) == profanity {
        words[i] = "****"
        break
      }
    }
  }
  return strings.Join(words, " ")
}
