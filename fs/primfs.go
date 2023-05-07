package fs

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"fsfc/config"
	lnk "github.com/parsiya/golnk"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Filesystem struct {
	ScanGap         int
	ScanMlGapUpdate int
	ScanMlGapSync   int
	MlTopK          int
	LocalPath       string
	RemotePath      string
	RecentPath      string

	FileMlHash map[string]*FileMlInfo
	FileList   []*FileMlInfo

	LastFileMlHash map[string]*FileMlInfo
	LastFileList   []*FileMlInfo
}

var PrimFs Filesystem

func init() {
	PrimFs.ScanGap = config.Config.Set.ScanGap
	PrimFs.ScanMlGapUpdate = config.Config.Set.ScanMlGapUpdate
	PrimFs.ScanMlGapSync = config.Config.Set.ScanMlGapSync
	PrimFs.MlTopK = config.Config.Set.MlTopK
	PrimFs.LocalPath = config.Config.Set.LocalPath
	PrimFs.RemotePath = config.Config.Set.RemotePath
	PrimFs.RecentPath = config.Config.Set.RecentPath

	PrimFs.FileMlHash = map[string]*FileMlInfo{}
	PrimFs.FileList = []*FileMlInfo{}
}

// Walk 扫描LocalPath及下面的所有文件，包括文件夹
// 返回 所有文件的绝对路径
func (f *Filesystem) Walk() ([]string, error) {
	var files []string

	if err := filepath.Walk(f.LocalPath, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	}); err != nil {
		panic(err)
	}

	return files, nil
}

