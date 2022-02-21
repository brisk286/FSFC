package rsync

import (
	"fmt"
	"testing"
)

type data struct {
	a int
	b string
}

func Test_CalculateDifferences(t *testing.T) {
	var dataA []data
	dataA = append(dataA, data{a: 1, b: "aefa"})

	fmt.Println(dataA)
}
