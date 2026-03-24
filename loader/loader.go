// Package loader provides document loader implementations.
package loader

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/njchilds90/goragkit/document"
)

// Loader reads documents from a source.
type Loader interface {
	Load(ctx context.Context) ([]document.Document, error)
}

// Directory loads text-like files recursively from disk.
type Directory struct {
	Path       string
	Extensions map[string]struct{}
}

// NewDirectory returns a directory loader.
func NewDirectory(path string) *Directory {
	return &Directory{Path: path, Extensions: map[string]struct{}{".txt": {}, ".md": {}, ".go": {}, ".json": {}}}
}

// Load implements Loader.
func (d *Directory) Load(_ context.Context) ([]document.Document, error) {
	out := make([]document.Document, 0)
	err := filepath.WalkDir(d.Path, func(path string, dir fs.DirEntry, err error) error {
		if err != nil || dir.IsDir() {
			return err
		}
		ext := strings.ToLower(filepath.Ext(path))
		if _, ok := d.Extensions[ext]; !ok {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out = append(out, document.Document{ID: path, Text: string(b), Metadata: map[string]string{"path": path, "ext": ext}})
		return nil
	})
	return out, err
}
