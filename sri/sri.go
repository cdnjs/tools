package sri

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"github.com/cdnjs/tools/util"
)

// CalculateFileSRI generates a Subresource Integrity string for a particular file.
func CalculateFileSRI(filepath string, out string) {
	bytes, err := ioutil.ReadFile(filepath)
	util.Check(err)

	res := CalculateSRI(bytes)

	err = ioutil.WriteFile(out, []byte(res), 0644)
	util.Check(err)
}

// CalculateSRI calculates a Subresource Integrity string from bytes.
func CalculateSRI(bytes []byte) string {
	h := sha512.New()
	_, err := h.Write(bytes)
	util.Check(err)

	sri := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("sha512-%s", sri)
}
