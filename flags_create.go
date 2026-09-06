package main

import "net/http"

func (s *server) handleCreateFlag(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

func (s *server) handleListFlags(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

func (s *server) handleGetFlag(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}
