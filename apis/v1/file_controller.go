package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fsfc/config"
	"fsfc/fs"
	"fsfc/logger"
	"fsfc/pkg"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
)

// GetFiles 获取root路径下所有文件的绝对路径
func GetFiles(c *gin.Context) {
	dirList, err := fs.PrimFs.Walk()
	if err != nil {
		logger.Logger.Error("rootWrong", logger.Any("LocalPath", fs.PrimFs.LocalPath))
		c.JSON(http.StatusOK, pkg.FailMsg(err.Error()))
	}
	c.JSON(http.StatusOK, pkg.SuccessMsg(dirList))
}

// AddSaveSpace 前端添加成功后，请求此接口，更新同步文件夹，再更新存储端的信息
func AddSaveSpace(c *gin.Context) {
	addSaveSpaceReq := new(pkg.AddSaveSpaceReq)
	err := c.ShouldBindBodyWith(&addSaveSpaceReq, binding.JSON)
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	// 更新本地
	for _, localPath := range addSaveSpaceReq.Dirs {
		fs.PrimFs.LocalPath = append(fs.PrimFs.LocalPath, localPath)
	}

	// 更新存储端
	bytesData, _ := json.Marshal(addSaveSpaceReq)
	url := "http://" + config.Config.Web.RemoteIp + ":" + config.Config.Web.RemotePort + "/addSaveSpace"
	_, err = http.Post(url,
		"application/json;charset=utf-8", bytes.NewBuffer(bytesData))
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
	}

	fmt.Println("监控文件夹更新")
	c.JSON(http.StatusOK, pkg.SuccessCodeMsg())
}
