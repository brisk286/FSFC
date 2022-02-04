package apis

import (
	"fsfc/config"
	"fsfc/fs"
	"fsfc/logger"
	"fsfc/response"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetFiles(c *gin.Context) {
	Fs := fs.GetFs()

	dirList, err := Fs.Walk()
	if err != nil {
		logger.Logger.Error("rootWrong", logger.Any("rootPath", config.GetConfig().Set.RootPath))
		c.JSON(http.StatusOK, response.FailMsg(err.Error()))
	}
	c.JSON(http.StatusOK, response.SuccessMsg(dirList))
}

func GetTest(c *gin.Context) {
	c.JSON(http.StatusOK, response.SuccessMsg("successGet"))
}
