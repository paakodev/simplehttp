package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"simplehttp/internal/auth"
	"simplehttp/internal/database"
	"time"

	"github.com/google/uuid"
)

func (c *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type NewUser struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	hashedPassword, err := auth.HashPassword(body.Password)
	if err != nil {
		respondWithJSON(w,
			http.StatusInternalServerError,
			map[string]string{"error": "Failed to hash password"},
		)
		return
	}

	newUser, err := c.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		ID:             uuid.New(),
		Email:          body.Email,
		HashedPassword: hashedPassword,
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

func (c *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	type LoginRequest struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		ExpiresIn *int64 `json:"expires_in,omitempty"` // Optional field for token expiration in seconds
	}
	decoder := json.NewDecoder(r.Body)
	loginReq := LoginRequest{}
	err := decoder.Decode(&loginReq)
	if err != nil {
		respondWithJSON(w,
			http.StatusBadRequest,
			map[string]string{"error": "Invalid request body"},
		)
		return
	}

	user, err := c.dbQueries.GetUserByEmail(r.Context(), loginReq.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithJSON(w,
				http.StatusUnauthorized,
				map[string]string{"error": "Incorrect email or password"},
			)
			return
		}
		respondWithJSON(w,
			http.StatusInternalServerError,
			map[string]string{"error": "Failed to retrieve user"},
		)
		return
	}

	match, err := auth.CheckPasswordHash(loginReq.Password, user.HashedPassword)
	if err != nil {
		respondWithJSON(w,
			http.StatusInternalServerError,
			map[string]string{"error": "Failed to check password"},
		)
		return
	}
	if !match {
		respondWithJSON(w,
			http.StatusUnauthorized,
			map[string]string{"error": "Incorrect email or password"},
		)
		return
	}
	expiresIn := time.Duration(1 * time.Hour) // Default expiration time of 1 hour
	if loginReq.ExpiresIn != nil {
		expiresIn = time.Duration(*loginReq.ExpiresIn) * time.Second
	}
	token, err := auth.MakeJWT(user.ID, c.tokenSecret, expiresIn)
	if err != nil {
		respondWithJSON(w,
			http.StatusInternalServerError,
			map[string]string{"error": "Failed to generate JWT"},
		)
		return
	}

	userResponse := UserResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	}

	respondWithJSON(w, http.StatusOK, userResponse)
}
