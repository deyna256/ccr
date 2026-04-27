package task

import (
	"os"
	"path/filepath"
)

type OsFileStorage struct {
	baseDir string
}

func NewOsFileStorage(baseDir string) *OsFileStorage {
	return &OsFileStorage{baseDir: baseDir}
}

func (s *OsFileStorage) MkdirAll(path string, perms uint32) error {
	return os.MkdirAll(filepath.Join(s.baseDir, path), os.FileMode(perms))
}

func (s *OsFileStorage) Create(taskID, filename string) (File, error) {
	path := filepath.Join(s.baseDir, taskID, filename)
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return &osFile{File: f, path: path}, nil
}

func (s *OsFileStorage) Open(taskID string) (File, error) {
	return nil, ErrFileNotFound
}

func (s *OsFileStorage) Remove(path string) error {
	return os.Remove(filepath.Join(s.baseDir, path))
}

type osFile struct {
	*os.File
	path string
}

func (f *osFile) Path() string {
	return f.path
}