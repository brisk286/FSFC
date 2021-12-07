package main

import (
	"flag"
	"fmt"
	"os"
)

// 全局变量
var (
	fileShare string
)

func parseFlags() {
	flag.StringVar(&fileShare, "fileshare", "dir/", "共享的文件夹")
	flag.Parse()
}

func initDir() {
	if _, err := os.Stat(fileShare); os.IsNotExist(err) {
		err := os.Mkdir("./dir", os.ModePerm)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func main() {

	parseFlags()

	initDir()

}
