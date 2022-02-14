package fs

import "os"

type FilePrimInfo struct {
	relaPath string
	os.FileInfo
}
