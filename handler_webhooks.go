package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/slizhunter/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerUpgradeUser(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Invalid API key", err)
		return
	}
	type params struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}
	var reqBody params
	err = json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	if reqBody.Event != "user.upgraded" {
		respondWithError(w, http.StatusNoContent, "Other, user not upgraded", nil)
		return
	}
	err = cfg.dbQueries.UpgradeUserToChirpyRed(r.Context(), reqBody.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found", err)
		return
	}
	respondWithJSON(w, http.StatusNoContent, "")
}
