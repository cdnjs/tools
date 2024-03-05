package r2_pump

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/cdnjs/tools/gcp"
	"github.com/cdnjs/tools/sentry"
	"github.com/pkg/errors"

	"cloud.google.com/go/pubsub"
)

var (
	TOPIC   = os.Getenv("QUEUE")
	PROJECT = os.Getenv("PROJECT")
)

func Invoke(ctx context.Context, e gcp.GCSEvent) error {
	sentry.Init()
	defer sentry.PanicHandler()

	log.Printf("File: %v\n", e.Name)
	log.Printf("Metadata: %v\n", e.Metadata)

	if err := publish(e); err != nil {
		return fmt.Errorf("failed to publish: %v", err)
	}
	return nil
}

type Message struct {
	GCSEvent gcp.GCSEvent `json:"gcsEvent"`
}

func publish(e gcp.GCSEvent) error {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, PROJECT)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}
	t := client.Topic(TOPIC)

	msg := Message{e}
	bytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("could not marshal message: %s", err)
	}
	result := t.Publish(ctx, &pubsub.Message{Data: bytes})

	// The Get method blocks until a server-generated ID or
	// an error is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to publish")
	}
	log.Printf("Published msg ID: %v\n", id)
	return nil
}