// WalkPath 扫描ph及下面的所有文件，包括文件夹
// 返回 所有文件的绝对路径
func WalkPath(ph string) ([]string, error) {
	var files []string

	if err := filepath.Walk(ph, func(path string, info os.FileInfo, err error) error {
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

// Scan 返回Path路径下所有文件信息
func (f *Filesystem) Scan() []FilePrimInfo {
	var fileInfos []FilePrimInfo

	files, err := f.Walk()
	if err != nil {
		panic(err)
	}

	// 匹配path和fileInfo
	for _, file := range files {
		stat, _ := os.Stat(file)
		if strings.Split(file, "\\")[len(strings.Split(file, "\\"))-1] == stat.Name() && !stat.IsDir() {
			// 将绝对路径转换为相对路径
			fileInfos = append(fileInfos, FilePrimInfo{file, strings.ReplaceAll(AbsToRela(file), "\\", "/"), stat})
		}
	}

	return fileInfos
}

// GetChangedFile 获取修改的文件信息
func (f *Filesystem) GetChangedFile() []FilePrimInfo {
	var changedFiles []FilePrimInfo

	fileInfos := f.Scan()

	lastScanTime := time.Now().Add(time.Duration(-f.ScanGap) * time.Second)
	for _, info := range fileInfos {
		if IsChanged(info.FileInfo, lastScanTime) == true {
			changedFiles = append(changedFiles, info)
		}
	}

	return RelaToAbsRemotePaths(changedFiles)
}

func (f *Filesystem) UpdateFileInfo() {
	// Recent文件夹中是快捷方式
	// C:\Users\14595\AppData\Roaming\Microsoft\Windows\Recent\2.cpp.lnk
	// 只要第一层文件
	lnkFilesInfo, err := ioutil.ReadDir(f.RecentPath)
	if err != nil {
		panic(err)
	}

	wrg := 0

	// 将快捷方式转实际地址
	var files []string
	for _, lnkFileInfo := range lnkFilesInfo {
		// 针对于文件夹、共享文件夹等非快捷方式
		if !strings.HasSuffix(lnkFileInfo.Name(), ".lnk") {
			wrg++
			continue
		}

		lnkPath := f.RecentPath + "\\" + lnkFileInfo.Name()
		absPath := LnkToAbs(lnkPath)
		files = append(files, absPath)
	}

	add, update := 0, 0

	// 报错：找不到快捷方式目标地址
	// 原因：1、目标文件在近期被移动过 2、系统文件，目标地址非标准格式
	// 解决：累计并continue
	for _, file := range files {
		fileInfo, err := os.Stat(file)
		if err != nil {
			wrg++
			continue
		}

		if fileInfo.IsDir() == false {
			if _, ok := f.FileMlHash[file]; !ok {
				add++
				f.addList(file, fileInfo)
			} else {
				update++
				f.updateList(file, fileInfo)
			}
		}
	}
	fmt.Println("【添加了】 ", add, "条数据")
	fmt.Println("【更新了】 ", update, "条数据")
	//fmt.Println("Avail:", len(files)-wrg)
}

func (f *Filesystem) addList(absPath string, info fs.FileInfo) {
	fileMlInfo := &FileMlInfo{
		absPath:    absPath,
		relaPath:   RelaToMlAbsRemotePath(strings.ReplaceAll(AbsToRela(absPath), "\\", "/")),
		modTime:    info.ModTime(),
		accessTime: GetFileAccessTime(info),
		createTime: GetFileCreateTime(info),
		label:      0,
		MlInfo: MlInfo{
			ModDist:      0,
			AccessDist:   0,
			CreateDist:   0,
			ModCnt:       0,
			AccessCnt:    0,
			ModAvgGap:    2 * float64(f.ScanMlGapSync),
			AccessAvgGap: 2 * float64(f.ScanMlGapSync),
			Size:         int(info.Size()),
			Category:     strings.Split(absPath, ".")[len(strings.Split(absPath, "."))-1],
		},
		FileInfo: info,
	}
	//fmt.Println(time.Now(), "【添加了】 ", AbsPath)

	f.FileMlHash[absPath] = fileMlInfo
	f.FileList = append(f.FileList, fileMlInfo)
}

func (f *Filesystem) updateList(absPath string, info fs.FileInfo) {
	lastScanTime := time.Now().Add(time.Duration(-f.ScanMlGapUpdate) * time.Second)

	if IsChanged(info, lastScanTime) {
		f.FileMlHash[absPath].modTime = info.ModTime()
		f.FileMlHash[absPath].accessTime = GetFileAccessTime(info)
		f.FileMlHash[absPath].MlInfo.ModCnt += 1
		f.FileMlHash[absPath].MlInfo.ModAvgGap = float64(f.ScanMlGapSync / f.FileMlHash[absPath].MlInfo.ModCnt)
		f.FileMlHash[absPath].MlInfo.Size = int(info.Size())
		f.FileMlHash[absPath].FileInfo = info
		fmt.Println(time.Now(), "【更新了】 文件的修改", absPath)
	}

	if IsAccessed(info, lastScanTime) {
		f.FileMlHash[absPath].accessTime = GetFileAccessTime(info)
		f.FileMlHash[absPath].MlInfo.AccessCnt += 1
		f.FileMlHash[absPath].MlInfo.AccessAvgGap = float64(f.ScanMlGapSync / f.FileMlHash[absPath].MlInfo.AccessCnt)
		f.FileMlHash[absPath].FileInfo = info
		fmt.Println(time.Now(), "【更新了】 文件的访问", absPath)
	}
}

// LnkToAbs 报错：找不到快捷方式目标地址
// 原因：1、目标文件在近期被移动过 2、系统文件，目标地址非标准格式
// 解决：累计并continue
func LnkToAbs(lnkPath string) string {
	lFile, err := lnk.File(lnkPath)
	if err != nil {
		panic(err)
	}

	absPath, err := simplifiedchinese.GBK.NewDecoder().String(lFile.LinkInfo.LocalBasePath + lFile.LinkInfo.CommonPathSuffix)

	return absPath
}

// 设置距离字段
func (f *Filesystem) setDist() {
	nowTime := time.Now()

	for _, fileMlInfo := range f.FileList {
		fileMlInfo.CreateDist = int(nowTime.Sub(fileMlInfo.createTime).Seconds())
		fileMlInfo.ModDist = int(nowTime.Sub(fileMlInfo.modTime).Seconds())
		fileMlInfo.AccessDist = int(nowTime.Sub(fileMlInfo.accessTime).Seconds())
	}
}

func (f *Filesystem) CollectData() {
	f.setDist()

	lastSyncTime := time.Now().Add(time.Duration(-f.ScanMlGapSync) * time.Second)
	if len(f.LastFileList) != 0 {
		for _, fileMlInfo := range f.FileList {
			if IsAccessed(fileMlInfo.FileInfo, lastSyncTime) {
				f.LastFileMlHash[fileMlInfo.absPath].label = 1
				//fmt.Println(time.Now(), "【赋值了】", fileMlInfo.AbsPath)
			}
		}
	}
	f.GenerateCsv()
	f.DeleteFsInfo()
}

func (f *Filesystem) GenerateCsv() {
	csvFile, err := os.OpenFile("data.csv", os.O_RDWR|os.O_APPEND, 0666)
	defer func() { csvFile.Close() }()
	if err != nil {
		csvFile, _ = os.Create("data.csv")
		_, err := csvFile.WriteString("\xEF\xBB\xBF")
		if err != nil {
			panic(err)
		}

		csvTitle := []string{
			"AbsPath",
			"ModDist",
			"AccessDist",
			"CreateDist",
			"ModCnt",
			"AccessCnt",
			"ModAvgGap",
			"AccessAvgGap",
			"Size",
			"Category",
			"Label",
		}

		// 初始化一个 csv writer，并通过这个 writer 写入数据到 csv 文件
		writer := csv.NewWriter(csvFile)
		err = writer.Write(csvTitle)
		if err != nil {
			panic(err)
		}
		writer.Flush()
		fmt.Println(time.Now(), "创建csv文件")
	}

	writer := csv.NewWriter(csvFile)

	// 每次写入的是上一个周期中修改的文件
	// 他们的label在本周期得到赋值
	// 他们的cnt和gap在上周期得到更新
	for _, lastFile := range f.LastFileList {
		line := []string{
			lastFile.absPath,
			strconv.Itoa(lastFile.MlInfo.ModDist),
			strconv.Itoa(lastFile.MlInfo.AccessDist),
			strconv.Itoa(lastFile.MlInfo.CreateDist),
			strconv.Itoa(lastFile.MlInfo.ModCnt),
			strconv.Itoa(lastFile.MlInfo.AccessCnt),
			fmt.Sprintf("%.2f", lastFile.MlInfo.ModAvgGap),
			fmt.Sprintf("%.2f", lastFile.MlInfo.AccessAvgGap),
			strconv.Itoa(lastFile.MlInfo.Size),
			lastFile.MlInfo.Category,
			strconv.Itoa(lastFile.label),
		}
		//将切片类型行数据写入 csv 文件
		err = writer.Write(line)
		if err != nil {
			panic(err)
		}
	}

	// 将 writer 缓冲中的数据都推送到 csv 文件，至此就完成了数据写入到 csv 文件
	writer.Flush()

	fmt.Println("【写入了】", len(f.LastFileList), "条数据")
}

func (f *Filesystem) DeleteFsInfo() {
	f.LastFileMlHash = f.FileMlHash
	f.LastFileList = f.FileList

	f.FileMlHash = map[string]*FileMlInfo{}
	f.FileList = []*FileMlInfo{}
}

// GetMlChangedFile post请求算法服务器，收到对应数据结构，分数结果进行排序，同步降序前TopK个文件，随后清空队列和字典
func (f *Filesystem) GetMlChangedFile() []FilePrimInfo {
	f.setDist()

	var predList []MlReq

	for _, fileMlInfo := range f.FileList {
		predList = append(predList, MlReq{
			RelaPath: fileMlInfo.relaPath,
			MlInfo:   fileMlInfo.MlInfo,
		})
	}

	bytesData, _ := json.Marshal(predList)
	// [{"RelaPath":"2.cpp","MlInfo":{"ModDist":78802585,"accessDist":129459,"CreateDist":78896132,"ModCnt":0,"AccessCnt":0,"ModAvgGap":120,"AccessAvgGap":120,"Size":2118,"Category":"cpp"}},
	// {"RelaPath":"2.任务书（设计）彭业诚.doc","MlInfo":{"ModDist":14698031,"accessDist":1346190,"CreateDist":15210402,"ModCnt":0,"AccessCnt":0,"ModAvgGap":120,"AccessAvgGap":120,"Size":46592,"Category":"doc"}}]

	url := "http://" + config.Config.Web.RemoteMlIp + ":" + config.Config.Web.RemoteMlPort + "/mlPred"
	res, err := http.Post(url,
		"application/json;charset=utf-8", bytes.NewBuffer(bytesData))
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
	}
	defer res.Body.Close()

	content, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
	}

	var mlResp []string
	err = json.Unmarshal(content, &mlResp)
	if err != nil {
		fmt.Println(err)
	}

	//lastScanTime := time.Now().Add(time.Duration(-f.ScanGap) * time.Second)
	//for _, info := range fileInfos {
	//	if IsChanged(info.FileInfo, lastScanTime) == true {
	//		changedFiles = append(changedFiles, info)
	//	}
	//}

	var changedMlFile []FilePrimInfo
	for _, relaFile := range mlResp {
		var absPath string

		for _, fileMlInfo := range f.FileList {
			if fileMlInfo.relaPath == relaFile {
				absPath = fileMlInfo.absPath
				break
			}
		}

		changedMlFile = append(changedMlFile, FilePrimInfo{
			AbsPath:  absPath,
			RelaPath: relaFile,
			FileInfo: nil,
		})
	}

	fmt.Println("系统算法模型检测到以下文件重要度达标，已自动同步")
	fmt.Println(mlResp)

	return changedMlFile
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

// RelaToAbsRemotePaths  相对路径转换为linux绝对路径
func RelaToAbsRemotePaths(filenames []FilePrimInfo) []FilePrimInfo {
	remotePath := config.Config.Set.RemotePath

	for i := 0; i < len(filenames); i++ {
		filenames[i].RelaPath = remotePath + "/" + filenames[i].RelaPath
	}

	return filenames
}

// RelaToMlAbsRemotePath Ml模块下的相对路径转换为linux绝对路径，ml模块的文件统一放在新建文件夹AutoSave中
func RelaToMlAbsRemotePath(filenames string) string {
	remotePath := config.Config.Set.RemotePath

	filenames = remotePath + "/" + "AutoSave" + "/" + filenames

	return filenames
}

func NanoToFileTime(sec int64) time.Time {
	return time.Unix(sec, 0)
}

func IsChanged(info fs.FileInfo, lastScanTime time.Time) bool {
	infoWin32 := info.Sys().(*syscall.Win32FileAttributeData)

	// 创建时间
	createTime := NanoToFileTime(infoWin32.CreationTime.Nanoseconds() / 1e9)

	// 比较修改时间、创建时间是否在上次同步时间后
	if info.ModTime().After(lastScanTime) || createTime.After(lastScanTime) {
		// 筛除文件夹
		if !info.IsDir() {
			return true
		}
	}
	return false
}

func IsAccessed(info fs.FileInfo, lastScanTime time.Time) bool {
	infoWin32 := info.Sys().(*syscall.Win32FileAttributeData)

	// 最后访问时间
	lastAccessTime := NanoToFileTime(infoWin32.LastAccessTime.Nanoseconds() / 1e9)

	// 比较最后访问时间是否在上次同步时间后
	if lastAccessTime.After(lastScanTime) {
		// 筛除文件夹
		if !info.IsDir() {
			return true
		}
	}
	return false
}

func GetFileCreateTime(info fs.FileInfo) time.Time {
	return NanoToFileTime(info.Sys().(*syscall.Win32FileAttributeData).CreationTime.Nanoseconds() / 1e9)
}

func GetFileAccessTime(info fs.FileInfo) time.Time {
	return NanoToFileTime(info.Sys().(*syscall.Win32FileAttributeData).LastAccessTime.Nanoseconds() / 1e9)
}
