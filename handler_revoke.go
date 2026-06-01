package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/slizhunter/chirpy/internal/auth"
)

// handleRevoke handles the revoke token endpoint.
func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid bearer token", err)
		return
	}
	// Get the refresh token from the database
	dbRefreshToken, err := cfg.dbQueries.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Trace: no refresh token row for /api/revoke (token_prefix=%q)", tokenPrefix(refreshToken))
		}
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	// Revoke the refresh token
	err = cfg.dbQueries.RevokeRefreshToken(r.Context(), dbRefreshToken.Token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to revoke refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
