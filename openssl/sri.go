package openssl

import (
	"fmt"
	"os/exec"

	"github.com/cdnjs/tools/util"
)

// CalculateFileSri generates a Subresource Integrity string for a particular file.
func CalculateFileSri(filename string) string {
	dgst, dgstErr := exec.Command("openssl", "dgst", "-sha256", "-binary", filename).Output()
	util.Check(dgstErr)

	encCmd := exec.Command("openssl", "enc", "-base64", "-A")
	encStdin, stdinErr := encCmd.StdinPipe()
	util.Check(stdinErr)

	// Feed digest into encoding
	_, writeErr := encStdin.Write(dgst)
	util.Check(writeErr)
	encStdin.Close()

	sri, outputErr := encCmd.Output()
	util.Check(outputErr)

	return fmt.Sprintf("sha256-%s", sri)
}
