package localstorage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("component", "localstorage")

type Storage struct {
}

// Init is mainly a placeholder for satisfying the top level storage interface.
func (s *Storage) Init() error {
	return nil
}

// Archive saves the file to the given path.
func (s *Storage) Write(reader io.Reader, fullPath string) error {
	// ensure file does not exist
	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("file at %s already exists", fullPath)
	}

	// create directories, if necessary
	path := filepath.Dir(fullPath)
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("creating directory path %s: %v", path, err)
	}

	// open file for writing
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", fullPath, err)
	}
	defer file.Close()

	// write to the file
	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("writing to file %s: %w", fullPath, err)
	}
	return nil
}
