package queue

import (
	"fmt"
	"os"
	"path/filepath"
)

func Deploy(queueFile string) error {
	dir := filepath.Dir(queueFile)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create queue dir: %w", err)
	}

	return os.WriteFile(queueFile, []byte("deploy\n"), 0600)
}