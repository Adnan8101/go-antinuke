package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type LogRotation struct {
	maxSize     int64
	maxAge      time.Duration
	currentFile string
}

func NewLogRotation(maxSize int64, maxAge time.Duration) *LogRotation {
	return &LogRotation{
		maxSize: maxSize,
		maxAge:  maxAge,
	}
}

func (lr *LogRotation) ShouldRotate(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	if info.Size() >= lr.maxSize {
		return true
	}

	age := time.Since(info.ModTime())
	return age >= lr.maxAge
}

func (lr *LogRotation) Rotate(path string) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	ext := filepath.Ext(path)
	base := path[:len(path)-len(ext)]

	newPath := fmt.Sprintf("%s-%s%s", base, timestamp, ext)

	err := os.Rename(path, newPath)
	return newPath, err
}
