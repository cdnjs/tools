package metrics

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cdnjs/tools/util"
)

const (
	baseURL          = "https://metrics.cdnjs.com"
	metricNewVersion = "new_version"
)

func reportMetric(metricType string) {
	token := util.GetEnv("METRICS_TOKEN")
	body := strings.NewReader("")
	_, err := http.Post(fmt.Sprintf("%s/%s?token=%s", baseURL, metricType, token), "text/plain", body)
	util.Check(err)
}

// ReportNewVersion reports a new version via http POST.
func ReportNewVersion() {
	reportMetric(metricNewVersion)
}
