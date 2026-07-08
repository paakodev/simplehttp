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

func (c *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithJSON(w,
			http.StatusUnauthorized,
			map[string]string{"error": "Missing or invalid token"},
		)
		return
	}

	userID, err := auth.ValidateJWT(token, c.tokenSecret)
	if err != nil {
		respondWithJSON(w,
			http.StatusUnauthorized,
			map[string]string{"error": "Invalid token"},
		)
		return
	}

	type UpdateUserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	updateRequest := UpdateUserRequest{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&updateRequest)
	if err != nil {
		respondWithJSON(w,
			http.StatusBadRequest,
			map[string]string{"error": "Invalid request body"},
		)
		return
	}

	user, err := c.dbQueries.GetUserByID(r.Context(), userID)
	if err != nil {
		respondWithJSON(w,
			http.StatusInternalServerError,
			map[string]string{"error": "Failed to retrieve user"},
		)
		return
	}

	if updateRequest.Email != "" {
		user.Email = updateRequest.Email
	}
	if updateRequest.Password != "" {
		hashedPassword, err := auth.HashPassword(updateRequest.Password)
		if err != nil {
			respondWithJSON(w,
				http.StatusInternalServerError,
				map[string]string{"error": "Failed to hash password"},
			)
			return
		}
		user.HashedPassword = hashedPassword
	}

	updatedUser, err := c.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             user.ID,
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
	})
	if err != nil {
		respondWithJSON(w,
			http.StatusInternalServerError,
			map[string]string{"error": "Failed to update user"},
		)
		return
	}

	userResponse := UserResponse{
		ID:        updatedUser.ID,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
		Email:     updatedUser.Email,
	}

	respondWithJSON(w, http.StatusOK, userResponse)
}

func (c *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	expiresIn := time.Duration(1 * time.Hour)
	token, err := auth.MakeJWT(user.ID, c.tokenSecret, expiresIn)
	if err != nil {
		respondWithJSON(w,
			http.StatusInternalServerError,
			map[string]string{"error": "Failed to generate JWT"},
		)
		return
	}
	refreshToken := auth.MakeRefreshToken()
	_, err = c.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		ID:        uuid.New(),
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour), // Refresh token valid for 60 days
		RevokedAt: sql.NullTime{Valid: false},          // Not revoked
	})
	if err != nil {
		respondWithJSON(w,
			http.StatusInternalServerError,
			map[string]string{"error": "Failed to create refresh token"},
		)
		return
	}

	userResponse := UserResponse{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
	}

	respondWithJSON(w, http.StatusOK, userResponse)
}

func (c *apiConfig) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != http.NoBody {
		respondWithJSON(w,
			http.StatusNotAcceptable,
			map[string]string{"error": "Cannot accept a request body for this endpoint"},
		)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithJSON(w,
			http.StatusUnauthorized,
			map[string]string{"error": "Missing or invalid token"},
		)
		return
	}

	userID, err := c.dbQueries.GetUserIDByRefreshToken(r.Context(), token)
	if err != nil {
		respondWithJSON(w,
			http.StatusNotAcceptable,
			map[string]string{"error": "Invalid refresh token"},
		)
		return
	}

	if userID.RevokedAt.Valid && userID.RevokedAt.Time.Before(time.Now()) {
		respondWithJSON(w,
			http.StatusUnauthorized,
			map[string]string{"error": "Refresh token is expired or revoked"},
		)
		return
	}

	newToken, err := auth.MakeJWT(userID.UserID, c.tokenSecret, time.Hour)
	if err != nil {
		respondWithJSON(w,
			http.StatusInternalServerError,
			map[string]string{"error": "Failed to generate new JWT"},
		)
		return
	}

	respondWithJSON(w,
		http.StatusOK,
		map[string]string{"token": newToken},
	)
}

func (c *apiConfig) revokeRefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != http.NoBody {
		respondWithJSON(w,
			http.StatusNotAcceptable,
			map[string]string{"error": "Cannot accept a request body for this endpoint"},
		)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithJSON(w,
			http.StatusUnauthorized,
			map[string]string{"error": "Missing or invalid token"},
		)
		return
	}

	err = c.dbQueries.RevokeRefreshTokenByToken(r.Context(), token)
	if err != nil {
		respondWithJSON(w,
			http.StatusInternalServerError,
			map[string]string{"error": "Failed to revoke refresh token"},
		)
		return
	}

	respondWithNoContent(w)
}
