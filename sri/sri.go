package sri

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"

	"github.com/cdnjs/tools/util"
)

// CalculateFileSRI generates a Subresource Integrity string for a particular file.
func CalculateFileSRI(filepath string) string {
	bytes, err := util.ReadLibFileSafely(filepath)
	util.Check(err)

	return CalculateSRI(bytes)
}

// CalculateSRI calculates a Subresource Integrity string from bytes.
func CalculateSRI(bytes []byte) string {
	h := sha512.New()
	_, err := h.Write(bytes)
	util.Check(err)

	sri := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("sha512-%s", sri)
}
