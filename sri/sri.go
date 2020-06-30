package sri

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/cdnjs/tools/util"
)

// CalculateFileSRI generates a Subresource Integrity string for a particular file.
func CalculateFileSRI(filename string) string {
	f, err := os.Open(filename)
	util.Check(err)
	defer f.Close()

	h := sha512.New()
	_, err = io.Copy(h, f)
	util.Check(err)

	sri := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("sha512-%s", sri)
}
