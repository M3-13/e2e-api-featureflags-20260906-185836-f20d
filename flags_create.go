package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

const maxFlagBodyBytes = 1 << 20

type createFlagRequest struct {
	Key            string `json:"key"`
	Description    string `json:"description"`
	Enabled        *bool  `json:"enabled"`
	RolloutPercent *int   `json:"rollout_percent"`
}

func (s *server) handleCreateFlag(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxFlagBodyBytes)

	var req createFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if req.Key == "" {
		writeError(w, http.StatusBadRequest, "key is required")
		return
	}
	if req.Enabled == nil {
		writeError(w, http.StatusBadRequest, "enabled is required")
		return
	}

	rollout := 0
	if req.RolloutPercent != nil {
		rollout = *req.RolloutPercent
		if rollout < 0 || rollout > 100 {
			writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
			return
		}
	}

	flag := Flag{
		Key:            req.Key,
		Description:    req.Description,
		Enabled:        *req.Enabled,
		RolloutPercent: rollout,
	}

	if err := s.store.Create(flag); err != nil {
		writeError(w, http.StatusConflict, "flag already exists")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]Flag{"flag": flag})
}

func (s *server) handleListFlags(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.List())
}

func (s *server) handleGetFlag(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	flag, ok := s.store.Get(key)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]Flag{"flag": flag})
}
