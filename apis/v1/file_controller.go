package v1

import (
	"fsfc/config"
	"fsfc/fs"
	"fsfc/logger"
	"fsfc/pkg/response"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetFiles(c *gin.Context) {
	Fs := fs.PrimFs

	dirList, err := Fs.Walk()
	if err != nil {
		logger.Logger.Error("rootWrong", logger.Any("LocalPath", config.Config.Set.LocalPath))
		c.JSON(http.StatusOK, response.FailMsg(err.Error()))
	}
	c.JSON(http.StatusOK, response.SuccessMsg(dirList))
}

func GetTest(c *gin.Context) {
	c.JSON(http.StatusOK, response.SuccessMsg("successGet"))
}
