package fssh

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"log"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	sshConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password("Qwebrisk286"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		ClientVersion:   "",
		Timeout:         10 * time.Second,
	}

	//建立与SSH服务器的连接
	sshClient, err := ssh.Dial("tcp", "152.136.187.78:22", sshConfig)
	if err != nil { // 如果有错
		log.Fatalln(err.Error())
	}
	defer sshClient.Close() //defer关闭连接

	sftpClient, err := sftp.NewClient(sshClient) //  *ssh.Client
	if err != nil {
		log.Fatalln(err.Error())
	}

	defer sftpClient.Close()

	//获取当前目录
	cwd, err := sftpClient.Getwd()
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Println("当前目录：", cwd)

	//显示文件/目录详情
	fi, err := sftpClient.Lstat(cwd)
	log.Println(fi)
}

func TestInt(t *testing.T) {
	fmt.Println(runtime.GOARCH)
	fmt.Println(runtime.GOOS)
	fmt.Println(strconv.IntSize)
}
