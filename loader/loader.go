package loader

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Loader defines an abstraction for reading template files.
type Loader interface {
	ReadFile(path string) ([]byte, error)
	Walk(extension string, fn func(name string, content []byte) error) error
}

// FSLoader loads template files from an fs.FS (e.g. embed.FS or os.DirFS).
type FSLoader struct {
	fsys fs.FS
	dir  string
}

// NewFSLoader creates a loader backed by an fs.FS.
func NewFSLoader(fsys fs.FS, dir string) *FSLoader {
	return &FSLoader{
		fsys: fsys,
		dir:  filepath.Clean(dir),
	}
}

func (l *FSLoader) ReadFile(path string) ([]byte, error) {
	fullPath := path
	if l.dir != "" && l.dir != "." && !strings.HasPrefix(path, l.dir) {
		fullPath = filepath.Join(l.dir, path)
	}
	return fs.ReadFile(l.fsys, filepath.ToSlash(fullPath))
}

func (l *FSLoader) Walk(extension string, fn func(name string, content []byte) error) error {
	walkRoot := l.dir
	if walkRoot == "" {
		walkRoot = "."
	}

	return fs.WalkDir(l.fsys, filepath.ToSlash(walkRoot), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == extension {
			content, err := fs.ReadFile(l.fsys, filepath.ToSlash(path))
			if err != nil {
				return err
			}
			// Derive template name relative to walkRoot
			relPath, err := filepath.Rel(walkRoot, path)
			if err != nil {
				relPath = path
			}
			relPath = filepath.ToSlash(relPath)
			name := strings.TrimSuffix(relPath, extension)
			return fn(name, content)
		}
		return nil
	})
}

// DiskLoader loads template files directly from the local OS filesystem.
type DiskLoader struct {
	dir string
}

// NewDiskLoader creates a loader backed by the OS filesystem.
func NewDiskLoader(dir string) *DiskLoader {
	return &DiskLoader{
		dir: filepath.Clean(dir),
	}
}

func (l *DiskLoader) ReadFile(path string) ([]byte, error) {
	fullPath := path
	if !filepath.IsAbs(path) {
		fullPath = filepath.Join(l.dir, path)
	}
	return os.ReadFile(fullPath)
}

func (l *DiskLoader) Walk(extension string, fn func(name string, content []byte) error) error {
	return filepath.Walk(l.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == extension {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			relPath, err := filepath.Rel(l.dir, path)
			if err != nil {
				relPath = path
			}
			relPath = filepath.ToSlash(relPath)
			name := strings.TrimSuffix(relPath, extension)
			return fn(name, content)
		}
		return nil
	})
}
