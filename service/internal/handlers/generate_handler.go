package handlers

import (
	"encoding/json"
	"net/http"
	"service/internal/entity"
)

type GenerateImpl struct {
	fileService FileService
}

func NewGenerateImpl(fileService FileService) *GenerateImpl {
	return &GenerateImpl{
		fileService: fileService,
	}
}

func (g *GenerateImpl) GenerateHandler(w http.ResponseWriter, r *http.Request) {
	cfg := struct {
		Iterations int `json:"iterations"`
		MaxWorkers int `json:"max_workers"`
		MaxFiles   int `json:"max_files"`
	}{}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		response := map[string]interface{}{
			"status": "error",
			"error":  "Invalid JSON format",
			"code":   http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}


	if cfg.MaxFiles <=0 {
		cfg.MaxFiles = 10
	}

	if cfg.Iterations <= 0 {
		response := map[string]interface{}{
			"status": "error",
			"error":  "Iterations must be positive",
			"code":   http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if cfg.MaxWorkers <= 0 {
		response := map[string]interface{}{
			"status": "error",
			"error":  "Max workers must be positive",
			"code":   http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	workerConf := entity.WorkerConf{
		Iterations: cfg.Iterations,
		MaxWorkers: cfg.MaxWorkers,
		MaxFiles:   cfg.MaxFiles,
	}

	errors, err := g.fileService.Process(r.Context(), workerConf)
	if err != nil {
		response := map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
			"code":   http.StatusInternalServerError,
		}
		if len(errors) > 0 {
			errorMessages := make([]string, len(errors))
			for i, e := range errors {
				errorMessages[i] = e.Error()
			}
			response["worker_errors"] = errorMessages
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"status": "success",
		"code":   http.StatusOK,
		"data": map[string]interface{}{
			"iterations":  cfg.Iterations,
			"max_workers": cfg.MaxWorkers,
			"max_files":   cfg.MaxFiles,
		},
	}
	if len(errors) > 0 {
		response["status"] = "partial_success"
		errorMessages := make([]string, len(errors))
		for i, e := range errors {
			errorMessages[i] = e.Error()
		}
		response["worker_errors"] = errorMessages
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
