package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPing(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	response := httptest.NewRecorder()

	newMux().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	if response.Body.String() != "pong\n" {
		t.Fatalf("expected pong response, got %q", response.Body.String())
	}
}

func TestPingRejectsNonGet(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/ping", nil)
	response := httptest.NewRecorder()

	newMux().ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, response.Code)
	}
}
