package fs

import (
	"fmt"
	"os"
	"testing"
)

func TestWalk(t *testing.T) {
	root, _ := os.Getwd()

	fmt.Println("_____", root)

	//fs := &Filesystem{root: root}
	//
	//files, _ := fs.Walk()
	Fs := GetFs()

	dirList, _ := Fs.Walk()

	for i, file := range dirList {
		fmt.Println(i, file)
	}
}
