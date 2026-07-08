package main

import (
	"fmt"
	"net/http"
)

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

// XXX: This also resets the users and all the chirps!
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
