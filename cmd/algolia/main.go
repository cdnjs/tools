package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xtuc/cdnjs-go/algolia"
	"github.com/xtuc/cdnjs-go/cloudstorage"
	"github.com/xtuc/cdnjs-go/github"
	"github.com/xtuc/cdnjs-go/util"

	algoliasearch "github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"golang.org/x/net/context"
)

type PackagesJSON struct {
	Packages []Package `json:"packages"`
}
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

type SearchEntry struct {
	Name             string          `json:"name"`
	Filename         string          `json:"filename"`
	Description      string          `json:"description"`
	Version          string          `json:"version"`
	Keywords         []string        `json:"keywords"`
	AlternativeNames []string        `json:"alternativeNames"`
	FileType         string          `json:"fileType"`
	Github           *GitHubMeta     `json:"github"`
	ObjectID         string          `json:"objectID"`
	License          string          `json:"license"`
	Homepage         string          `json:"homepage"`
	Namespace        string          `json:"namespace,omitempty"`
	Repository       *RepositoryMeta `json:"repository"`
	Author           string          `json:"author"`
	OriginalName     string          `json:"originalName"`
	Sri              string          `json:"sri"`
}

type RepositoryMeta struct {
	Repotype string `json:"type"`
	Url      string `json:"url"`
}

type GitHubMeta struct {
	User              string `json:"user"`
	Repo              string `json:"repo"`
	Stargazers_count  int    `json:"stargazers_count"`
	Forks             int    `json:"forks"`
	Subscribers_count int    `json:"subscribers_count"`
}

func getPackagesBuffer() bytes.Buffer {
	ctx := context.Background()

	bkt, err := cloudstorage.GetBucket(ctx)
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

var re1 = regexp.MustCompile(`[^a-zA-Z]`)
var re2 = regexp.MustCompile(`(^[^A-Z]*|[A-Z]*)([A-Z][^A-Z]+|$)`)

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
			break
		}
	case map[string]interface{}:
		{
			if v["name"] != nil {
				license = v["name"].(string)
			}
			break
		}
	case nil:
		{
			break
		}
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
			break
		}
	case map[string]interface{}:
		{
			if v["name"] != nil {
				author = v["name"].(string)
			}
			break
		}
	case nil:
		{
			break
		}
	default:
		{
			return "", fmt.Errorf("unsupported author value: `%s`", p.Author)
		}
	}

	return author, nil
}

func parseRepository(p *Package) (*RepositoryMeta, error) {
	switch v := p.Repository.(type) {
	case string:
		return &RepositoryMeta{
			Repotype: "",
			Url:      v,
		}, nil
	case map[string]interface{}:
		{
			if v["type"] != nil && v["url"] != nil {
				return &RepositoryMeta{
					Repotype: v["type"].(string),
					Url:      v["url"].(string),
				}, nil
			} else {
				return nil, nil
			}
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

var githubUrl = regexp.MustCompile(`github\.com[/|:]([\w\.-]+)\/([\w\.-]+)\/?`)

func getGitHubMeta(repo *RepositoryMeta) (*GitHubMeta, error) {
	if repo == nil {
		// no repo configured
		return nil, nil
	}
	if repo.Repotype != "git" {
		return nil, fmt.Errorf("unsupported repo type `%s`", repo.Repotype)
	}

	res := githubUrl.FindAllStringSubmatch(repo.Url, -1)
	if len(res) == 0 {
		return nil, fmt.Errorf("could not parse repo URL `%s`", repo.Url)
	}

	client := github.GetClient()
	api, _, err := client.Repositories.Get(context.Background(), res[0][1], strings.ReplaceAll(res[0][2], ".git", ""))
	if err != nil {
		return nil, err
	}

	return &GitHubMeta{
		User:              api.GetOwner().GetLogin(),
		Repo:              api.GetName(),
		Stargazers_count:  api.GetStargazersCount(),
		Forks:             api.GetForksCount(),
		Subscribers_count: api.GetSubscribersCount(),
	}, nil
}

func getSRI(p *Package) (string, error) {
	jsonFile := path.Join(".", "sri", p.Name, p.Version+".json")

	var j map[string]interface{}

	data, err := ioutil.ReadFile(jsonFile)

	if err != nil {
		return "", nil
	}

	util.Check(json.Unmarshal(data, &j))

	return j[p.Filename].(string), nil
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
	flag.Parse()
	subcommand := flag.Arg(0)

	if subcommand == "update" {
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
			err := indexPackage(p, tmpIndex)
			if err != nil {
				fmt.Printf("%s\n", err)
			} else {
				fmt.Printf("Ok\n")
			}
		}
		fmt.Printf("Ok\n")

		return
	}

	panic("unknown subcommand")
}
