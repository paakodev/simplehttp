package main

import (
	"net/http"
)

func registerRoutes(mux *http.ServeMux, apiCfg *apiConfig) {
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
		http.HandlerFunc(apiCfg.chirpPost),
		apiCfg.middlewareMetricsInc,
		middlewareLog,
	))
	mux.Handle("GET /api/chirps", chain(
		http.HandlerFunc(apiCfg.getChirps),
		apiCfg.middlewareMetricsInc,
		middlewareLog,
	))
	mux.Handle("GET /api/chirps/{id}", chain(
		http.HandlerFunc(apiCfg.getChirpByID),
		apiCfg.middlewareMetricsInc,
		middlewareLog,
	))
	mux.Handle("DELETE /api/chirps/{chirpID}", chain(
		http.HandlerFunc(apiCfg.deleteChirpByChirpID),
		apiCfg.middlewareMetricsInc,
		middlewareLog,
	))

	mux.Handle("POST /api/users", chain(
		http.HandlerFunc(apiCfg.createUser),
		apiCfg.middlewareMetricsInc,
		middlewareLog,
	))
	mux.Handle("PUT /api/users", chain(
		http.HandlerFunc(apiCfg.updateUser),
		apiCfg.middlewareMetricsInc,
		middlewareLog,
	))

	mux.Handle("POST /api/login", chain(
		http.HandlerFunc(apiCfg.loginHandler),
		apiCfg.middlewareMetricsInc,
		middlewareLog,
	))
	mux.Handle("POST /api/refresh", chain(
		http.HandlerFunc(apiCfg.refreshTokenHandler),
		apiCfg.middlewareMetricsInc,
		middlewareLog,
	))
	mux.Handle("POST /api/revoke", chain(
		http.HandlerFunc(apiCfg.revokeRefreshTokenHandler),
		apiCfg.middlewareMetricsInc,
		middlewareLog,
	))

	mux.Handle("POST /api/polka/webhooks", chain(
		http.HandlerFunc(apiCfg.polkaWebhookHandler),
		middlewareLog,
	))
}
