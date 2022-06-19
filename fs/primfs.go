package fs

import (
	"fmt"
	"fsfc/config"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type Filesystem struct {
	root    string
	ScanGap int
}

var PrimFs Filesystem

func init() {
	PrimFs.root = config.Config.Set.LocalPath
	PrimFs.ScanGap = config.Config.Set.ScanGap
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

	stats, _ := ioutil.ReadDir(f.root)

	for _, file := range files {
		for _, stat := range stats {
			if strings.Split(file, "\\")[len(strings.Split(file, "\\"))-1] == stat.Name() {
				fileInfos = append(fileInfos, FilePrimInfo{strings.ReplaceAll(AbsToRela(file), "\\", "/"), stat})
			}
		}
	}

	return fileInfos
}

// GetChangedFile 传的是绝对地址
func (f *Filesystem) GetChangedFile() []string {
	var changedFiles []string

	fileInfos := PrimFs.Scan()

	scanGap := config.Config.Set.ScanGap
	lastScanTime := time.Now().Add(time.Duration(-scanGap) * time.Second)

	for _, info := range fileInfos {
		infoWin32 := info.FileInfo.Sys().(*syscall.Win32FileAttributeData)
		createTime := NanoToFileTime(infoWin32.CreationTime.Nanoseconds() / 1e9)
		lastAccessTime := NanoToFileTime(infoWin32.LastAccessTime.Nanoseconds() / 1e9)

		if info.FileInfo.ModTime().After(lastScanTime) || createTime.After(lastScanTime) || lastAccessTime.After(lastScanTime) { //只会传文件，不传文件夹
			if !info.FileInfo.IsDir() {
				changedFiles = append(changedFiles, info.relaPath)
				fmt.Println(lastAccessTime)
			}
		}
	}
	return RelaToAbsRemotePath(changedFiles)
}

// AbsToRela 如果找不到，可能是lastDir，传文件名
func AbsToRela(absPath string) string {
	var RelaPath string

	lastDir := "\\" + GetLastDir(config.Config.Set.LocalPath) + "\\"

	if strings.Index(absPath, lastDir) != -1 {
		RelaPath = absPath[strings.Index(absPath, lastDir)+1:]
	} else {
		seqList := strings.Split(absPath, "\\")
		RelaPath = seqList[len(seqList)-1]
	}
	return RelaPath
}

func FixDir(localPath string) string {
	lastDir := GetLastDir(localPath)
	return localPath[:len(localPath)-len(lastDir)]
}

func GetLastDir(path string) string {
	seqList := strings.Split(path, "\\")
	lastDir := seqList[len(seqList)-1]

	return lastDir
}

func RelaToAbsRemotePath(filenames []string) []string {
	remotePath := config.Config.Set.RemotePath

	for i := 0; i < len(filenames); i++ {
		filenames[i] = remotePath + "/" + filenames[i]
	}

	return filenames
}

func NanoToFileTime(sec int64) time.Time {
	return time.Unix(sec, 0)
}
