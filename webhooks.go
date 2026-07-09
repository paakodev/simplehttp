package main

import (
	"encoding/json"
	"net/http"
	"simplehttp/internal/auth"

	"github.com/google/uuid"
)

func (c *apiConfig) polkaWebhookHandler(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || apiKey != c.polkaKey {
		respondWithJSON(w,
			http.StatusUnauthorized,
			map[string]string{"error": "Missing or invalid API key"},
		)
		return
	}

	type UpgradeRequest struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	updateRequest := UpgradeRequest{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&updateRequest)
	if err != nil {
		respondWithJSON(w,
			http.StatusBadRequest,
			map[string]string{"error": "Invalid request body"},
		)
		return
	}

	if updateRequest.Event != "user.upgraded" {
		respondWithNoContent(w)
		return
	}

	userID, err := uuid.Parse(updateRequest.Data.UserID)
	if err != nil {
		respondWithJSON(w,
			http.StatusBadRequest,
			map[string]string{"error": "Invalid user ID"},
		)
		return
	}

	_, err = c.dbQueries.UpgradeUserToChirpyRed(r.Context(), userID)
	if err != nil {
		respondWithJSON(w,
			http.StatusInternalServerError,
			map[string]string{"error": "Failed to update user"},
		)
		return
	}

	respondWithNoContent(w)
}
