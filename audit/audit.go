package audit

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

const (
	GH_OWNER = "cdnjs"
	GH_REPO  = "logs"
	GH_NAME  = "robocdnjs"
	GH_EMAIL = "cdnjs-github@cloudflare.com"
)

var (
	GH_TOKEN  = os.Getenv("AUDIT_GH_TOKEN")
	GH_BRANCH = os.Getenv("AUDIT_GH_BRANCH")
)

func getPath(pkgName, version string) string {
	firstLetter := pkgName[0:1]
	return fmt.Sprintf("packages/%s/%s/%s.log", firstLetter, pkgName, version)
}

func getClient(ctx context.Context) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GH_TOKEN},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func get(ctx context.Context, pkgName string, version string) (*github.RepositoryContent, error) {
	client := getClient(ctx)
	file := getPath(pkgName, version)
	opts := github.RepositoryContentGetOptions{
		Ref: GH_BRANCH,
	}
	res, _, _, err := client.Repositories.GetContents(ctx, GH_OWNER, GH_REPO, file, &opts)
	if err != nil {
		return nil, errors.Wrap(err, "could not get file")
	}
	return res, nil
}

func update(ctx context.Context, pkgName string, version string,
	newContent *bytes.Buffer) error {
	message := fmt.Sprintf("Update %s %s", pkgName, version)
	file := getPath(pkgName, version)
	client := getClient(ctx)

	currContent, err := get(ctx, pkgName, version)
	if err != nil {
		return errors.Wrap(err, "failed to get current file")
	}

	var content bytes.Buffer
	content.WriteString(*currContent.Content)
	content.Write(newContent.Bytes())

	commitOption := &github.RepositoryContentFileOptions{
		Branch:  github.String(GH_BRANCH),
		Message: github.String(message),
		Committer: &github.CommitAuthor{
			Name:  github.String(GH_NAME),
			Email: github.String(GH_EMAIL),
		},
		Author: &github.CommitAuthor{
			Name:  github.String(GH_NAME),
			Email: github.String(GH_EMAIL),
		},
		Content: content.Bytes(),
		SHA:     currContent.SHA,
	}

	c, resp, err := client.Repositories.UpdateFile(ctx, GH_OWNER, GH_REPO, file, commitOption)
	if err != nil {
		return errors.Wrap(err, "could not update file")
	}
	log.Printf("audit updated: resp.Status=%v commit=%s", resp.Status, *c.SHA)
	return nil
}

func create(ctx context.Context, pkgName string, version string,
	content *bytes.Buffer) error {
	message := fmt.Sprintf("Create %s %s", pkgName, version)
	file := getPath(pkgName, version)

	client := getClient(ctx)

	commitOption := &github.RepositoryContentFileOptions{
		Branch:  github.String(GH_BRANCH),
		Message: github.String(message),
		Committer: &github.CommitAuthor{
			Name:  github.String(GH_NAME),
			Email: github.String(GH_EMAIL),
		},
		Author: &github.CommitAuthor{
			Name:  github.String(GH_NAME),
			Email: github.String(GH_EMAIL),
		},
		Content: content.Bytes(),
	}

	c, resp, err := client.Repositories.CreateFile(ctx, GH_OWNER, GH_REPO, file, commitOption)
	if err != nil {
		return errors.Wrap(err, "could not create file")
	}
	log.Printf("audit created: resp.Status=%v commit=%s", resp.Status, *c.SHA)
	return nil
}

func NewVersionDetected(ctx context.Context, pkgName string, version string) error {

	content := bytes.NewBufferString("")
	fmt.Fprintf(content, "New version: %s\n", version)

	if err := create(ctx, pkgName, version, content); err != nil {
		return errors.Wrap(err, "could not create audit log file")
	}
	return nil
}

func ProcessedVersion(ctx context.Context, pkgName string, version string, log string) error {
	content := bytes.NewBufferString("")
	fmt.Fprintf(content, "Processing log: %s\n", log)

	if err := update(ctx, pkgName, version, content); err != nil {
		return errors.Wrap(err, "could not create audit log file")
	}
	return nil
}

func WroteKV(ctx context.Context, pkgName string, version string,
	sris map[string]string, keys []string, config string) error {

	content := bytes.NewBufferString("")
	fmt.Fprintf(content, "config: %s\n", config)
	fmt.Fprint(content, "KV keys:\n")
	for _, key := range keys {
		fmt.Fprintf(content, "- %s\n", key)
	}
	fmt.Fprint(content, "SRIs:\n")
	for name, sri := range sris {
		fmt.Fprintf(content, "- %s: %s\n", name, sri)
	}

	if err := update(ctx, pkgName, version, content); err != nil {
		return errors.Wrap(err, "could not create audit log file")
	}
	return nil
}
