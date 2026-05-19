package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func main() {
	mux := newMux(commandOCR{})

	addr := ":" + getEnv("PORT", "8080")
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

type OCRRunner interface {
	ExtractText(imagePath string) (string, error)
}

type commandOCR struct{}

func (commandOCR) ExtractText(imagePath string) (string, error) {
	command := getEnv("OCR_COMMAND", "tesseract")
	args := []string{imagePath, "stdout"}

	if configuredArgs := strings.Fields(os.Getenv("OCR_ARGS")); len(configuredArgs) > 0 {
		args = append(args, configuredArgs...)
	}

	output, err := exec.Command(command, args...).CombinedOutput()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("OCR command %q was not found; install tesseract or set OCR_COMMAND", command)
		}
		return "", fmt.Errorf("OCR command failed: %s", strings.TrimSpace(string(output)))
	}

	return strings.TrimSpace(string(output)), nil
}

func newMux(ocr OCRRunner) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", pingHandler)
	mux.HandleFunc("/ocr", ocrHandler(ocr))
	return mux
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("pong\n"))
}

func ocrHandler(ocr OCRRunner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "expected multipart form with an image file", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("image")
		if err != nil {
			http.Error(w, "missing image file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		tempFile, err := os.CreateTemp("", "ocr-*"+safeExt(header.Filename))
		if err != nil {
			http.Error(w, "failed to create temporary file", http.StatusInternalServerError)
			return
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		if _, err := io.Copy(tempFile, file); err != nil {
			http.Error(w, "failed to read uploaded image", http.StatusBadRequest)
			return
		}

		if err := tempFile.Close(); err != nil {
			http.Error(w, "failed to save uploaded image", http.StatusInternalServerError)
			return
		}

		text, err := ocr.ExtractText(tempFile.Name())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"text": text})
	}
}

func safeExt(filename string) string {
	index := strings.LastIndex(filename, ".")
	if index == -1 {
		return ""
	}

	ext := filename[index:]
	if len(ext) > 10 || strings.ContainsAny(ext, `/\`) {
		return ""
	}

	return ext
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
