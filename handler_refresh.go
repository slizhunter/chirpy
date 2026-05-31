package main

import (
	"net/http"
	"time"

	"github.com/slizhunter/chirpy/internal/auth"
)

// handleRefresh handles the refresh token endpoint.
func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	//
	type response struct {
		Token string `json:"token"`
	}
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid bearer token", err)
		return
	}
	// Get the refresh token from the database
	dbRefreshToken, err := cfg.dbQueries.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	// Check if the refresh token has expired
	if time.Now().After(dbRefreshToken.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token has expired", nil)
		return
	}

	if dbRefreshToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Refresh token has been revoked", nil)
		return
	}

	// Generate a new JWT token
	token, err := auth.MakeJWT(dbRefreshToken.UserID, cfg.secret, accessTokenExpiration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate JWT", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{Token: token})
}
