package finder

import (
	"os"
	"path/filepath"
	"time"
)

func FindModifiedFiles(limitTime time.Time) ([]string, error) {
	var results []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.ModTime().After(limitTime) {
			results = append(results, path)
		}
		return nil
	})

	return results, err
}
