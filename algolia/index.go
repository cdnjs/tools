package algolia

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/cdnjs/tools/github"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sentry"
	"github.com/cdnjs/tools/util"
)

func init() {
	sentry.Init()
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

var githubURL = regexp.MustCompile(`github\.com[/|:]([\w\.-]+)\/([\w\.-]+)\/?`)

func getGitHubMeta(repo *packages.Repository) (*GitHubMeta, error) {
	if repo == nil {
		// no repo configured
		return nil, nil
	}

	res := githubURL.FindAllStringSubmatch(*repo.URL, -1)
	if len(res) == 0 {
		return nil, fmt.Errorf("could not parse repo URL `%s`", *repo.URL)
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

func getSRI(p *packages.Package) (string, error) {
	jsonFile := path.Join(util.GetSRIsPath(), *p.Name, *p.Version+".json")

	var j map[string]interface{}

	data, err := ioutil.ReadFile(jsonFile)

	if err != nil {
		return "", nil
	}

	util.Check(json.Unmarshal(data, &j))

	if p.Filename == nil {
		return "", errors.New("SRI could not get converted to a string (nil filename)")
	}
	if str, ok := j[*p.Filename].(string); ok {
		return str, nil
	}
	return "", errors.New("SRI could not get converted to a string")
}

// IndexPackage saves a package to the Algolia.
func IndexPackage(p packages.Package, index *search.Index) error {
	var author string
	if p.Author != nil {
		author = *p.Author
	}

	var license string
	if p.License != nil {
		license = *p.License
	}

	var filename string
	if p.Filename != nil {
		filename = *p.Filename
	}

	var homepage string
	if p.Homepage != nil {
		homepage = *p.Homepage
	}

	repository := p.Repository

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
		Name:             *p.Name,
		Filename:         filename,
		Description:      *p.Description,
		Keywords:         p.Keywords,
		AlternativeNames: getAlternativeNames(*p.Name),
		FileType:         strings.ReplaceAll(filepath.Ext(filename), ".", ""),
		Github:           github,
		ObjectID:         *p.Name,
		Version:          *p.Version,
		License:          license,
		Homepage:         homepage,
		Repository:       repository,
		Author:           author,
		OriginalName:     *p.Name,
		Sri:              sri,
	}

	_, err := index.SaveObject(searchEntry)
	return err
}
