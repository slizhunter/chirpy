package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

var badWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

// handlerValidate validates the incoming request body and responds with an HTTP 200 OK status if the validation is successful,
// or an HTTP 400 Bad Request status if the validation fails.
func handlerValidate(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
	}
	var reqBody requestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}
	if len(reqBody.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	cleanBody := profanityFilter(reqBody.Body)
	respondWithJSON(w, http.StatusOK, returnVals{CleanedBody: cleanBody})
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
