package algolia

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cdnjs/tools/git"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/pkg/errors"
)

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

	client := git.GetClient()
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

func getSRI(p *packages.Package, srimap map[string]string) (string, error) {
	if p.Filename == nil {
		return "", errors.New("SRI could not get converted to a string (nil filename)")
	}
	str, ok := srimap[*p.Filename]
	if !ok {
		return "", errors.Errorf("SRI could not be found for file %s", *p.Filename)
	}
	return str, nil
}

// IndexPackage saves a package to the Algolia.
func IndexPackage(p *packages.Package, index *search.Index, srimap map[string]string) error {
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

	github, err := getGitHubMeta(p.Repository)
	if err != nil {
		fmt.Printf("%s", err)
		if strings.Contains(err.Error(), "403 API rate limit") {
			return fmt.Errorf("Fatal error `%s`", err)
		}
	}

	sri, err := getSRI(p, srimap)
	if err != nil {
		fmt.Printf("failed to get SRI: %s", err)
	}

	if p.Version == nil {
		s := ""
		p.Version = &s
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
		Repository:       p.Repository,
		Author:           author,
		OriginalName:     *p.Name,
		Sri:              sri,
	}

	_, err = index.SaveObject(searchEntry)
	return err
}
