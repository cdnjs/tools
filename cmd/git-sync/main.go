package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Item struct {
	MediaLink   string       `json:"mediaLink"`
	Name        string       `json:"name"`
	TimeCreated string       `json:"timeCreated"`
	Metadata    ItemMetadata `json:"metadata"`
}

func (i Item) Time() time.Time {
	t, err := time.Parse(time.RFC3339Nano, i.TimeCreated)
	if err != nil {
		panic("could not parse date")
	}
	return t
}

type ItemMetadata struct {
	Pkg     string `json:"package"`
	Version string `json:"version"`
}

type List struct {
	NextPageToken *string `json:"nextPageToken"`
	Items         []Item  `json:"items"`
}

func readLastSync(path string) (time.Time, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "could not read file")
	}
	return time.Parse(time.RFC3339Nano, strings.TrimSpace(string(data)))
}

func updateLastSync(path string, t time.Time) error {
	log.Printf("advance last sync to %s\n", t)
	data := t.Format(time.RFC3339Nano)
	err := ioutil.WriteFile(path, []byte(data), 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write file")
	}

	if err := git("add", "--sparse", path); err != nil {
		return errors.Wrap(err, "failed to run git")
	}

	if err := git("commit", "-m", "update last sync marker"); err != nil {
		return errors.Wrap(err, "failed to run git")
	}
	return nil
}

var (
	DEBUG               = os.Getenv("DEBUG") == "1"
	MAX_UPDATES_PER_RUN = 100
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("last sync file and bucket missing")
	}
	lastSync, err := readLastSync(os.Args[1])
	if err != nil {
		log.Fatalf("failed to get last sync: %s", err)
	}

	bucket := os.Args[2]
	items, err := getItems(bucket)
	if err != nil {
		log.Fatalf("failed to get items: %s", err)
	}

	newVersions, err := diff(lastSync, items)
	if err != nil {
		log.Fatalf("failed to detect new versions: %s", err)
	}

	log.Printf("%d updates since %s\n", len(newVersions), lastSync)
	if DEBUG {
		for _, version := range newVersions {
			log.Printf("%s new version %s\n", version.TimeCreated, version.Name)
		}
	}

	if len(newVersions) > MAX_UPDATES_PER_RUN {
		newVersions = newVersions[:MAX_UPDATES_PER_RUN]
		log.Printf("too many updates; limiting to %d updates\n", MAX_UPDATES_PER_RUN)
	}

	// Keep track of the last successful version we addedd
	lastSuccessfullSync := lastSync

	for _, version := range newVersions {
		t, err := addNewVersion(version)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to add new version: %s\n", err)
		} else {
			lastSuccessfullSync = *t
		}
	}

	if lastSuccessfullSync != lastSync {
		if err := updateLastSync(os.Args[1], lastSuccessfullSync); err != nil {
			log.Fatalf("could not update last sync: %s", err)
		}
	}
}

func writeFile(target string, r io.Reader) error {
	outFile, err := os.Create(target)
	if err != nil {
		return errors.Wrap(err, "failed to create file")
	}
	if _, err := io.Copy(outFile, r); err != nil {
		return errors.Wrap(err, "failed to write file")
	}
	outFile.Close()
	return nil
}

func dirExists(path string) bool {
	if DEBUG {
		return false
	}
	cmd := exec.Command("git", "ls-tree", "-d", "origin/master:"+path)
	if err := cmd.Run(); err != nil {
		return false
	} else {
		return true
	}
}

