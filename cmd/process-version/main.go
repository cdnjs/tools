package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"
)

var (
	PROJECT      = os.Getenv("PROJECT")
	SUBSCRIPTION = os.Getenv("SUBSCRIPTION")
)

func main() {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, PROJECT)
	if err != nil {
		log.Fatalf("could not create pubsub Client: %v", err)
	}
	sub := client.Subscription(SUBSCRIPTION)

	for {
		if err := consume(client, sub); err != nil {
			log.Fatalf("could not pull messages: %s", err)
		}
	}
}

func consume(client *pubsub.Client, sub *pubsub.Subscription) error {
	ctx := context.Background()
	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		msg.Ack()
		fmt.Printf("Got message: %q\n", string(msg.Data))
	})
	if err != nil {
		return errors.Wrap(err, "could not receive from subscription")
	}
	return nil
}
