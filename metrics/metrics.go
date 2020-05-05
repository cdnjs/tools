package metrics

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cdnjs/tools/util"
)

const (
	BASE_URL           = "https://metrics.cdnjs.com"
	METRIC_NEW_VERSION = "new_version"
)

func reportMetric(metricType string) {
	token := util.GetEnv("METRICS_TOKEN")
	body := strings.NewReader("")
	_, err := http.Post(fmt.Sprintf("%s/%s?token=%s", BASE_URL, metricType, token), "text/plain", body)
	util.Check(err)
}

func ReportNewVersion() {
	reportMetric(METRIC_NEW_VERSION)
}
