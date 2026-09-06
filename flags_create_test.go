package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func jsonInt(v int) string {
	return strconv.Itoa(v)
}

func newTestHandler() http.Handler {
	return newHandler(&server{store: NewStore()})
}

func TestCreateFlagThenGet(t *testing.T) {
	h := newTestHandler()

	body := `{"key":"myflag","enabled":true,"description":"d","rollout_percent":25}`
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Flag Flag `json:"flag"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Flag.Key != "myflag" || !resp.Flag.Enabled || resp.Flag.RolloutPercent != 25 || resp.Flag.Description != "d" {
		t.Fatalf("unexpected flag: %+v", resp.Flag)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/flags/myflag", nil)
	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected GET 200, got %d", getRec.Code)
	}
	var getResp struct {
		Flag Flag `json:"flag"`
	}
	if err := json.Unmarshal(getRec.Body.Bytes(), &getResp); err != nil {
		t.Fatalf("unmarshal get response: %v", err)
	}
	if getResp.Flag.Key != "myflag" {
		t.Fatalf("unexpected get flag: %+v", getResp.Flag)
	}
}

func TestListFlagsContainsCreated(t *testing.T) {
	h := newTestHandler()

	postReq := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"a","enabled":true}`))
	postRec := httptest.NewRecorder()
	h.ServeHTTP(postRec, postReq)
	if postRec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", postRec.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/flags", nil)
	listRec := httptest.NewRecorder()
	h.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", listRec.Code)
	}

	var flags []Flag
	if err := json.Unmarshal(listRec.Body.Bytes(), &flags); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	if len(flags) != 1 || flags[0].Key != "a" {
		t.Fatalf("unexpected list: %+v", flags)
	}
}

func TestListFlagsEmpty(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := strings.TrimSpace(rec.Body.String())
	if body != "[]" {
		t.Fatalf("expected empty array, got %q", body)
	}
}

func TestGetFlagNotFound(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/flags/unknown", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestCreateFlagMissingKey(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"enabled":true}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateFlagWrongJSONType(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"a","enabled":"yes"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateFlagRolloutOutOfRange(t *testing.T) {
	for _, v := range []int{-1, 101} {
		h := newTestHandler()
		body := `{"key":"a","enabled":true,"rollout_percent":` + jsonInt(v) + `}`
		req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for rollout %d, got %d", v, rec.Code)
		}
	}
}

func TestCreateFlagBodyTooLarge(t *testing.T) {
	h := newTestHandler()
	filler := strings.Repeat("x", maxFlagBodyBytes+1)
	body := `{"key":"a","enabled":true,"description":"` + filler + `"}`

	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

func TestCreateFlagDefaultRollout(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"a","enabled":true}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var resp struct {
		Flag Flag `json:"flag"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Flag.RolloutPercent != 0 {
		t.Fatalf("expected default rollout 0, got %d", resp.Flag.RolloutPercent)
	}
}

func TestCreateFlagDuplicateKey(t *testing.T) {
	h := newTestHandler()
	post := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"dup","enabled":true}`))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}
	if post().Code != http.StatusCreated {
		t.Fatal("expected first create to succeed")
	}
	if post().Code != http.StatusConflict {
		t.Fatal("expected duplicate create to return 409")
	}
}
