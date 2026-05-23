package main

import (
	"net/http"
)

// handlerReset resets the fileserverHits counter to 0 and responds with an HTTP 200 OK status.
func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0) // Reset the fileserverHits counter to 0
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Fileserver hits counter reset to 0"))
}
