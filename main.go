package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

type server struct {
	store *Store
}

func newHandler(srv *server) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealth)
	mux.HandleFunc("POST /flags", srv.handleCreateFlag)
	mux.HandleFunc("GET /flags", srv.handleListFlags)
	mux.HandleFunc("GET /flags/{key}", srv.handleGetFlag)
	mux.HandleFunc("PUT /flags/{key}", srv.handleUpdateFlag)
	mux.HandleFunc("DELETE /flags/{key}", srv.handleDeleteFlag)
	mux.HandleFunc("GET /flags/{key}/evaluate", srv.handleEvaluate)
	apiKey := os.Getenv("FLAG_API_KEY")
	return logMiddleware(authMiddleware(apiKey, mux))
}

func main() {
	srv := &server{store: NewStore()}
	handler := newHandler(srv)

	host := os.Getenv("HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := host + ":" + port
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	log.Printf("listening on %s", addr)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
