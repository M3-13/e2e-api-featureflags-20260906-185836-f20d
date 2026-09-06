package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func newUpdateTestServer() (*server, http.Handler) {
	srv := &server{store: NewStore()}
	return srv, newHandler(srv)
}

func TestUpdateFlagChangesFields(t *testing.T) {
	srv, handler := newUpdateTestServer()
	_ = srv.store.Create(Flag{Key: "myflag", Description: "old", Enabled: false, RolloutPercent: 10})

	body := `{"enabled":true,"description":"new desc","rollout_percent":75}`
	req := httptest.NewRequest(http.MethodPut, "/flags/myflag", strings.NewReader(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Flag Flag `json:"flag"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unexpected response body: %v", err)
	}
	if resp.Flag.Key != "myflag" || resp.Flag.Description != "new desc" || !resp.Flag.Enabled || resp.Flag.RolloutPercent != 75 {
		t.Fatalf("unexpected flag in response: %+v", resp.Flag)
	}

	// persisted in the store (subsequent GET would reflect this)
	got, ok := srv.store.Get("myflag")
	if !ok {
		t.Fatal("expected flag to still exist")
	}
	if got.Description != "new desc" || !got.Enabled || got.RolloutPercent != 75 {
		t.Fatalf("flag not persisted: %+v", got)
	}
}

func TestUpdateFlagPartialUpdate(t *testing.T) {
	srv, handler := newUpdateTestServer()
	_ = srv.store.Create(Flag{Key: "myflag", Description: "old", Enabled: false, RolloutPercent: 10})

	req := httptest.NewRequest(http.MethodPut, "/flags/myflag", strings.NewReader(`{"enabled":true}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	got, _ := srv.store.Get("myflag")
	if !got.Enabled {
		t.Fatal("expected enabled to be updated")
	}
	if got.Description != "old" || got.RolloutPercent != 10 {
		t.Fatalf("unexpected fields changed: %+v", got)
	}
}

func TestUpdateFlagUnknownKeyReturns404(t *testing.T) {
	_, handler := newUpdateTestServer()
	req := httptest.NewRequest(http.MethodPut, "/flags/nope", strings.NewReader(`{"enabled":true}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestUpdateFlagRolloutPercentOutOfRangeReturns400(t *testing.T) {
	for _, pct := range []int{-1, 101} {
		srv, handler := newUpdateTestServer()
		_ = srv.store.Create(Flag{Key: "myflag"})
		req := httptest.NewRequest(http.MethodPut, "/flags/myflag", strings.NewReader(fmt.Sprintf(`{"rollout_percent":%d}`, pct)))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("rollout_percent %d: expected status 400, got %d", pct, rec.Code)
		}
	}
}

func TestUpdateFlagInvalidJSONReturns400(t *testing.T) {
	srv, handler := newUpdateTestServer()
	_ = srv.store.Create(Flag{Key: "myflag"})
	req := httptest.NewRequest(http.MethodPut, "/flags/myflag", strings.NewReader(`{invalid`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestUpdateFlagBodyTooLargeReturns413(t *testing.T) {
	srv, handler := newUpdateTestServer()
	_ = srv.store.Create(Flag{Key: "myflag"})
	big := strings.Repeat("a", maxBodyBytes+1024)
	req := httptest.NewRequest(http.MethodPut, "/flags/myflag", strings.NewReader(`{"description":"`+big+`"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status 413, got %d", rec.Code)
	}
}

func TestDeleteFlagRemovesFlag(t *testing.T) {
	srv, handler := newUpdateTestServer()
	_ = srv.store.Create(Flag{Key: "myflag", Enabled: true})

	req := httptest.NewRequest(http.MethodDelete, "/flags/myflag", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
	if _, ok := srv.store.Get("myflag"); ok {
		t.Fatal("expected flag to be gone after delete")
	}
}

func TestDeleteFlagUnknownKeyReturns404(t *testing.T) {
	_, handler := newUpdateTestServer()
	req := httptest.NewRequest(http.MethodDelete, "/flags/nope", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestUpdateFlagDescriptionTooLongReturns400(t *testing.T) {
	srv, handler := newUpdateTestServer()
	_ = srv.store.Create(Flag{Key: "myflag"})
	desc := strings.Repeat("d", 4097)
	req := httptest.NewRequest(http.MethodPut, "/flags/myflag", strings.NewReader(`{"description":"`+desc+`"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestConcurrentDeleteAndUpdateIfExists(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "myflag", Enabled: true})

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		s.Delete("myflag")
	}()
	go func() {
		defer wg.Done()
		s.UpdateIfExists("myflag", Flag{Key: "myflag", Enabled: false, Description: "changed"})
	}()
	wg.Wait()

	if _, ok := s.Get("myflag"); ok {
		t.Fatal("expected flag to be deleted, not resurrected by UpdateIfExists")
	}
}
