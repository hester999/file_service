package entity

import "time"

type WorkerConf struct {
	Iterations int           `json:"iterations"`
	MaxWorkers int           `json:"max_workers"`
	MaxFiles   int           `json:"max_files"`
	Timeout    time.Duration `json:"timeout_ms"`
}
