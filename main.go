package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/samthesomebody/chirpy/internal/database"
)

var apiCfg *apiConfig

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	tokenSecret := os.Getenv("TOKEN_SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := *database.New(db)
	apiCfg = &apiConfig{
		DB:          dbQueries,
		Platform:    platform,
		TokenSecret: tokenSecret,
	}

	mux := http.NewServeMux()
	handlerServeSite := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handlerServeSite))
	mux.HandleFunc("GET /api/healthz", handlerGetHealth)
	mux.HandleFunc("GET /admin/metrics", apiCfg.getFileserverHits)
	mux.HandleFunc("POST /admin/reset", handlerRemoveUsers)
	mux.HandleFunc("POST /api/users", handlerAddUser)
	mux.HandleFunc("POST /api/login", handlerLoginUser)
	mux.HandleFunc("POST /api/chirps", handlerAddChirp)
	mux.HandleFunc("GET /api/chirps", handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", handlerGetChirp)

	server := &http.Server{}
	server.Addr = ":8080"
	server.Handler = mux

	log.Fatal(server.ListenAndServe())
}

func handlerGetHealth(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
