package main

import (
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"os"
	"service/internal/handlers"
	"service/internal/usecases"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	fileService := usecases.NewFileService(logger)
	generator := handlers.NewGenerateImpl(fileService)
	process := handlers.NewProcessHandler(fileService)

	handlers := handlers.NewServiceHandlers(generator, process)
	router := mux.NewRouter()

	router.StrictSlash(true)

	router.HandleFunc("/file/{id}", handlers.GetFileHandler).Methods(http.MethodGet)
	router.HandleFunc("/files", handlers.GetFilesHandler).Methods(http.MethodGet)
	router.HandleFunc("/generate", handlers.GenerateHandler).Methods(http.MethodPost)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		return
	}

}
