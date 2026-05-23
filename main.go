package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	const filePathRoot = "."
	const port = "8080"

	// Initialize the API configuration and set up the HTTP server
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}
	mux := http.NewServeMux() // Set up the file server handler with the middleware to count hits
	fileserverHandler := http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot)))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileserverHandler))
	mux.HandleFunc("/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("/reset", apiCfg.handlerReset)
	mux.HandleFunc("/healthz", handlerReadiness)

	// Start the HTTP server
	newServer := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	// Log the server startup message
	log.Printf("Serving files from %s on port %s", filePathRoot, port)
	// Start the server and log any errors that occur
	log.Fatal(newServer.ListenAndServe())
}

// middlewareMetricsInc is a middleware function that increments the fileserverHits counter each time a request is made to the file server.
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	// Return a new handler function that wraps the original handler and increments the fileserverHits counter
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1) // Increment the fileserverHits counter
		next.ServeHTTP(w, r)      // Call the next handler in the chain
	})
}

// writeMetrics writes the current value of the fileserverHits counter to the HTTP response.
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())))
}

// handlerReadiness responds with an HTTP 200 OK status and a plain text message indicating that the server is ready to handle requests.
func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
