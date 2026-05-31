package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/slizhunter/chirpy/internal/auth"
	"github.com/slizhunter/chirpy/internal/database"
)

var badWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

// handlerPostChirp handles the creation of a new chirp. It expects a JSON body with the chirp content,
// the user ID, and a JWT for authentication. It returns the created chirp in the response.
func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
		JWT    string    `json:"jwt"`
	}
	type response struct {
		Chirp
	}
	var reqBody requestBody
	// Decode the JSON body into the requestBody struct
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}
	// Check chirp length
	if len(reqBody.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	// Extract and validate the Bearer token from the Authorization header
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or missing Bearer token", err)
		return
	}
	// Validate the JWT and extract the authenticated user ID
	authenticatedUserID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or expired JWT", err)
		return
	}
	// Filter out any profane words from the chirp body
	cleanBody := profanityFilter(reqBody.Body)
	// Create the chirp in the database
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanBody,
		UserID: authenticatedUserID, // Use the authenticated user ID from the JWT
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		Chirp: Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		},
	})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Chirp
	}
	dbChirps, err := cfg.dbQueries.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get chirps", err)
		return
	}
	chirps := []Chirp{}
	for _, c := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserID:    c.UserID,
		})
	}
	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDStr := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}
	chirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Failed to get chirp", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func profanityFilter(text string) string {
	words := strings.Split(text, " ")
	for i, word := range words {
		if _, exists := badWords[strings.ToLower(word)]; exists {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}
