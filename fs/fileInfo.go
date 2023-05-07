package fs

import (
	"os"
	"time"
)

type FilePrimInfo struct {
	AbsPath  string // Windows
	RelaPath string // linux
	os.FileInfo
}

type FileMlInfo struct {
	absPath    string
	relaPath   string
	modTime    time.Time
	accessTime time.Time
	createTime time.Time
	label      int

	MlInfo
	os.FileInfo
}

type MlReq struct {
	RelaPath string `json:"RelaPath"`

	MlInfo `json:"MlInfo"`
}

type MlResp struct {
	RelaPath string `json:"RelaPath"`
}

type MlInfo struct {
	ModDist      int     `json:"ModDist"`
	AccessDist   int     `json:"AccessDist"`
	CreateDist   int     `json:"CreateDist"`
	ModCnt       int     `json:"ModCnt"`
	AccessCnt    int     `json:"AccessCnt"`
	ModAvgGap    float64 `json:"ModAvgGap"`
	AccessAvgGap float64 `json:"AccessAvgGap"`
	Size         int     `json:"Size"`
	Category     string  `json:"Category"`
}
