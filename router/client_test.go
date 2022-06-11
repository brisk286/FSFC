package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fsfc/config"
	"fsfc/pkg/response"
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

func Test_ConnectFail(t *testing.T) {

	//primfs := fs.GetFs()
	changedFiles := []string{"test"}

	changedFilesJson, _ := json.Marshal(changedFiles)

	remoteIp := "152.136.187.78"
	remotePort := config.GetConfig().Web.RemotePort

	_, err := http.Post("http://"+remoteIp+":"+remotePort+"/changedFile", "application/json", bytes.NewBuffer(changedFilesJson))
	if err != nil {
		fmt.Println(err)
		return
	}

	//body, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body))
}

//存储端后台未开启:
//Post "http://127.0.0.1:5555/changedFile": dial tcp 127.0.0.1:5555:
//connectex: No connection could be made because the target machine actively refused it.

//发送错误字段
//无反馈

//断网:
//Post "http://152.136.187.78:5555/changedFile": EOF
func Test_1(t *testing.T) {
	fmt.Println("1")
}
