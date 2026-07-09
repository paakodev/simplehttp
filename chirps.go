package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"simplehttp/internal/auth"
	"simplehttp/internal/database"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

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

func (c *apiConfig) chirpPost(w http.ResponseWriter, r *http.Request) {
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

	type Chirp struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
		Token  string    `json:"token"`
	}
	decoder := json.NewDecoder(r.Body)
	chirpData := Chirp{}
	err = decoder.Decode(&chirpData)
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
		UserID: userID,
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

func (c *apiConfig) deleteChirpByChirpID(w http.ResponseWriter, r *http.Request) {
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

	chirpID := uuid.MustParse(r.PathValue("chirpID"))

	chirpUserID, err := c.dbQueries.GetUserIdFromChirpId(r.Context(), chirpID)
	if err != nil {
		respondWithJSON(w,
			http.StatusNotFound,
			map[string]string{"error": "Chirp not found"},
		)
		return
	}

	if chirpUserID != userID {
		respondWithJSON(w,
			http.StatusForbidden,
			map[string]string{"error": "You are not authorized to delete this chirp"},
		)
		return
	}

	err = c.dbQueries.DeleteChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithJSON(w,
			http.StatusInternalServerError,
			map[string]string{"error": "Failed to delete chirp"},
		)
		return
	}

	respondWithNoContent(w)
}

func (c *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	opts, err := parseChirpQueryOptions(r)
	if err != nil {
		respondWithJSON(w,
			http.StatusBadRequest,
			map[string]string{"error": err.Error()},
		)
		return
	}

	var chirps []database.Chirp
	switch {
	case opts.AuthorID != nil && opts.SortOrder == "asc":
		chirps, err = c.dbQueries.GetChirpsByUserIDAsc(r.Context(), database.GetChirpsByUserIDAscParams{
			UserID: *opts.AuthorID,
			Limit:  opts.Limit,
			Offset: opts.Offset,
		})
	case opts.AuthorID != nil && opts.SortOrder == "desc":
		chirps, err = c.dbQueries.GetChirpsByUserIDDesc(r.Context(), database.GetChirpsByUserIDDescParams{
			UserID: *opts.AuthorID,
			Limit:  opts.Limit,
			Offset: opts.Offset,
		})
	case opts.AuthorID == nil && opts.SortOrder == "asc":
		chirps, err = c.dbQueries.GetAllChirpsAsc(r.Context(), database.GetAllChirpsAscParams{
			Limit:  opts.Limit,
			Offset: opts.Offset,
		})
	case opts.AuthorID == nil && opts.SortOrder == "desc":
		chirps, err = c.dbQueries.GetAllChirpsDesc(r.Context(), database.GetAllChirpsDescParams{
			Limit:  opts.Limit,
			Offset: opts.Offset,
		})
	default:
		log.Printf("unexpected sort order %q; falling back to ascending", opts.SortOrder)

		if opts.AuthorID != nil {
			chirps, err = c.dbQueries.GetChirpsByUserIDAsc(r.Context(), database.GetChirpsByUserIDAscParams{
				UserID: *opts.AuthorID,
				Limit:  opts.Limit,
				Offset: opts.Offset,
			})
		} else {
			chirps, err = c.dbQueries.GetAllChirpsAsc(r.Context(), database.GetAllChirpsAscParams{
				Limit:  opts.Limit,
				Offset: opts.Offset,
			})
		}
	}

	if err != nil {
		respondWithJSON(w,
			http.StatusInternalServerError,
			map[string]string{"error": "Failed to retrieve chirps"},
		)
		return
	}

	chirpResponses := make([]ChirpResponse, len(chirps))
	for i, chirp := range chirps {
		chirpResponses[i] = ChirpResponse{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
	}

	respondWithJSON(w, http.StatusOK, chirpResponses)
}

func parseChirpQueryOptions(r *http.Request) (ChirpQueryOptions, error) {
	const maxLimit = 1000

	options := ChirpQueryOptions{
		Limit:     100,   // default limit
		Offset:    0,     // default offset
		SortOrder: "asc", // default sort order
		AuthorID:  nil,   // default author ID
	}

	query := r.URL.Query()

	authorID := query.Get("author_id")
	if authorID != "" {
		uid, err := uuid.Parse(authorID)
		if err != nil {
			return ChirpQueryOptions{}, fmt.Errorf("invalid author_id")
		}
		options.AuthorID = &uid
	}

	sort := query.Get("sort")
	switch sort {
	case "desc":
		options.SortOrder = "desc"
	case "asc", "":
		options.SortOrder = "asc"
	default:
		return ChirpQueryOptions{}, fmt.Errorf("invalid sort option")
	}

	n, err := strconv.ParseInt(query.Get("limit"), 10, 32)
	if err == nil {
		limit := int32(n)
		if limit > maxLimit {
			limit = maxLimit
		}
		options.Limit = limit
	}

	n, err = strconv.ParseInt(query.Get("offset"), 10, 32)
	if err == nil {
		options.Offset = int32(n)
	}

	return options, nil
}

func (c *apiConfig) getChirpByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	uid, err := uuid.Parse(id)
	if err != nil {
		respondWithJSON(w,
			http.StatusBadRequest,
			map[string]string{"error": "Invalid chirp ID"},
		)
		return
	}

	chirp, err := c.dbQueries.GetChirpByID(r.Context(), uid)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithJSON(w,
				http.StatusNotFound,
				map[string]string{"error": "Chirp not found"},
			)
			return
		}
		respondWithJSON(w,
			http.StatusInternalServerError,
			map[string]string{"error": "Failed to retrieve chirp"},
		)
		return
	}

	chirpResponse := ChirpResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, http.StatusOK, chirpResponse)
}
