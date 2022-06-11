package v1

import (
	"encoding/json"
	"fmt"
	"fsfc/config"
	"fsfc/fs"
	"fsfc/pkg/response"
	"io/ioutil"
	"net/http"
	"testing"
)

func Test_file(t *testing.T) {
	Fs := fs.GetFs()
	dirList, _ := Fs.Walk()

	for _, file := range dirList {
		fmt.Println(file)
	}
	//fmt.Println("1")
}

func Test_ar(t *testing.T) {
	arr := []string{"1", "23", "34", "23"}

	fmt.Println(arr)
}

func Test_GetLastSyncTime(t *testing.T) {
	localIp := config.GetConfig().Web.LocalIp
	localPort := config.GetConfig().Web.LocalPort

	resp, err := http.Get("http://" + localIp + ":" + localPort + "/v1/getLastSyncTime")
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
	fmt.Printf("%v\n", successCodeMsg)
}
