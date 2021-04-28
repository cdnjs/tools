package git_pump

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/cdnjs/tools/gcp"
	"github.com/cdnjs/tools/sentry"

	"github.com/pkg/errors"
)

const (
	CDNJS_GIT  = "https://github.com/cdnjs/cdnjs.git"
	CDNJS_PATH = "/tmp/cdnjs"
)

func Invoke(ctx context.Context, e gcp.GCSEvent) error {
	sentry.Init()
	defer sentry.PanicHandler()

	if err := gitClone(); err != nil {
		return fmt.Errorf("could not git clone: %s", err)
	}

	log.Println("OK")
	return nil
}

func gitClone() error {
	dir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "failed to get cwd")
	}

	cmd := exec.Command(
		dir+"/git", "--depth", "1",
		"--filter=blob:none", "--sparse",
		CDNJS_GIT, CDNJS_PATH)
	var out bytes.Buffer
	cmd.Stdout = &out
	log.Println(cmd)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "failed to clone cdnjs")
	}
	fmt.Printf("out: %q\n", out.String())
	return nil
}
