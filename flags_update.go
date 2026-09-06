package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const maxBodyBytes = 1 << 20 // 1 MiB

type updateFlagRequest struct {
	Enabled        *bool   `json:"enabled"`
	Description    *string `json:"description"`
	RolloutPercent *int    `json:"rollout_percent"`
}

func (s *server) handleUpdateFlag(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	existing, ok := s.store.Get(key)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	var req updateFlagRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			// empty body: no fields to update
			writeJSON(w, http.StatusOK, map[string]Flag{"flag": existing})
			return
		}
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	// reject trailing data after the first JSON value
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.RolloutPercent != nil {
		if *req.RolloutPercent < 0 || *req.RolloutPercent > 100 {
			writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
			return
		}
		existing.RolloutPercent = *req.RolloutPercent
	}

	s.store.Update(key, existing)
	writeJSON(w, http.StatusOK, map[string]Flag{"flag": existing})
}

func (s *server) handleDeleteFlag(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if !s.store.Delete(key) {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
