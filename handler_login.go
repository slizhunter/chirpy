package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/slizhunter/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Email            string         `json:"email"`
		Password         string         `json:"password"`
		ExpiresInSeconds *time.Duration `json:"expires_in_seconds"`
	}
	var reqBody params
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}
	if reqBody.Email == "" || reqBody.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required", nil)
		return
	}

	dbUser, err := cfg.dbQueries.GetUserByEmail(r.Context(), reqBody.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get user", err)
		return
	}

	match, err := auth.CheckPasswordHash(reqBody.Password, dbUser.HashedPassword)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password", nil)
		return
	}

	expiresIn := time.Hour // Default to 1 hour
	if reqBody.ExpiresInSeconds != nil &&
		*reqBody.ExpiresInSeconds > 0 &&
		*reqBody.ExpiresInSeconds < time.Hour {
		expiresIn = *reqBody.ExpiresInSeconds
	}
	token, err := auth.MakeJWT(dbUser.ID, cfg.secret, expiresIn)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate JWT", err)
		return
	}

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
		Token:     token,
	}
	respondWithJSON(w, http.StatusOK, user)
}
