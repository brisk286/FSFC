package fs

import (
	"fmt"
	"os"
	"testing"
)

func TestWalk(t *testing.T) {
	root, _ := os.Getwd()

	fmt.Println("_____", root)

	fs := &filesystem{root: root}

	files, _ := fs.walk()

	for i, file := range files {
		fmt.Println(i, file)
	}
}
