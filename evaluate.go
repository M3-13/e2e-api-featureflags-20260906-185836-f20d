package main

import (
	"crypto/sha256"
	"encoding/binary"
	"net/http"
)

func (s *server) handleEvaluate(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" {
		writeError(w, http.StatusBadRequest, "user parameter is required")
		return
	}

	key := r.PathValue("key")
	flag, ok := s.store.Get(key)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"result": rolloutDecision(flag.Key, user, flag.RolloutPercent)})
}

func rolloutDecision(key, user string, percent int) bool {
	if percent <= 0 {
		return false
	}
	if percent >= 100 {
		return true
	}

	h := sha256.Sum256([]byte(key + user))
	n := binary.BigEndian.Uint64(h[:8])
	return n%100 < uint64(percent)
}
