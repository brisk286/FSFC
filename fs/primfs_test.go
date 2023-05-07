package fs

import (
	"fmt"
	lnk "github.com/parsiya/golnk"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"
)

func Test_Walk(t *testing.T) {
	root, _ := os.Getwd()

	fmt.Println(root)

	//primfs := &Filesystem{root: root}
	//
	//files, _ := primfs.Walk()
	//Fs := &Filesystem{LocalPath: "C:\\Users\\14595\\AppData\\Roaming\\Microsoft\\Windows\\Recent"}
	Fs := PrimFs

	dirList, _ := Fs.Walk()

	for i, file := range dirList {
		fmt.Println(i, file)
	}
}

func Test_Name(t *testing.T) {
	fileInfo, _ := os.Stat("testfile.txt")

	//fmt.Println(fileInfo.Name())
	//fmt.Println(fileInfo.Mode())
	//fmt.Println(fileInfo.IsDir())
	fmt.Println("修改时间：", fileInfo.ModTime())
	fmt.Println("文件大小：", fileInfo.Size())
	//fmt.Println("修改时间：", fileInfo.Sys())
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
	//fileAttributes := fileSys.FileAttributes
	//fmt.Println(fileAttributes)

	nanoseconds := fileSys.CreationTime.Nanoseconds() // 返回的是纳秒
	createTime := nanoseconds / 1e9                   //秒
	fmt.Println("创建时间:", NanoToFileTime(createTime))
	nanoseconds = fileSys.LastAccessTime.Nanoseconds() // 返回的是纳秒
	lastAccessTime := nanoseconds / 1e9                //秒
	fmt.Println("最后访问时间:", NanoToFileTime(lastAccessTime))
	nanoseconds = fileSys.LastWriteTime.Nanoseconds() // 返回的是纳秒
	lastWriteTime := nanoseconds / 1e9                //秒
	fmt.Println("最后修改时间:", NanoToFileTime(lastWriteTime))
}

func Test_Scan(t *testing.T) {
	fileInfos := PrimFs.Scan()

	for _, info := range fileInfos {
		fmt.Println(info.Name())
		//fmt.Println(info.RelaPath)
		//fmt.Println(info.IsDir())
		//fmt.Println(info.ModTime())
	}
}

func Test_ChangedFile(t *testing.T) {
	fileInfos := PrimFs.GetChangedFile()

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

func Test_LastAccessTime(t *testing.T) {
	//fileInfos := primfs.Scan()
	//path := "C:\\Users\\14595\\Desktop\\FSFC\\fsfc_windows"
	//
	//fileInfos, _ := ioutil.ReadDir(path)
	fileInfos := PrimFs.Scan()

	for _, info := range fileInfos {
		infoWin32 := info.FileInfo.Sys().(*syscall.Win32FileAttributeData)
		lastAccessTime := NanoToFileTime(infoWin32.LastAccessTime.Nanoseconds() / 1e9)
		fmt.Println(info.Name(), lastAccessTime)
	}
	//fmt.Println("_________")
	//
	//fileInfosS := primfs.Scan()
	//
	//for _, infoS := range fileInfosS {
	//	fmt.Println(infoS.FileInfo.Name())
	//}
}

func Test_RecentDir(t *testing.T) {
	Fs := &Filesystem{LocalPath: "C:\\Users\\14595\\AppData\\Roaming\\Microsoft\\Windows\\Recent"}
	//Fs := PrimFs

	//dirList, _ := Fs.WalkLocalPath()

	//ioutil.ReadDir(Fs.root)

	stats, _ := ioutil.ReadDir(Fs.LocalPath)

	for _, stat := range stats {
		fmt.Println(stat.Name())
	}

	//for i, file := range dirList {
	//	fmt.Println(i, file)
	//}
}

func Test_Lnk(t *testing.T) {
	Lnk, err := lnk.File("C:\\Users\\14595\\AppData\\Roaming\\Microsoft\\Windows\\Recent\\p2525983101.jpg.lnk")
	if err != nil {
		panic(err)
	}

	targetPath := ""
	if Lnk.LinkInfo.LocalBasePath != "" {
		targetPath = Lnk.LinkInfo.LocalBasePath
	}
	if Lnk.LinkInfo.LocalBasePathUnicode != "" {
		targetPath = Lnk.LinkInfo.LocalBasePathUnicode
	}

	// 中文路径需要解码，英文路径可忽略
	targetPath, _ = simplifiedchinese.GBK.NewDecoder().String(targetPath)
	fmt.Println("BasePath", targetPath)
}

func Test_Lnk2(t *testing.T) {
	Lnk, err := lnk.File("C:\\Users\\14595\\Desktop\\FinalShell.lnk")
	//Lnk, err := lnk.File("C:\\Users\\14595\\AppData\\Roaming\\Microsoft\\Windows\\Recent\\德育-20195325彭业诚.pptx.lnk")
	//Lnk, err := lnk.File("C:\\Users\\14595\\AppData\\Roaming\\Microsoft\\Windows\\Recent\\教程地址.txt.lnk")
	//f, err := os.Open("C:\\Users\\14595\\AppData\\Roaming\\Microsoft\\Windows\\Recent\\德育-20195325彭业诚.pptx.lnk")
	//if err != nil {
	//	return
	//}
	//fmt.Println(f.Stat())

	if err != nil {
		panic(err)
	}

	// 中文路径需要解码，英文路径可忽略
	targetPath, _ := simplifiedchinese.GBK.NewDecoder().String(Lnk.LinkInfo.LocalBasePath + Lnk.LinkInfo.CommonPathSuffix)
	fmt.Println(targetPath)
	fmt.Println(Lnk)
}

func Test_A(t *testing.T) {
	//str := "C:\\Users\\14595\\AppData\\Roaming\\Microsoft\\Windows\\Recent\\*.lnk"
	//str := "D:\\go\\src\\f
	str := "C:\\Users\\14595\\AppData\\Roaming\\Microsoft\\Windows\\Recent"

	files, _ := ioutil.ReadDir(str)
	//files, err := filepath.Glob(str)
	//if err != nil {
	//	return
	//}

	fmt.Println(len(files))
	cnt := 0
	wrg := 0
	for _, f := range files {
		fmt.Println(f.Name())

		// 针对于共享文件夹
		if !strings.HasPrefix(f.Name(), ".lnk") {
			wrg++
			continue
		}

		abs := str + "\\" + f.Name()
		absPath := LnkToAbs(abs)

		fmt.Println(absPath)

		// 报错：找不到快捷方式目标地址
		// 原因：1、目标文件在近期被移动过 2、系统文件，目标地址非标准格式
		// 解决：累计并continue
		info, err := os.Stat(absPath)
		if err != nil {
			//panic(err)
			wrg++
			continue
		}

		if info.IsDir() == true {
			cnt++
		}
	}
	fmt.Println(wrg)
	fmt.Println(cnt)
}

func Test_Break(t *testing.T) {
priority:
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			if j == 50 {
				break priority
			}
		}
		if i == 2 {
			fmt.Println(2)
		}
	}

	fmt.Println(1)
}
