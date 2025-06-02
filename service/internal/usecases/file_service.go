package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"service/internal/entity"
	"service/internal/utils"
	"service/internal/worker"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

type Worker interface {
	GenerateTask()
	ProcessData()
	Errors() <-chan error
}

type FileService interface {
	Process(ctx context.Context, cfg entity.WorkerConf) ([]error, error)
	GetFiles(ctx context.Context) ([]string, error)
	GetFile(ctx context.Context, id int) (entity.Data, error)
}

type FileServiceImpl struct {
	logger *slog.Logger
	wg     sync.WaitGroup
	worker Worker
}

func NewFileService(logger *slog.Logger) *FileServiceImpl {
	return &FileServiceImpl{
		logger: logger,
	}
}

func (f *FileServiceImpl) Process(ctx context.Context, cfg entity.WorkerConf) ([]error, error) {
	filesPath, err := utils.EnsureFilesDirectory()
	if err != nil {
		return nil, fmt.Errorf("failed to ensure files directory: %v", err)
	}
	f.logger.Info("Files directory path", "path", filesPath)

	f.worker = worker.NewWorker(cfg, f.logger)
	f.logger.Info("Worker created",
		"maxFiles", cfg.MaxFiles,
		"maxWorkers", cfg.MaxWorkers,
		"iterations", cfg.Iterations,
		"timeout", cfg.Timeout)

	var errors []error
	errorChan := make(chan error, cfg.Iterations*cfg.MaxFiles)

	go func() {
		for err := range f.worker.Errors() {
			errorChan <- err
			f.logger.Error("Worker error", "error", err)
		}
		close(errorChan)
	}()

	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		f.logger.Info("Starting task generation")
		f.worker.GenerateTask()
		f.logger.Info("Starting data processing")
		f.worker.ProcessData()
		f.logger.Info("Data processing completed")
	}()

	done := make(chan struct{})
	go func() {
		f.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		f.logger.Error("Operation cancelled by context")
		return nil, ctx.Err()
	case <-time.After(time.Duration(cfg.Timeout) * time.Second):
		f.logger.Error("Operation timed out")
		return nil, nil
	case <-done:
		f.logger.Info("Operation completed successfully")
	}

	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		f.logger.Error("Errors occurred during processing", "count", len(errors))
	} else {
		f.logger.Info("No errors during processing")
	}

	return errors, nil
}

func (f *FileServiceImpl) GetFiles(ctx context.Context) ([]string, error) {
	filesPath, err := utils.GetFilesPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get files path: %v", err)
	}
	f.logger.Info("Looking for files in", "path", filesPath)

	files, err := os.ReadDir(filesPath)
	if err != nil {
		f.logger.Error("Failed to read directory", "path", filesPath, "error", err)
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	var result []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".yml" {
			result = append(result, file.Name())
			f.logger.Debug("Found file", "name", file.Name())
		}
	}

	f.logger.Info("Found files", "count", len(result), "files", result)
	return result, nil
}

func (f *FileServiceImpl) GetFile(ctx context.Context, id int) (entity.Data, error) {
	filesPath, err := utils.GetFilesPath()
	if err != nil {
		return entity.Data{}, fmt.Errorf("failed to get files path: %v", err)
	}

	filename := filepath.Join(filesPath, fmt.Sprintf("output_%d.yml", id))

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			f.logger.Info("File not found", "path", filename)
			return entity.Data{}, fmt.Errorf("file not found: output_%d.yml", id)
		}
		f.logger.Error("Failed to read file", "path", filename, err)
		return entity.Data{}, err
	}

	var result entity.Data
	if err := yaml.Unmarshal(data, &result); err != nil {
		f.logger.Error("Failed to unmarshal YAML", "path", filename, err)
		return entity.Data{}, fmt.Errorf("failed to unmarshal YAML: %v", err)
	}

	f.logger.Info("File read successfully", "path", filename)
	return result, nil
}

func (f *FileServiceImpl) ProcessFiles(ctx context.Context, cfg entity.WorkerConf) error {

	w := worker.NewWorker(cfg, f.logger)

	errCh := w.Errors()

	
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.GenerateTask()
		w.ProcessData()
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		f.logger.Error("Operation cancelled by context")
		return ctx.Err()
	case <-time.After(time.Duration(cfg.Timeout) * time.Second):
		f.logger.Error("Operation timed out")
		return nil
	case <-done:
		f.logger.Info("Operation completed successfully")
	}

	
	var errors []error
	for err := range errCh {
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		f.logger.Error("Errors occurred during processing", "count", len(errors))
		return errors[0]
	}

	return nil
}
