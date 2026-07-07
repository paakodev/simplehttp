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
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	DB             *sql.DB
	platform       string
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type ChirpResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
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
	platform := os.Getenv("PLATFORM")

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
		platform:       platform,
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
	mux.Handle("POST /api/chirps", chain(
		http.HandlerFunc(apiCfg.chirpsHandler),
		middlewareLog,
	))
	mux.Handle("POST /api/users", chain(
		http.HandlerFunc(apiCfg.createUser),
		middlewareLog,
	))

	log.Printf("Starting server on %s", server.Addr)
	server.ListenAndServe()
}

func validateChirp(chirp string) (string, error) {
	if len(chirp) > 140 {
		return "", fmt.Errorf("chirp body exceeds 140 characters")
	}

	return cleanChirpBody(chirp), nil
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

func (c *apiConfig) chirpsHandler(w http.ResponseWriter, r *http.Request) {
	type Chirp struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	decoder := json.NewDecoder(r.Body)
	chirpData := Chirp{}
	err := decoder.Decode(&chirpData)
	if err != nil {
		respondWithJSON(w,
			http.StatusBadRequest,
			map[string]string{"error": "Invalid request body"},
		)
		return
	}
	validatedBody, err := validateChirp(chirpData.Body)
	if err != nil {
		respondWithJSON(w,
			http.StatusBadRequest,
			map[string]string{"error": err.Error()},
		)
		return
	}
	newChirp, err := c.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		ID:     uuid.New(),
		Body:   validatedBody,
		UserID: chirpData.UserID,
	})
	if err != nil {
		respondWithJSON(w,
			http.StatusInternalServerError,
			map[string]string{"error": "Failed to create chirp"},
		)
		return
	}
	chirpResponse := ChirpResponse{
		ID:        newChirp.ID,
		CreatedAt: newChirp.CreatedAt,
		UpdatedAt: newChirp.UpdatedAt,
		Body:      newChirp.Body,
		UserID:    newChirp.UserID,
	}
	respondWithJSON(w, http.StatusCreated, chirpResponse)
}

func (c *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type NewUser struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	body := NewUser{}
	err := decoder.Decode(&body)
	if err != nil {
		respondWithJSON(w,
			http.StatusBadRequest,
			map[string]string{"error": "Invalid request body"},
		)
		return
	}

	newUser, err := c.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		ID:    uuid.New(),
		Email: body.Email,
	})
	if err != nil {
		respondWithJSON(w,
			http.StatusInternalServerError,
			map[string]string{"error": "Failed to create user"},
		)
		return
	}

	userResponse := UserResponse{
		ID:        newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email:     newUser.Email,
	}

	respondWithJSON(w, http.StatusCreated, userResponse)
}

// Helpers
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

// Hits handlers
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

// XXX: This also resets the users...
func (c *apiConfig) resetHits(w http.ResponseWriter, r *http.Request) {
	if c.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Not allowed")
		return
	}
	c.fileserverHits.Store(0)
	c.dbQueries.ResetUsers(r.Context())
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits and users reset"))
}

// Middleware functions
func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (c *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
