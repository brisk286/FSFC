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
)

func PostChangedFile() {
	primfs := fs.GetFs()
	changedFiles := primfs.GetChangedFile()

	fmt.Println(changedFiles)

	changedFilesJson, _ := json.Marshal(changedFiles)

	remoteIp := config.GetConfig().Web.RemoteIp
	remotePort := config.GetConfig().Web.RemotePort

	resp, err := http.Post("http://"+remoteIp+":"+remotePort+"/changedFile", "application/json", bytes.NewBuffer(changedFilesJson))
	if err != nil {
		fmt.Println(err)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)

	blockHashesReps := &response.BlockHashesReps{}
	err = json.Unmarshal(body, &blockHashesReps)
	if err != nil {
		return
	}

	fmt.Println(blockHashesReps)

	fileBlockHashes := blockHashesReps.Data

	for _, blockHashes := range fileBlockHashes {
		filename := blockHashes.Filename
		relaPath := fs.AbsToRela(filename)
		fmt.Println(relaPath)
		localPath := config.GetConfig().Set.RemotePath
		localPath = fs.FixDir(localPath)
		fmt.Println(localPath)
		absPath := localPath + relaPath

		modified, err := ioutil.ReadFile(absPath)
		if err != nil {
			fmt.Println(absPath)
			panic("读取本地文件失败")
		}
		fmt.Println("成功找到本地文件:")

		rsyncOps := rsync.CalculateDifferences(modified, blockHashes.BlockHashes)
		//fmt.Println(rsyncOps)
		fmt.Println("对比差异完成")

		rsyncOpsReq := request.RsyncOpsReq{filename, rsyncOps, len(modified)}
		PostRsyncOps(rsyncOpsReq)
	}

	defer resp.Body.Close()

}

func PostRsyncOps(rsyncOpsReq request.RsyncOpsReq) {
	rsyncOpsJson, _ := json.Marshal(rsyncOpsReq)

	remoteIp := config.GetConfig().Web.RemoteIp
	remotePort := config.GetConfig().Web.RemotePort

	resp, err := http.Post("http://"+remoteIp+":"+remotePort+"/rebuildFile", "application/json", bytes.NewBuffer(rsyncOpsJson))
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

}
