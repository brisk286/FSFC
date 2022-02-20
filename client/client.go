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

	fileBlockHashes := blockHashesReps.Data

	for _, blockHashes := range fileBlockHashes {
		filename := blockHashes.Filename
		relaPath := fs.AbsToRela(filename)
		localPath := config.GetConfig().Set.LocalPath
		absPath := localPath + "\\" + relaPath

		modified, _ := ioutil.ReadFile(absPath)

		rsyncOps := rsync.CalculateDifferences(modified, blockHashes.BlockHashes)

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
