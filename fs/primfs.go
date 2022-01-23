package fs

import (
	"os"
	"path/filepath"
)

type filesystem struct {
	root string
}

//扫描root下的所有文件，包括root
//返回所有文件的绝对路径
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
