package v1

import (
	"fsfc/config"
	"fsfc/pkg"
	"github.com/gin-gonic/gin"
	"net/http"
)

//// GetLastSyncTime 获取上一次同步时间
//func GetLastSyncTime(c *gin.Context) {
//	db := DB.GetDB()
//
//	sqlStr := "select RsyncFile_Id, RsyncFile_Filename, RsyncFile_RsyncTime from rsyncfile order by RsyncFile_RsyncTime desc limit 1"
//	var r models.RsyncFile
//	err := db.QueryRow(sqlStr).Scan(&r.RsyncFileId, &r.RsyncFileFilename, &r.RsyncFileRsyncTime)
//	if err != nil {
//		//fmt.Printf("scan failed, err:%v\n", err)
//		logger.Logger.Error("scan failed", logger.Any("err", err))
//		c.JSON(http.StatusOK, response.FailMsg(err.Error()))
//	}
//
//	c.JSON(http.StatusOK, response.SuccessMsg(r.RsyncFileRsyncTime))
//}

// GetSyncGap 获取同步间隔
func GetSyncGap(c *gin.Context) {
	scanGap := config.Config.Set.ScanGap
	c.JSON(http.StatusOK, pkg.SuccessMsg(scanGap))
}

func GetStoreMemory(c *gin.Context) {

}
