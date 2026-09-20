package tests

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	platformhttp "server/internal/platform/http"
)

func TestWriteJSON(t *testing.T) {
	recorder := httptest.NewRecorder()

	platformhttp.WriteJSON(recorder, http.StatusCreated, map[string]string{"status": "created"})

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if recorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("content type = %q, want application/json", recorder.Header().Get("Content-Type"))
	}

	body := map[string]string{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["status"] != "created" {
		t.Fatalf("unexpected body: %q err=%v", recorder.Body.String(), err)
	}
}

func TestWriteError(t *testing.T) {
	recorder := httptest.NewRecorder()

	platformhttp.WriteError(recorder, http.StatusConflict, errors.New("duplicated"))

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
	}

	body := map[string]string{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["error"] != "duplicated" {
		t.Fatalf("unexpected body: %q err=%v", recorder.Body.String(), err)
	}
}

func TestDecodeJSON(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	t.Run("decodes known fields", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Ana"}`))
		var destination payload

		if err := platformhttp.DecodeJSON(recorder, request, &destination); err != nil {
			t.Fatalf("DecodeJSON() error = %v", err)
		}
		if destination.Name != "Ana" {
			t.Fatalf("unexpected payload: %#v", destination)
		}
	})

	t.Run("rejects unknown fields", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Ana","extra":1}`))
		var destination payload

		if err := platformhttp.DecodeJSON(recorder, request, &destination); err == nil {
			t.Fatal("DecodeJSON() expected an error for unknown fields")
		}
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":`))
		var destination payload

		if err := platformhttp.DecodeJSON(recorder, request, &destination); err == nil {
			t.Fatal("DecodeJSON() expected an error for invalid json")
		}
	})

	t.Run("rejects oversized body", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"`+strings.Repeat("a", 1<<20)+`"}`))
		var destination payload

		if err := platformhttp.DecodeJSON(recorder, request, &destination); err == nil {
			t.Fatal("DecodeJSON() expected an error for oversized body")
		}
	})
}
