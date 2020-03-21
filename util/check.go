package util

import (
	"fmt"
)

func Check(err interface{}) {
	if err != nil {
		panic(err)
	}
}

func CheckCmd(out []byte, err error) string {
	if err != nil {
		fmt.Println(string(out))
		panic(err)
	}
	return string(out)
}
