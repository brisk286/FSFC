package request

import "fsfc/rsync"

type RsyncOpsReq struct {
	Filename       string          `json:"filename"`
	RsyncOps       []rsync.RSyncOp `json:"rsyncOps"`
	ModifiedLength int32           `json:"ModifiedLength"`
}
