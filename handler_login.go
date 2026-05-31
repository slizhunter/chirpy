package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/slizhunter/chirpy/internal/auth"
	"github.com/slizhunter/chirpy/internal/database"
)

// expirationDays defines the number of days after which the refresh token expires.
const expirationDays = 60 * 24 * time.Hour // 60 days

// accessTokenExpiration defines the duration after which the access token expires.
const accessTokenExpiration = time.Hour

// handlerLogin handles user login requests. It expects a JSON body with email and password, validates the credentials,
// and responds with a JSON object containing the user information, JWT token, and refresh token.
func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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

	// Gets the user from the database by email
	dbUser, err := cfg.dbQueries.GetUserByEmail(r.Context(), reqBody.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get user", err)
		return
	}

	// Check the password hash
	match, err := auth.CheckPasswordHash(reqBody.Password, dbUser.HashedPassword)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password", nil)
		return
	}

	// Generate JWT token
	expiresIn := accessTokenExpiration
	token, err := auth.MakeJWT(dbUser.ID, cfg.secret, expiresIn)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate JWT", err)
		return
	}
	// Generate refresh token
	refreshTokenNum := auth.MakeRefreshToken()
	// Store the refresh token in the database
	refreshToken, err := cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshTokenNum,
		UserID:    dbUser.ID,
		ExpiresAt: time.Now().Add(expirationDays),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create refresh token", err)
		return
	}

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
		Token:     token,
	}
	respondWithJSON(w, http.StatusOK, response{
		User:         user,
		Token:        token,
		RefreshToken: refreshToken.Token,
	})
}
