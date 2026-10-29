package host

import (
	"io/fs"
	"os"
	"path/filepath"
)

// File is a file being written: the part of an open file an install uses.
type File interface {
	Name() string
	Write(raw []byte) (int, error)
	Chmod(mode fs.FileMode) error
	Close() error
}

// Files is the file system a colourway is installed on.
type Files interface {
	Stat(name string) (fs.FileInfo, error)
	Lstat(name string) (fs.FileInfo, error)
	ReadDir(name string) ([]fs.DirEntry, error)
	ReadFile(name string) ([]byte, error)
	EvalSymlinks(path string) (string, error)
	MkdirAll(path string, mode fs.FileMode) error
	CreateTemp(dir, pattern string) (File, error)
	Remove(name string) error
	Rename(from, to string) error
}

// onDisk is the file system of the machine.
type onDisk struct{}

// NewDisk is the file system of the machine.
func NewDisk() Files { return onDisk{} }

// Stat is os.Stat.
func (onDisk) Stat(name string) (fs.FileInfo, error) { return os.Stat(name) }

// Lstat is os.Lstat.
func (onDisk) Lstat(name string) (fs.FileInfo, error) { return os.Lstat(name) }

// ReadDir is os.ReadDir.
func (onDisk) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

// ReadFile is os.ReadFile.
func (onDisk) ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }

// EvalSymlinks is filepath.EvalSymlinks.
func (onDisk) EvalSymlinks(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}

// MkdirAll is os.MkdirAll.
func (onDisk) MkdirAll(path string, mode fs.FileMode) error {
	return os.MkdirAll(path, mode)
}

// CreateTemp is os.CreateTemp. When it fails it gives no file, and not a nil
// file wrapped in the interface.
func (onDisk) CreateTemp(dir, pattern string) (File, error) {
	opened, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return nil, err
	}
	return opened, nil
}

// Remove is os.Remove.
func (onDisk) Remove(name string) error { return os.Remove(name) }

// Rename is os.Rename.
func (onDisk) Rename(from, to string) error { return os.Rename(from, to) }
