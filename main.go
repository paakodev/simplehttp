package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"simplehttp/internal/database"
	"strings"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	DB             *sql.DB
}

func (c *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

type middleware func(http.Handler) http.Handler

func chain(h http.Handler, middlewares ...middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries:      database.New(db),
		DB:             db,
	}

	mux.Handle("/app/", http.StripPrefix("/app/", chain(
		http.FileServer(http.Dir(".")),
		apiCfg.middlewareMetricsInc,
		middlewareLog,
	)))
	mux.Handle("GET /api/healthz", chain(
		http.HandlerFunc(healthz),
		middlewareLog,
	))
	mux.Handle("GET /admin/metrics", chain(
		http.HandlerFunc(apiCfg.getHits),
		middlewareLog,
	))
	mux.Handle("POST /admin/reset", chain(
		http.HandlerFunc(apiCfg.resetHits),
		middlewareLog,
	))
	mux.Handle("POST /api/validate_chirp", chain(
		http.HandlerFunc(validateChirp),
		middlewareLog,
	))

	log.Printf("Starting server on %s", server.Addr)
	server.ListenAndServe()
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type Chirp struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	body := Chirp{}
	err := decoder.Decode(&body)
	if err != nil {
		respondWithJSON(w,
			http.StatusBadRequest,
			map[string]string{"error": "Invalid request body"},
		)
		return
	}

	if len(body.Body) > 140 {
		respondWithJSON(w,
			http.StatusBadRequest,
			map[string]string{"error": "Chirp body exceeds 140 characters"},
		)
		return
	}

	cleanedBody := cleanChirpBody(body.Body)

	respondWithJSON(w, http.StatusOK, map[string]string{"cleaned_body": cleanedBody})
}

func cleanChirpBody(body string) string {
	badwords := []string{"kerfuffle", "sharbert", "fornax"}
	parts := make([]string, len(badwords))
	for i, word := range badwords {
		parts[i] = regexp.QuoteMeta(word)
	}
	re := regexp.MustCompile(`(?i)(` + strings.Join(parts, "|") + `)\b`)
	return re.ReplaceAllString(body, "****")
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}

func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (c *apiConfig) getHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	page := `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`
	w.Write(fmt.Appendf(nil, page, c.fileserverHits.Load()))
}

func (c *apiConfig) resetHits(w http.ResponseWriter, r *http.Request) {
	c.fileserverHits.Store(0)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset"))
}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
