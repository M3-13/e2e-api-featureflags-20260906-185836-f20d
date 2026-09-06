package main

import "net/http"

func (s *server) handleEvaluate(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

func rolloutDecision(key, user string, percent int) bool {
	return false
}
