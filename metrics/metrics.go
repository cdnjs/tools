package metrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

const (
	METRICS_ENDPOINT = "https://metrics-worker.cloudflare-cdnjs.workers.dev"
)

var (
	METRICS_TOKEN = os.Getenv("METRICS_TOKEN")
)

type IncMetricPayload struct {
	Name   string                 `json:"name"`
	Labels IncMetricPayloadLabels `json:"labels"`
}

type IncMetricPayloadLabels struct {
	Type *string `json:"type"`
}

func NewUpdateDetected() error {
	return sendMetrics(&IncMetricPayload{
		Name:   "new_update_detected",
		Labels: IncMetricPayloadLabels{},
	})
}

func NewUpdateProccessed() error {
	return sendMetrics(&IncMetricPayload{
		Name:   "new_update_processed",
		Labels: IncMetricPayloadLabels{},
	})
}

func NewUpdatePublishedKV() error {
	return sendMetrics(&IncMetricPayload{
		Name:   "new_update_published_kv",
		Labels: IncMetricPayloadLabels{},
	})
}

func NewUpdatePublishedR2() error {
	t := "instance"

	return sendMetrics(&IncMetricPayload{
		Name: "new_update_published_r2",
		Labels: IncMetricPayloadLabels{
			Type: &t,
		},
	})
}

func NewUpdatePublishedAlgolia() error {
	return sendMetrics(&IncMetricPayload{
		Name:   "new_update_published_algolia",
		Labels: IncMetricPayloadLabels{},
	})
}

func sendMetrics(payload *IncMetricPayload) error {
	json, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "failed to marshall payload")
	}

	req, err := http.NewRequest("POST", METRICS_ENDPOINT, bytes.NewBuffer(json))
	if err != nil {
		return errors.Wrap(err, "failed to build request")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", METRICS_TOKEN))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return errors.Errorf("metrics endpoint returned %s", resp.Status)
	}
	return nil
}
