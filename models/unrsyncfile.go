package models

import "time"

type UnRsyncFile struct {
	UnRsyncFileId        int
	UnRsyncFileFilename  string
	UnRsyncFileRsyncTime time.Time
}
