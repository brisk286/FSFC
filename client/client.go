package client

import (
	"encoding/json"
	"fmt"
	"fsfc/config"
	"fsfc/fs"
	"fsfc/response"
	"fsfc/rsync"
	"io/ioutil"
	"net/http"
	"net/url"
)

func postChangedFile() {
	primfs := fs.GetFs()
	changedFiles := primfs.GetChangedFile()

	postParam := url.Values{
		"changedFiles": changedFiles,
	}

	// 数据的键值会经过URL编码后作为请求的body传递
	//todo:设置接收端接口

	remoteIp := config.GetConfig().Web.RemoteIp
	remotePort := config.GetConfig().Web.RemotePort

	resp, err := http.PostForm("http://"+remoteIp+":"+remotePort+"/changedFile", postParam)
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

		rsync.CalculateDifferences(, blockHashes.BlockHashes)
	}

	defer resp.Body.Close()

}
