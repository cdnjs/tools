package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cdnjs/tools/cloudstorage"
	"github.com/cdnjs/tools/util"

	"cloud.google.com/go/storage"
)

type Logs = chan interface{}

// Wrapper around a log. Used to add the type
type LogEntry struct {
	Type   string      `json:"type"`
	At     string      `json:"at"`
	Detail interface{} `json:"detail"`
}

// The auto update process started
type LogAutoupdateStarted struct {
	Source string `json:"source"`
}

// A new version has been detected and can be downloaded, but that version
// already exists locally but hasn't been commited yet for some reason.
type LogNewVersionExistsLocally struct {
	Version string `json:"version"`
}

// A new version was downloaded but when trying to copy files in the final
// directory no files matched the pattern defined in the package's configuration
type LogNoFilesMatchedThePattern struct {
	Version string `json:"version"`
}

// A new version has been created locally
type LogCreatedNewVersion struct {
	Version string `json:"version"`
}

// Package has no existing version locally, we will import all the versions
type LogImportAllVersions struct {
	Versions []string `json:"versions"`
}

// New version was created, commit it
type LogNewVersionCommit struct {
	Version string `json:"version"`
}

// No new versions were detected. Package is up to date.
type LogNoNewVersion struct{}

func logToString(v interface{}) string {
	switch log := v.(type) {
	case LogAutoupdateStarted:
		return fmt.Sprintf("autoupdate using %s\n", log.Source)
	case LogNewVersionExistsLocally:
		return fmt.Sprintf("%s: exists; ignore\n", log.Version)
	case LogNoFilesMatchedThePattern:
		return fmt.Sprintf("%s: no files matched the pattern\n", log.Version)
	case LogCreatedNewVersion:
		return fmt.Sprintf("%s: created new version\n", log.Version)
	case LogImportAllVersions:
		return "no local data; import all versions"
	case LogNewVersionCommit:
		return fmt.Sprintf("commit version %s\n", log.Version)
	case LogNoNewVersion:
		return "no new version available"
	default:
		panic("unknown log entry")
	}
}

func logToType(v interface{}) string {
	switch v.(type) {
	case LogAutoupdateStarted:
		return "LogAutoupdateStarted"
	case LogNewVersionExistsLocally:
		return "LogNewVersionExistsLocally"
	case LogNoFilesMatchedThePattern:
		return "LogNoFilesMatchedThePattern"
	case LogCreatedNewVersion:
		return "LogCreatedNewVersion"
	case LogImportAllVersions:
		return "LogImportAllVersions"
	case LogNewVersionCommit:
		return "LogNewVersionCommit"
	case LogNoNewVersion:
		return "LogNoNewVersion"
	default:
		panic("unknown log entry")
	}
}

func log(ctx context.Context, entry interface{}) {
	// Print log to output
	util.Printf(ctx, logToString(entry))

	// send logs in the log sink
	logSink := ctx.Value("logs").(Logs)

	select {
	case logSink <- entry:
	default:
		util.Printf(ctx, "could not send message in logging sink")
	}
}

func publishAutoupdateLog(ctx context.Context, name string) {
	logSink := ctx.Value("logs").(Logs)
	close(logSink)

	at := time.Now().Format(time.RFC3339)

	content := ""
	for entry := range logSink {
		str, err := json.Marshal(LogEntry{
			At:     at,
			Type:   logToType(entry),
			Detail: entry,
		})
		util.Check(err)
		content += string(str) + "\n"
	}

	bkt, err := cloudstorage.GetRobotcdnjsBucket(ctx)
	util.Check(err)

	obj := bkt.Object(fmt.Sprintf("autoupdate/%s", name))

	w := obj.NewWriter(ctx)
	_, err = io.Copy(w, strings.NewReader(content))
	util.Check(err)
	util.Check(w.Close())
	util.Check(obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader))
}
