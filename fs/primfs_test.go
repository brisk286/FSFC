package fs

import (
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"
)

func Test_Walk(t *testing.T) {
	root, _ := os.Getwd()

	fmt.Println("_____", root)

	//primfs := &Filesystem{root: root}
	//
	//files, _ := primfs.Walk()
	Fs := GetFs()

	dirList, _ := Fs.Walk()

	for i, file := range dirList {
		fmt.Println(i, file)
	}
}

func Test_Name(t *testing.T) {
	fileInfo, _ := os.Stat("primfs.go")

	fmt.Println(fileInfo.Name())
	fmt.Println(fileInfo.Mode())
	fmt.Println(fileInfo.IsDir())
	fmt.Println(fileInfo.ModTime())
	fmt.Println(fileInfo.Size())
	fmt.Println(fileInfo.Sys())
	//primfs.go
	//	-rw-rw-rw-  # 权限
	//		false  # 是否是文件夹
	//2022-02-04 16:22:15.1663344 +0800 CST # 修改时间
	//598  # 字节
	//&{32 {276763625 30936383} {1368681684 30939552} {1358771952 30939552} 0 598}
	fileSys := fileInfo.Sys().(*syscall.Win32FileAttributeData)
	//type Win32FileAttributeData struct {
	//	FileAttributes uint32
	//	CreationTime   Filetime
	//	LastAccessTime Filetime
	//	LastWriteTime  Filetime
	//	FileSizeHigh   uint32
	//	FileSizeLow    uint32
	//}
	fileAttributes := fileSys.FileAttributes
	fmt.Println(fileAttributes)

	nanoseconds := fileSys.CreationTime.Nanoseconds() // 返回的是纳秒
	createTime := nanoseconds / 1e9                   //秒
	fmt.Println(createTime)
}

func Test_Scan(t *testing.T) {
	fileInfos := primfs.Scan()

	for _, info := range fileInfos {
		fmt.Println(info.Name())
		fmt.Println(info.relaPath)
		//fmt.Println(info.IsDir())
		//fmt.Println(info.ModTime())
	}
}

func Test_ChangedFile(t *testing.T) {
	fileInfos := primfs.GetChangedFile()

	for _, info := range fileInfos {
		fmt.Println(info)
	}
}

func Test_AtR(t *testing.T) {
	filePath := "D:\\go\\src\\fsfc\\fs\\primfs.go"

	fmt.Println(AbsToRela(filePath))
}

func Test_CreateFile(t *testing.T) {
	//os.Create("123.exe")

	//time.Sleep(2 * time.Second)
	stat, _ := os.Stat("C:\\Users\\14595\\Desktop\\重要资料\\简历-彭业诚.pdf")

	// Sys()返回的是interface{}，所以需要类型断言，不同平台需要的类型不一样，linux上为*syscall.Stat_t
	stat_t := stat.Sys().(*syscall.Win32FileAttributeData)
	//fmt.Println(stat_t)
	// atime，ctime，mtime分别是访问时间，创建时间和修改时间，具体参见man 2 stat
	fmt.Println(timespecToTime(stat_t.CreationTime.Nanoseconds() / 1e9))
	fmt.Println(timespecToTime(stat_t.LastAccessTime.Nanoseconds() / 1e9))
	fmt.Println(timespecToTime(stat_t.LastWriteTime.Nanoseconds() / 1e9))
	fmt.Println(stat.ModTime())

	createTime := timespecToTime(stat_t.CreationTime.Nanoseconds() / 1e9)
	fmt.Println(createTime.After(time.Now()))
}

func timespecToTime(sec int64) time.Time {
	return time.Unix(sec, 0)
}
