package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const testAPIKey = "test-api-key"

func TestMain(m *testing.M) {
	os.Setenv("FLAG_API_KEY", testAPIKey)
	os.Exit(m.Run())
}

func withAuthHeader(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Authorization", "Bearer "+testAPIKey)
		h.ServeHTTP(w, r)
	})
}

func newAuthHandler() http.Handler {
	return newHandler(&server{store: NewStore()})
}

func TestAuthMissingHeaderReturns401(t *testing.T) {
	h := newAuthHandler()
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthWrongKeyReturns401(t *testing.T) {
	h := newAuthHandler()
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	req.Header.Set("Authorization", "Bearer wrong-key")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestHealthzWithoutKeyReturns200(t *testing.T) {
	h := newAuthHandler()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestFlagsWithEmptyAPIKeyReturns503(t *testing.T) {
	t.Setenv("FLAG_API_KEY", "")
	h := newAuthHandler()
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestAuthValidKeyBothHeaders(t *testing.T) {
	h := newAuthHandler()

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	req.Header.Set("Authorization", "Bearer "+testAPIKey)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Authorization header: expected 200, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/flags", nil)
	req2.Header.Set("X-API-Key", testAPIKey)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("X-API-Key header: expected 200, got %d", rec2.Code)
	}
}
