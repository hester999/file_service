package entity

import "time"

type Data struct {
	ID        int            `yaml:"id"`
	Name      string         `yaml:"name"`
	Timestamp time.Time      `yaml:"timestamp"`
	Values    []float64      `yaml:"values"`
	Metadata  map[string]int `yaml:"metadata"`
}
