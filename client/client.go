package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fsfc/config"
	"fsfc/fs"
	"fsfc/request"
	"fsfc/response"
	"fsfc/rsync"
	"io/ioutil"
	"net/http"
	"time"
)

// PostChangedFile 扫描修改的文件夹
func PostChangedFile() {
	primfs := fs.GetFs()
	changedFiles := primfs.GetChangedFile()

	if changedFiles == nil {
		//fmt.Println(time.Now(), "未检测到文件修改")
		return
	}
	fmt.Println(time.Now(), "检测到修改的文件：", changedFiles)

	changedFilesJson, _ := json.Marshal(changedFiles)

	remoteIp := config.GetConfig().Web.RemoteIp
	remotePort := config.GetConfig().Web.RemotePort

	resp, err := http.Post("http://"+remoteIp+":"+remotePort+"/v1/changedFile",
		"application/json", bytes.NewBuffer(changedFilesJson))
	if err != nil {
		fmt.Println(err)
		fmt.Println("changedFiles发送失败，请检查网络连接，及存储端后台是否开启")
		fmt.Println("本次结果将存储到缓存中，将在通信成功后再次发送")

		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	blockHashesReps := response.BlockHashesReps{}
	err = json.Unmarshal(body, &blockHashesReps)
	if err != nil {
		fmt.Println(err)
		return
	}

	fileBlockHashes := blockHashesReps.Data
	for _, blockHashes := range fileBlockHashes {
		filename := blockHashes.Filename
		relaPath := fs.AbsToRela(filename)
		//localPath := fs.FixDir(config.GetConfig().Set.RemotePath)
		localPath := fs.FixDir(config.GetConfig().Set.LocalPath)
		absPath := localPath + relaPath

		modified, err := ioutil.ReadFile(absPath)
		if err != nil {
			fmt.Println(absPath)
			panic("读取本地文件失败")
		}
		fmt.Println("成功找到本地文件:", absPath)

		rsyncOps := rsync.CalculateDifferences(modified, blockHashes.BlockHashes)
		//fmt.Println(rsyncOps)
		fmt.Println("对比差异完成, 发送RsyncOps")

		rsyncOpsReq := request.RsyncOpsReq{
			Filename:       filename,
			RsyncOps:       rsyncOps,
			ModifiedLength: len(modified),
		}

		PostRsyncOps(rsyncOpsReq)
	}

	defer resp.Body.Close()

}

func PostRsyncOps(rsyncOpsReq request.RsyncOpsReq) {
	rsyncOpsJson, _ := json.Marshal(rsyncOpsReq)

	remoteIp := config.GetConfig().Web.RemoteIp
	remotePort := config.GetConfig().Web.RemotePort

	resp, err := http.Post("http://"+remoteIp+":"+remotePort+"/v1/rebuildFile", "application/json", bytes.NewBuffer(rsyncOpsJson))
	if err != nil {
		fmt.Println(err)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)

	successCodeMsg := response.ResponseMsg{}
	err = json.Unmarshal(body, &successCodeMsg)
	if err != nil {
		fmt.Println(err)
		return
	}

	if successCodeMsg.Msg == "SUCCESS" {
		fmt.Println("文件同步成功")
	}

	defer resp.Body.Close()

}
