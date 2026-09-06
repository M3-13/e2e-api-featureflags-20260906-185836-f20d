package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRolloutDecisionPercentZero(t *testing.T) {
	if rolloutDecision("flag", "alice", 0) {
		t.Fatal("expected false for percent 0")
	}
}

func TestRolloutDecisionPercentHundred(t *testing.T) {
	if !rolloutDecision("flag", "alice", 100) {
		t.Fatal("expected true for percent 100")
	}
}

func TestRolloutDecisionDeterministic(t *testing.T) {
	first := rolloutDecision("flag", "alice", 50)
	for i := 0; i < 100; i++ {
		if rolloutDecision("flag", "alice", 50) != first {
			t.Fatal("expected deterministic result for same key+user+percent")
		}
	}
}

func TestHandleEvaluateMissingUser(t *testing.T) {
	srv := &server{store: NewStore()}
	srv.store.Create(Flag{Key: "flag", RolloutPercent: 100})

	req := httptest.NewRequest(http.MethodGet, "/flags/flag/evaluate", nil)
	rec := httptest.NewRecorder()
	srv.handleEvaluate(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleEvaluateUnknownKey(t *testing.T) {
	srv := &server{store: NewStore()}

	req := httptest.NewRequest(http.MethodGet, "/flags/missing/evaluate?user=alice", nil)
	req.SetPathValue("key", "missing")
	rec := httptest.NewRecorder()
	srv.handleEvaluate(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
