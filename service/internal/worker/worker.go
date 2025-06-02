package worker

import (
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"service/internal/entity"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	channelBufferSize = 10000
)

// Worker - структура для многопоточной обработки файлов
type Worker struct {
	wg            sync.WaitGroup         // Для синхронизации горутин
	maxFiles      int                    // Количество файлов
	maxWorkers    int                    // Количество воркеров
	maxIterations int                    // Количество итераций
	taskCh        chan int               // Канал для передачи задач
	errCh         chan error             // Канал для сбора ошибок
	logger        *slog.Logger           // Логгер
	mutexMap      map[string]*sync.Mutex // Мьютексы для синхронизации записи в файлы
	filesPath     string                 // Путь к директории с файлами
	startTime     time.Time              // Время начала работы
}

// NewWorker создает новый экземпляр Worker
func NewWorker(cfg entity.WorkerConf, logger *slog.Logger) *Worker {

	exePath, err := os.Executable()
	if err != nil {
		logger.Error("Failed to get executable path", err)
		exePath = "."
	}

	projectPath := filepath.Dir(filepath.Dir(exePath))
	filesPath := filepath.Join(projectPath, "internal", "files")

	if err := os.MkdirAll(filesPath, 0755); err != nil {
		logger.Error("Failed to create files directory", "error", err)
	}

	totalOperations := cfg.Iterations * cfg.MaxFiles
	bufferSize := min(channelBufferSize, totalOperations)

	w := &Worker{
		maxFiles:      cfg.MaxFiles,
		maxWorkers:    cfg.MaxWorkers,
		maxIterations: cfg.Iterations,
		taskCh:        make(chan int, bufferSize),
		errCh:         make(chan error, bufferSize),
		logger:        logger,
		mutexMap:      make(map[string]*sync.Mutex, cfg.MaxFiles),
		filesPath:     filesPath,
		startTime:     time.Now(),
	}

	// Создаем мьютексы для всех файлов
	for i := 0; i < cfg.MaxFiles; i++ {
		filename := filepath.Join(filesPath, fmt.Sprintf("output_%d.yml", i))
		w.mutexMap[filename] = &sync.Mutex{}
	}

	logger.Info("Worker initialized",
		"maxFiles", cfg.MaxFiles,
		"maxWorkers", cfg.MaxWorkers,
		"iterations", cfg.Iterations,
		"totalOperations", totalOperations,
		"bufferSize", bufferSize)

	return w

}

// GenerateTask генерирует задачи для обработки
func (w *Worker) GenerateTask() {

	for i := 0; i < w.maxIterations; i++ {
		for id := 0; id < w.maxFiles; id++ {
			taskID := i*w.maxFiles + id
			w.taskCh <- taskID

		}
	}
	close(w.taskCh)
}

// ProcessData обрабатывает задачи и записывает данные в файлы
func (w *Worker) ProcessData() {
	w.wg.Add(w.maxWorkers)
	for i := 0; i < w.maxWorkers; i++ {

		go func(workerID int) {
			defer w.wg.Done()
			for taskID := range w.taskCh {
				data := w.generateRandomData(taskID)
				filename := filepath.Join(w.filesPath, fmt.Sprintf("output_%d.yml", taskID%w.maxFiles))
				lock := w.mutexMap[filename]
				lock.Lock()
				err := w.writeToFile(data, filename)
				lock.Unlock()
				if err != nil {
					w.errCh <- err
				}
			}
		}(i)
	}
	w.wg.Wait()
	close(w.errCh)

}

// generateRandomData генерирует случайные данные для записи
func (w *Worker) generateRandomData(id int) entity.Data {
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))

	name := fmt.Sprintf("Item-%d-%d", id, r.Intn(10000))

	values := make([]float64, r.Intn(10)+1)
	for i := range values {
		values[i] = r.Float64() * 100
	}

	metadata := map[string]int{
		"alpha": r.Intn(1000),
		"beta":  r.Intn(1000),
		"gamma": r.Intn(1000),
		"delta": r.Intn(1000),
	}

	return entity.Data{
		ID:        id,
		Name:      name,
		Timestamp: time.Now(),
		Values:    values,
		Metadata:  metadata,
	}
}

func (w *Worker) writeToFile(data entity.Data, filename string) error {

	yamlData, err := yaml.Marshal(&data)
	if err != nil {
		w.logger.Error("Failed to marshal YAML",
			"file", filename,
			"error", err)
		return err
	}

	err = os.WriteFile(filename, yamlData, 0644)
	if err != nil {
		w.logger.Error("Failed to write file",
			"file", filename,
			"error", err)
		return err
	}

	return nil

}

// Errors возвращает канал с ошибками
func (w *Worker) Errors() <-chan error {
	return w.errCh
}
