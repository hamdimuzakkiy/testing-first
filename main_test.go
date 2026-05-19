package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeOCR struct {
	text string
}

func (f fakeOCR) ExtractText(string) (string, error) {
	return f.text, nil
}

func TestPing(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	response := httptest.NewRecorder()

	newMux(fakeOCR{}).ServeHTTP(response, request)

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

	newMux(fakeOCR{}).ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, response.Code)
	}
}

func TestOCR(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("image", "receipt.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("fake image bytes")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/ocr", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()

	newMux(fakeOCR{text: "hello from local OCR"}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload map[string]string
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload["text"] != "hello from local OCR" {
		t.Fatalf("expected OCR text, got %q", payload["text"])
	}
}

func TestOCRRequiresImage(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/ocr", strings.NewReader(""))
	request.Header.Set("Content-Type", "multipart/form-data; boundary=missing")
	response := httptest.NewRecorder()

	newMux(fakeOCR{}).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}
