package v1

import (
	"fsfc/fs"
	"fsfc/logger"
	"fsfc/pkg/response"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetFiles 获取root路径下所有文件的绝对路径
func GetFiles(c *gin.Context) {
	dirList, err := fs.PrimFs.Walk()
	if err != nil {
		logger.Logger.Error("rootWrong", logger.Any("LocalPath", fs.PrimFs.LocalPath))
		c.JSON(http.StatusOK, response.FailMsg(err.Error()))
	}
	c.JSON(http.StatusOK, response.SuccessMsg(dirList))
}

// GetTest 测试接口
func GetTest(c *gin.Context) {
	c.JSON(http.StatusOK, response.SuccessMsg("successGet"))
}
