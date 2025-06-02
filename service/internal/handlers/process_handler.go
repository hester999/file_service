package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ProcessHandler struct {
	fileService FileService
}

func NewProcessHandler(fileService FileService) *ProcessHandler {
	return &ProcessHandler{
		fileService: fileService,
	}
}

func (p *ProcessHandler) GetFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	files, err := p.fileService.GetFiles(r.Context())
	if err != nil {
		response := map[string]interface{}{
			"status": "failed",
			"error":  err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status": "success",
		"files":  files,
	}
	json.NewEncoder(w).Encode(response)
	w.WriteHeader(http.StatusOK)
}

func (p *ProcessHandler) GetFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL: expected /file/<id>", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, "Invalid ID: must be an integer", http.StatusBadRequest)
		return
	}

	content, err := p.fileService.GetFile(r.Context(), id)
	if err != nil {
		response := map[string]interface{}{
			"status": "failed",
			"error":  err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	respStruct := struct {
		ID        int            `yaml:"id"`
		Name      string         `yaml:"name"`
		Timestamp time.Time      `yaml:"timestamp"`
		Values    []float64      `yaml:"values"`
		Metadata  map[string]int `yaml:"metadata"`
	}{
		ID:        content.ID,
		Name:      content.Name,
		Timestamp: content.Timestamp,
		Values:    content.Values,
		Metadata:  content.Metadata,
	}
	json.NewEncoder(w).Encode(&respStruct)
	w.WriteHeader(http.StatusOK)
}
