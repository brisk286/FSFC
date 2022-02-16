package response

import (
	"fsfc/rsync"
)

type BlockHashesReps struct {
	Code int                     `json:"code"`
	Msg  string                  `json:"msg"`
	Data []rsync.FileBlockHashes `json:"data"`
}
