package main

import (
	"database/sql"
	"encoding/json"
	"log"
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

// handlerPostChirp handles the creation of a new chirp. It expects a JSON body with the chirp content.
// The authenticated user ID is extracted from the JWT in the Authorization header. It returns the created chirp in the response.
func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Body string `json:"body"`
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
	var reqBody requestBody
	// Decode the JSON body into the requestBody struct
	err = json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}
	// Check chirp length
	if len(reqBody.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
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

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
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
		if err == sql.ErrNoRows {
			log.Printf("Trace: chirp not found for GET /api/chirps/{chirpID} (chirp_id=%q)", chirpID)
		}
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

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
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
	// Extract the chirp ID from the URL path
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}
	// Retrieve the chirp from the database to check ownership before deletion
	chirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Trace: chirp not found for DELETE /api/chirps/{chirpID} (chirp_id=%q)", chirpID)
		}
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}
	// Check if the authenticated user is the owner of the chirp
	if chirp.UserID != authenticatedUserID {
		respondWithError(w, http.StatusForbidden, "You are not allowed to delete this chirp", nil)
		return
	}
	// Delete the chirp from the database
	err = cfg.dbQueries.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete chirp", err)
		return
	}
	respondWithJSON(w, http.StatusNoContent, nil)
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