func addNewVersion(item Item) (*time.Time, error) {
	log.Printf("add new version %s %s", item.Metadata.Pkg, item.Metadata.Version)

	tar, err := download(item)
	if err != nil {
		return nil, errors.Wrap(err, "could not download object")
	}
	defer tar.Close()

	dest := fmt.Sprintf("ajax/libs/%s/%s", item.Metadata.Pkg, item.Metadata.Version)
	if dirExists(dest) {
		log.Printf("version %s already exists, ignoring\n", dest)

		// Version already exists for some reason, don't make any changes.
		t := item.Time()
		return &t, nil
	}

	if err := os.MkdirAll(dest, 0755); err != nil {
		return nil, errors.Wrap(err, "failed to create version directory")
	}

	hasFiles := false
	onFile := func(name string, r io.Reader) error {
		ext := filepath.Ext(name)
		if ext == ".woff2" {
			// woff2 files are not compressed, write as is
			target := path.Join(dest, name)
			if err := os.MkdirAll(path.Dir(target), 0755); err != nil {
				return errors.Wrap(err, "failed to create directory")
			}
			if err := writeFile(target, r); err != nil {
				return errors.Wrap(err, "failed to write file")
			}
		}
		if ext == ".gz" {
			name = strings.ReplaceAll(name, ".gz", "")
			target := path.Join(dest, name)
			if err := os.MkdirAll(path.Dir(target), 0755); err != nil {
				return errors.Wrap(err, "failed to create directory")
			}
			uncompressed, err := gunzip(r)
			if err != nil {
				return errors.Wrap(err, "failed to uncompress")
			}
			if err := writeFile(target, bytes.NewReader(uncompressed)); err != nil {
				return errors.Wrap(err, "failed to write file")
			}
		}

		hasFiles = true
		return nil
	}
	if err := inflate(tar, onFile); err != nil {
		return nil, errors.Wrap(err, "failed to extract files")
	}

	if hasFiles {
		if err := git("add", "--sparse", dest); err != nil {
			return nil, errors.Wrap(err, "failed to run git")
		}

		commitMsg := fmt.Sprintf("Add %s (%s)", item.Metadata.Pkg, item.Metadata.Version)
		if err := git("commit", "-m", commitMsg); err != nil {
			return nil, errors.Wrap(err, "failed to run git")
		}
	} else {
		log.Printf("no files present")
	}

	t := item.Time()
	return &t, nil
}

func git(args ...string) error {
	cmd := exec.Command("git", args...)
	log.Printf("running: %s", cmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if !DEBUG {
		if err := cmd.Run(); err != nil {
			return errors.Wrap(err, "failed to run command")
		}
	}
	return nil
}

func download(item Item) (io.ReadCloser, error) {
	resp, err := http.Get(item.MediaLink)
	if err != nil {
		return nil, errors.Wrap(err, "could not get object")
	}

	if resp.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "could not read response")
		}
		return nil, errors.Errorf("returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp.Body, nil
}

func getItems(bucket string) ([]Item, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	items := make([]Item, 0)

	nextPageToken := ""
	for {
		r, err := client.Get(fmt.Sprintf("https://www.googleapis.com/storage/v1/b/%s/o?pageToken=%s", bucket, nextPageToken))
		if err != nil {
			return nil, errors.Wrap(err, "could not get listing")
		}
		defer r.Body.Close()

		target := new(List)
		if err := json.NewDecoder(r.Body).Decode(target); err != nil {
			return nil, errors.Wrap(err, "could not decode response")
		}

		items = append(items, target.Items...)
		if target.NextPageToken == nil {
			break
		}
		nextPageToken = *target.NextPageToken
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Time().Before(items[j].Time())
	})
	return items, nil
}

func diff(lastSync time.Time, items []Item) ([]Item, error) {
	changes := make([]Item, 0)

	for _, item := range items {
		t := item.Time()
		if !t.Equal(lastSync) && t.After(lastSync) {
			changes = append(changes, item)
		}
	}
	return changes, nil
}

func inflate(gzipStream io.Reader, onFile func(string, io.Reader) error) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return errors.Wrap(err, "ExtractTarGz: NewReader failed")
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return errors.Wrap(err, "ExtractTarGz: Next() faileds")
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// ignore
		case tar.TypeReg:
			if err := onFile(header.Name, tarReader); err != nil {
				return errors.Wrap(err, "failed to handle file")
			}
		default:
			return errors.Errorf(
				"ExtractTarGz: uknown type: %x in %s",
				header.Typeflag,
				header.Name)
		}

	}
	return nil
}

func gunzip(in io.Reader) ([]byte, error) {
	r, err := gzip.NewReader(in)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gzip reader")
	}

	var out bytes.Buffer
	_, err = out.ReadFrom(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read from gzip")
	}

	return out.Bytes(), nil
}
