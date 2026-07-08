package main

import (
	"database/sql"
	"simplehttp/internal/database"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	DB             *sql.DB
	platform       string
	tokenSecret    string
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token,omitempty"`
}

type ChirpResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}
