package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cdnjs/tools/algolia"
	"github.com/cdnjs/tools/cloudstorage"
	"github.com/cdnjs/tools/github"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sentry"
	"github.com/cdnjs/tools/util"

	algoliasearch "github.com/algolia/algoliasearch-client-go/v3/algolia/search"
)

func init() {
	sentry.Init()
}

// PackagesJSON is used to wrap around a slice of []Packages
// when JSON unmarshalling.
type PackagesJSON struct {
	Packages []Package `json:"packages"`
}

// Package contains metadata for a particular package.
// FIXME(sven): remove parsing here in favor of github.com/cdnjs/tools/packages
type Package struct {
	Name        string   `json:"name"`
	Filename    string   `json:"filename"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Keywords    []string `json:"keywords"`
	// TODO: handle the case where multiple licenses are specified in the
	// `licenses` key
	License    interface{} `json:"license,omitempty"`
	Homepage   string      `json:"homepage"`
	Author     interface{} `json:"author,omitempty"`
	Repository interface{} `json:"repository,omitempty"`
	Namespace  interface{} `json:"namespace,omitempty"`
}

// SearchEntry represents an entry in the Algolia Search index.
type SearchEntry struct {
	Name             string               `json:"name"`
	Filename         string               `json:"filename"`
	Description      string               `json:"description"`
	Version          string               `json:"version"`
	Keywords         []string             `json:"keywords"`
	AlternativeNames []string             `json:"alternativeNames"`
	FileType         string               `json:"fileType"`
	Github           *GitHubMeta          `json:"github"`
	ObjectID         string               `json:"objectID"`
	License          string               `json:"license"`
	Homepage         string               `json:"homepage"`
	Namespace        string               `json:"namespace,omitempty"`
	Repository       *packages.Repository `json:"repository"`
	Author           string               `json:"author"`
	OriginalName     string               `json:"originalName"`
	Sri              string               `json:"sri"`
}

// GitHubMeta contains metadata for a particular GitHub repository.
type GitHubMeta struct {
	User             string `json:"user"`
	Repo             string `json:"repo"`
	StargazersCount  int    `json:"stargazers_count"`
	Forks            int    `json:"forks"`
	SubscribersCount int    `json:"subscribers_count"`
}

func getPackagesBuffer() bytes.Buffer {
	ctx := context.Background()

	bkt, err := cloudstorage.GetAssetsBucket(ctx)
	util.Check(err)

	obj := bkt.Object("package.min.js")

	r, err := obj.NewReader(ctx)
	util.Check(err)
	defer r.Close()

	var b bytes.Buffer

	_, copyerr := io.Copy(bufio.NewWriter(&b), r)
	util.Check(copyerr)

	return b
}

var (
	re1 = regexp.MustCompile(`[^a-zA-Z]`)
	re2 = regexp.MustCompile(`(^[^A-Z]*|[A-Z]*)([A-Z][^A-Z]+|$)`)
)

func getAlternativeNames(name string) []string {
	names := make([]string, 0)
	names = append(names, re1.ReplaceAllString(name, `$1 $2`))
	names = append(names, re2.ReplaceAllString(name, `$1 $2`))
	return names
}

func parseLicense(p *Package) (string, error) {
	license := ""
	switch v := p.License.(type) {
	case string:
		{
			license = v
		}
	case map[string]interface{}:
		{
			if v["name"] != nil {
				license = v["name"].(string)
			}
		}
	case nil:
	default:
		{
			return "", fmt.Errorf("unsupported license value: `%s`", p.License)
		}
	}

	return license, nil
}

func parseAuthor(p *Package) (string, error) {
	author := ""
	switch v := p.Author.(type) {
	case string:
		{
			author = v
		}
	case map[string]interface{}:
		{
			if v["name"] != nil {
				author = v["name"].(string)
			}
		}
	case nil:
	default:
		{
			return "", fmt.Errorf("unsupported author value: `%s`", p.Author)
		}
	}

	return author, nil
}

func parseRepository(p *Package) (*packages.Repository, error) {
	switch v := p.Repository.(type) {
	case string:
		{
			return &packages.Repository{
				Repotype: "",
				URL:      v,
			}, nil
		}
	case map[string]interface{}:
		{
			if v["type"] != nil && v["url"] != nil {
				return &packages.Repository{
					Repotype: v["type"].(string),
					URL:      v["url"].(string),
				}, nil
			}
			return nil, nil
		}
	case nil:
		{
			return nil, nil
		}
	default:
		{
			return nil, fmt.Errorf("unsupported Repository value: `%s`", p.Repository)
		}
	}
}

var githubURL = regexp.MustCompile(`github\.com[/|:]([\w\.-]+)\/([\w\.-]+)\/?`)

func getGitHubMeta(repo *packages.Repository) (*GitHubMeta, error) {
	if repo == nil {
		// no repo configured
		return nil, nil
	}
	if repo.Repotype != "git" {
		return nil, fmt.Errorf("unsupported repo type `%s`", repo.Repotype)
	}

	res := githubURL.FindAllStringSubmatch(repo.URL, -1)
	if len(res) == 0 {
		return nil, fmt.Errorf("could not parse repo URL `%s`", repo.URL)
	}

	client := github.GetClient()
	api, _, err := client.Repositories.Get(util.ContextWithEntries(), res[0][1], strings.ReplaceAll(res[0][2], ".git", ""))
	if err != nil {
		return nil, err
	}

	return &GitHubMeta{
		User:             api.GetOwner().GetLogin(),
		Repo:             api.GetName(),
		StargazersCount:  api.GetStargazersCount(),
		Forks:            api.GetForksCount(),
		SubscribersCount: api.GetSubscribersCount(),
	}, nil
}

func getSRI(p *Package) (string, error) {
	jsonFile := path.Join(util.SRIPath, p.Name, p.Version+".json")

	var j map[string]interface{}

	data, err := ioutil.ReadFile(jsonFile)

	if err != nil {
		return "", nil
	}

	util.Check(json.Unmarshal(data, &j))

	if str, ok := j[p.Filename].(string); ok {
		return str, nil
	}
	return "", errors.New("SRI could not get converted to a string")
}

func indexPackage(p Package, index *algoliasearch.Index) error {
	author, authorerr := parseAuthor(&p)
	if authorerr != nil {
		fmt.Printf("%s", authorerr)
	}

	license, licenseerr := parseLicense(&p)
	if licenseerr != nil {
		fmt.Printf("%s", licenseerr)
	}

	repository, repositoryerr := parseRepository(&p)
	if repositoryerr != nil {
		fmt.Printf("%s", repositoryerr)
	}

	github, githuberr := getGitHubMeta(repository)
	if githuberr != nil {
		fmt.Printf("%s", githuberr)
		if strings.Contains(githuberr.Error(), "403 API rate limit") {
			return fmt.Errorf("Fatal error `%s`", githuberr)
		}
	}

	sri, srierr := getSRI(&p)
	if srierr != nil {
		fmt.Printf("%s", srierr)
	}

	searchEntry := SearchEntry{
		Name:             p.Name,
		Filename:         p.Filename,
		Description:      p.Description,
		Keywords:         p.Keywords,
		AlternativeNames: getAlternativeNames(p.Name),
		FileType:         strings.ReplaceAll(filepath.Ext(p.Filename), ".", ""),
		Github:           github,
		ObjectID:         p.Name,
		Version:          p.Version,
		License:          license,
		Homepage:         p.Homepage,
		Repository:       repository,
		Author:           author,
		OriginalName:     p.Name,
		Sri:              sri,
	}
	if str, ok := p.Namespace.(string); ok {
		searchEntry.Namespace = str
	}

	_, err := index.SaveObject(searchEntry)
	return err
}

func main() {
	defer sentry.PanicHandler()
	flag.Parse()

	switch subcommand := flag.Arg(0); subcommand {
	case "update":
		{
			fmt.Printf("Downloading package.min.js...")
			b := getPackagesBuffer()
			fmt.Printf("Ok\n")

			var j PackagesJSON
			util.Check(json.Unmarshal(b.Bytes(), &j))

			fmt.Printf("Building index...\n")

			algoliaClient := algolia.GetClient()
			tmpIndex := algolia.GetTmpIndex(algoliaClient)

			for _, p := range j.Packages {
				fmt.Printf("%s: ", p.Name)
				util.Check(indexPackage(p, tmpIndex))
				fmt.Printf("Ok\n")
			}
			fmt.Printf("Ok\n")

			fmt.Printf("Promoting index to production...")
			algolia.PromoteIndex(algoliaClient)
			fmt.Printf("Ok\n")
		}
	default:
		panic(fmt.Sprintf("unknown subcommand: `%s`", subcommand))
	}
}
