package client

import (
	"fmt"
	"fsfc/fs"
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
	resp, err := http.PostForm("http://localhost：8080/login.do", postParam)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

}
