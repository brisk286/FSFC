package models

import "time"

type RsyncFile struct {
	RsyncFileId        int
	RsyncFileFilename  string
	RsyncFileRsyncTime time.Time
}
