package utils

import (
	"os"
	"path/filepath"
)

func GetFilesPath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	projectPath := filepath.Dir(filepath.Dir(exePath))

	return filepath.Join(projectPath, "internal", "files"), nil
}

func EnsureFilesDirectory() (string, error) {
	filesPath, err := GetFilesPath()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filesPath, 0755); err != nil {
		return "", err
	}

	return filesPath, nil
}
