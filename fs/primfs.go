package fs

import (
	"os"
	"path/filepath"
)

type filesystem struct {
	root string
}

func (f *filesystem) walk() ([]string, error) {
	var files []string

	if err := filepath.Walk(f.root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	}); err != nil {
		panic(err)
	}

	return files, nil
}
