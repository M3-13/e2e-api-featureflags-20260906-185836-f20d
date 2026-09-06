package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogMiddlewareLogsMethodPathStatus(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	handler := logMiddleware(next)

	req := httptest.NewRequest(http.MethodPost, "/flags/myflag?user=alice&secret=42", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	out := buf.String()
	if !strings.Contains(out, http.MethodPost) {
		t.Fatalf("expected log to contain method %q, got %q", http.MethodPost, out)
	}
	if !strings.Contains(out, "/flags/myflag") {
		t.Fatalf("expected log to contain path /flags/myflag, got %q", out)
	}
	if !strings.Contains(out, "201") {
		t.Fatalf("expected log to contain status 201, got %q", out)
	}
	if strings.Contains(out, "?") {
		t.Fatalf("expected log to omit query string, got %q", out)
	}
	if strings.Contains(out, "user=alice") || strings.Contains(out, "alice") {
		t.Fatalf("expected log to omit user parameter, got %q", out)
	}
	if strings.Contains(out, "secret=42") {
		t.Fatalf("expected log to omit query values, got %q", out)
	}
}

func TestLogMiddlewareMasksFlagKeyInPattern(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /flags/{key}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := logMiddleware(mux)

	req := httptest.NewRequest(http.MethodGet, "/flags/alice-experimental?user=bob&token=xyz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	out := buf.String()
	if !strings.Contains(out, "/flags/{key}") {
		t.Fatalf("expected log to contain pattern /flags/{key}, got %q", out)
	}
	if strings.Contains(out, "alice-experimental") {
		t.Fatalf("expected log to omit concrete flag key, got %q", out)
	}
	if strings.Contains(out, "?") || strings.Contains(out, "token=xyz") || strings.Contains(out, "user=bob") {
		t.Fatalf("expected log to omit query string, got %q", out)
	}
	if strings.Contains(out, "bob") {
		t.Fatalf("expected log to omit user value, got %q", out)
	}
}

func TestLogMiddlewareDefaultsStatusTo200(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler := logMiddleware(next)

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	out := buf.String()
	if !strings.Contains(out, "200") {
		t.Fatalf("expected log to contain default status 200, got %q", out)
	}
}
