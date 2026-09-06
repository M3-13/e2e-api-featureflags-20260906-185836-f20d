package main

import (
	"log"
	"net/http"
	"os"
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
	return logMiddleware(mux)
}

func main() {
	srv := &server{store: NewStore()}
	handler := newHandler(srv)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
