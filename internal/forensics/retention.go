package forensics

import (
	"os"
	"path/filepath"
	"time"
)

type RetentionPolicy struct {
	RetentionDays int
	ArchivePath   string
}

type RetentionManager struct {
	policy *RetentionPolicy
}

func NewRetentionManager(retentionDays int, archivePath string) *RetentionManager {
	return &RetentionManager{
		policy: &RetentionPolicy{
			RetentionDays: retentionDays,
			ArchivePath:   archivePath,
		},
	}
}

func (rm *RetentionManager) Cleanup(logPath string) error {
	entries, err := os.ReadDir(logPath)
	if err != nil {
		return err
	}

	cutoff := time.Now().AddDate(0, 0, -rm.policy.RetentionDays)

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			fullPath := filepath.Join(logPath, entry.Name())
			os.Remove(fullPath)
		}
	}

	return nil
}

func (rm *RetentionManager) Archive(sourcePath, destPath string) error {
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	return os.WriteFile(destPath, data, 0644)
}
