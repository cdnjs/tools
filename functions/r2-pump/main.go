package r2_pump

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cdnjs/tools/gcp"
	"github.com/cdnjs/tools/sentry"
	"github.com/pkg/errors"
)

var (
	R2_PUMP_ENDPOINT = os.Getenv("R2_PUMP_ENDPOINT")
)

// same as functions/r2-pump-http/main.go
type InvokePayload struct {
	Package string `json:"package"`
	Bucket  string `json:"bucket"`
	Version string `json:"version"`
	Config  string `json:"config"`
	Name    string `json:"name"`
}

// At the time of writing the timeout for function triggerd by an event is 9min and
// HTTP triggered function can go up to 1 hour.
// Some packages hit the timeout, to bypass the limiation this function acts a
// proxy. It's called by an event trigger and calls another function via HTTP
// to do the actual uploading to R2.
func Invoke(ctx context.Context, e gcp.GCSEvent) error {
	sentry.Init()
	defer sentry.PanicHandler()

	data := InvokePayload{
		Package: e.Metadata["package"].(string),
		Version: e.Metadata["version"].(string),
		Config:  e.Metadata["config"].(string),
		Bucket:  e.Bucket,
		Name:    e.Name,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "failed to marshall data")
	}

	req, err := http.NewRequest("POST", R2_PUMP_ENDPOINT, bytes.NewBuffer(body))
	if err != nil {
		return errors.Wrap(err, "failed create new request")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s returned %d: %s", R2_PUMP_ENDPOINT, resp.StatusCode, body)
	}

	return nil
}
