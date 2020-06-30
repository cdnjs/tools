package util

import (
	"fmt"
)

// Check enforces that its argument is nil, panicking otherwise.
func Check(err interface{}) {
	if err != nil {
		panic(err)
	}
}

// CheckCmd enforces that an error is nil, printing
// the output and panicking otherwise.
func CheckCmd(out []byte, err error) string {
	if err != nil {
		fmt.Println(string(out))
		panic(err)
	}
	return string(out)
}
