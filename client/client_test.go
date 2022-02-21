package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fsfc/config"
	"fsfc/response"
	"io/ioutil"
	"net/http"
	"testing"
)

func Test_Client(t *testing.T) {
	changedFiles := []string{"C:\\Users\\14595\\Desktop\\储存\\重要资料\\新建文本文档.txt"}

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
	fmt.Println(string(body))

	blockHashesReps := &response.BlockHashesReps{}
	err = json.Unmarshal(body, &blockHashesReps)
	if err != nil {
		return
	}

	//fileBlockHashes := blockHashesReps.Data

	fmt.Println(blockHashesReps)
}
