package metrics

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/cdnjs/tools/util"
)

const (
	baseURL          = "https://metrics.cdnjs.com"
	metricNewVersion = "new_version"
)

func reportMetric(ctx context.Context, metricType string) {
	if token, ok := os.LookupEnv("METRICS_TOKEN"); ok {
		body := strings.NewReader("")
		_, err := http.Post(fmt.Sprintf("%s/%s?token=%s", baseURL, metricType, token), "text/plain", body)
		util.Check(err)
	} else {
		util.Debugf(ctx, "ignoring metric report (env missing)")
	}
}

// ReportNewVersion reports a new version via http POST.
func ReportNewVersion(ctx context.Context) {
	reportMetric(ctx, metricNewVersion)
}
