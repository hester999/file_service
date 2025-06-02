package handlers

import (
	"context"
	"net/http"
	"service/internal/entity"
)

type Generate interface {
	GenerateHandler(w http.ResponseWriter, r *http.Request)
}
type Process interface {
	GetFilesHandler(w http.ResponseWriter, r *http.Request)
	GetFileHandler(w http.ResponseWriter, r *http.Request)
}

type FileService interface {
	Process(ctx context.Context, cfg entity.WorkerConf) ([]error, error)
	GetFiles(ctx context.Context) ([]string, error)
	GetFile(ctx context.Context, id int) (entity.Data, error)
}

type ServiceHandlers struct {
	generate Generate
	process  Process
}

func (s *ServiceHandlers) GenerateHandler(w http.ResponseWriter, r *http.Request) {
	s.generate.GenerateHandler(w, r)
}

func (s *ServiceHandlers) GetFilesHandler(w http.ResponseWriter, r *http.Request) {
	s.process.GetFilesHandler(w, r)
}

func (s *ServiceHandlers) GetFileHandler(w http.ResponseWriter, r *http.Request) {
	s.process.GetFileHandler(w, r)
}

func NewServiceHandlers(generate Generate, process Process) *ServiceHandlers {
	return &ServiceHandlers{
		generate: generate,
		process:  process,
	}
}
