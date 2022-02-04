package fs

import (
	"fsfc/config"
	"os"
	"path/filepath"
)

type Filesystem struct {
	root string
}

var fs Filesystem

func init() {
	fs.root = config.GetConfig().Set.RootPath
}

//扫描root下的所有文件，包括root
//返回所有文件的绝对路径
func (f *Filesystem) Walk() ([]string, error) {
	var files []string

	if err := filepath.Walk(f.root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	}); err != nil {
		panic(err)
	}

	return files, nil
}

func GetFs() Filesystem {
	return fs
}
