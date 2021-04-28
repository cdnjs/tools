package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cdnjs/tools/audit"
	"github.com/cdnjs/tools/sentry"

	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"
)

var (
	PROJECT      = os.Getenv("PROJECT")
	SUBSCRIPTION = os.Getenv("SUBSCRIPTION")
	DOCKER_IMAGE = os.Getenv("DOCKER_IMAGE")
)

func init() {
	sentry.Init()
}

func main() {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, PROJECT)
	if err != nil {
		log.Fatalf("could not create pubsub Client: %v", err)
	}
	sub := client.Subscription(SUBSCRIPTION)
	sub.ReceiveSettings.Synchronous = true
	sub.ReceiveSettings.MaxOutstandingMessages = 5
	sub.ReceiveSettings.NumGoroutines = runtime.NumCPU()

	for {
		if err := consume(client, sub); err != nil {
			log.Fatalf("could not pull messages: %s", err)
		}
	}
}

type Message struct {
	OutgoingSignedURL string           `json:"outgoingSignedURL"`
	Tar               string           `json:"tar"`
	Pkg               string           `json:"package"`
	Version           string           `json:"version"`
	Config            *json.RawMessage `json:"config"`
}

func consume(client *pubsub.Client, sub *pubsub.Subscription) error {
	ctx := context.Background()
	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		log.Printf("received message: %s\n", msg.Data)

		if err := processMessage(ctx, msg.Data); err != nil {
			log.Fatalf("failed to process message: %s", err)
		}
		msg.Ack()
	})
	if err != nil {
		return errors.Wrap(err, "could not receive from subscription")
	}
	return nil
}

func processMessage(ctx context.Context, data []byte) error {
	var message Message
	if err := json.Unmarshal(data, &message); err != nil {
		return errors.Wrap(err, "failed to parse")
	}

	inDir, outDir, err := setupSandbox()
	if err != nil {
		return errors.Wrap(err, "failed to setup sandbox")
	}
	defer os.RemoveAll(inDir)
	defer os.RemoveAll(outDir)

	if err := writeConfig(inDir, message.Config); err != nil {
		return errors.Wrap(err, "failed to write configuration")
	}
	if err := download(inDir, message.Tar); err != nil {
		return errors.Wrapf(err, "failed to download: %s", message.Tar)
	}

	logs, err := runSandbox(inDir, outDir)
	if err != nil {
		return errors.Wrap(err, "failed to run sandbox")
	}
	log.Println("logs", logs)

	if err := audit.ProcessedVersion(ctx, message.Pkg, message.Version, logs); err != nil {
		log.Printf("could not audit: %s", err)
	}

	var buff bytes.Buffer
	if err := compress(outDir, &buff); err != nil {
		return errors.Wrap(err, "failed to compress out dir")
	}

	if err := uploadToOutgoing(buff, message); err != nil {
		return errors.Wrap(err, "failed to upload to outgoing bucket")
	}

	return nil
}

func uploadToOutgoing(content bytes.Buffer, msg Message) error {
	r := bytes.NewReader(content.Bytes())
	req, err := http.NewRequest("PUT", msg.OutgoingSignedURL, r)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	encodedConfig := b64.StdEncoding.EncodeToString(*msg.Config)

	req.Header.Set("x-goog-meta-package", msg.Pkg)
	req.Header.Set("x-goog-meta-version", msg.Version)
	req.Header.Set("x-goog-meta-config", encodedConfig)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "request failed")
	}

	if res.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return errors.Wrap(err, "could not read response body")
		}
		log.Printf("returned %s: %s\n", res.Status, string(bodyBytes))
	} else {
		log.Println("OK")
	}
	return nil
}

func writeConfig(dstDir string, config *json.RawMessage) error {
	if err := ioutil.WriteFile(path.Join(dstDir, "config.json"), *config, 0644); err != nil {
		return errors.Wrap(err, "could not write config file")
	}
	return nil
}

func download(dstDir string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dst, err := os.Create(path.Join(dstDir, "new-version.tgz"))
	if err != nil {
		return errors.Wrap(err, "could not write tmp file")
	}
	defer dst.Close()

	_, err = io.Copy(dst, resp.Body)
	return err
}

func compress(src string, buf io.Writer) error {
	// tar > gzip > buf
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	// walk through every file in the folder
	err := filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		// remove the /tmp/out** prefix
		relFile := strings.ReplaceAll(file, src, "")

		// must provide real name
		// (see https://golang.org/src/archive/tar/common.go?#L626)
		header.Name = filepath.ToSlash(relFile)

		// write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// if not a dir, write file content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// produce tar
	if err := tw.Close(); err != nil {
		return err
	}
	// produce gzip
	if err := zr.Close(); err != nil {
		return err
	}
	//
	return nil
}
