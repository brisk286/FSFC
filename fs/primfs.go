package fs

import (
	"fsfc/config"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Filesystem struct {
	root string
}

var primfs Filesystem

func init() {
	primfs.root = config.GetConfig().Set.LocalPath
}

// Walk 扫描root下的所有文件，包括root
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

//output:
//0 D:\go\src\fsfc\primfs
//1 D:\go\src\fsfc\primfs\logs
//2 D:\go\src\fsfc\primfs\logs\chat.log
//3 D:\go\src\fsfc\primfs\primfs.go

func (f *Filesystem) Scan() []FilePrimInfo {
	var fileInfos []FilePrimInfo

	files, err := f.Walk()
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		stat, _ := os.Stat(file)
		fileInfos = append(fileInfos, FilePrimInfo{AbsToRela(file), stat})
	}

	return fileInfos
}

// GetChangedFile 传的是绝对地址
func (f *Filesystem) GetChangedFile() []string {
	var changedFiles []string

	fileInfos := primfs.Scan()
	remotePath := config.GetConfig().Set.RemotePath

	scanGap := config.GetConfig().Set.ScanGap
	lastScanTime := time.Now().Add(time.Duration(-scanGap) * time.Second)

	for _, info := range fileInfos {
		if info.ModTime().After(lastScanTime) && !info.IsDir() { //只会传文件，不传文件夹
			changedFiles = append(changedFiles, remotePath+"\\"+info.relaPath)
		}
	}

	return changedFiles
}

func GetFs() Filesystem {
	return primfs
}

// AbsToRela 如果找不到，可能是lastDir，传文件名
func AbsToRela(absPath string) string {
	var RelaPath string

	lastDir := "\\" + GetLastDir(config.GetConfig().Set.LocalPath) + "\\"

	if strings.Index(absPath, lastDir) != -1 {
		RelaPath = absPath[strings.Index(absPath, lastDir)+1:]
	} else {
		seqList := strings.Split(absPath, "\\")
		RelaPath = seqList[len(seqList)-1]
	}
	return RelaPath
}

func GetLastDir(path string) string {
	seqList := strings.Split(path, "\\")

	lastDir := seqList[len(seqList)-1]

	return lastDir
}
