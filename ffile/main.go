package main

import (
	"flag"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"time"
)

// 全局变量
var (
	localDirShare  string
	remoteDirShare string
)

func parseFlags() {
	flag.StringVar(&localDirShare, "localDirShare", "dir/", "本地共享的文件夹")
	flag.StringVar(&remoteDirShare, "remoteDirShare", "dir/", "远程共享的文件夹")
	flag.Parse()
}

func initLocalDir() {
	if _, err := os.Stat(localDirShare); os.IsNotExist(err) {
		err := os.Mkdir("./dir", os.ModePerm)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func initRemoteDir(sftpClient *sftp.Client) {
	if _, err := sftpClient.Stat(remoteDirShare); os.IsNotExist(err) {
		err := sftpClient.Mkdir("./dir")
		if err != nil {
			fmt.Println(err)
		}
	}
}

func main() {

	parseFlags()

	initLocalDir()

	sshConfig := &ssh.ClientConfig{
		User: "admin", // 账号
		Auth: []ssh.AuthMethod{
			ssh.Password("Qwebrisk286"), // 密码
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		ClientVersion:   "",
		Timeout:         10 * time.Second,
	}

	// 建立SSH连接
	sshClient, err := ssh.Dial("tcp", "152.136.187.78:22", sshConfig)
	if err != nil { // 如果有错
		log.Fatalln(err.Error())
	}
	defer func(sshClient *ssh.Client) {
		err := sshClient.Close()
		if err != nil {
			log.Fatalln(err.Error())
		}
	}(sshClient) //defer关闭连接

	// 通过SSH连接建立sftp连接
	sftpClient, err := sftp.NewClient(sshClient) //  *ssh.Client
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer func(sftpClient *sftp.Client) {
		err := sftpClient.Close()
		if err != nil {
			log.Fatalln(err.Error())
		}
	}(sftpClient)

	initRemoteDir(sftpClient)

}
