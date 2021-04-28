package git

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/cdnjs/tools/util"

	"github.com/davecgh/go-spew/spew"
	githubapi "github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	githuboauth2 "golang.org/x/oauth2/github"
)

var (
	GH_TOKEN = os.Getenv("GH_TOKEN")
)

// Version represents a version of a git repo.
type Version struct {
	Version    string `json:"name"`
	TarballURL string `json:"tarball_url"`
}

// Stars holds the number of stars for a GitHub repository.
type Stars struct {
	Stars uint `json:"stargazers_count"`
}

func getRepo(gitURL string) string {
	// Ex gitURL:
	// "git@github.com:chris-pearce/backpack.css.git"
	// "git+https://github.com/18F/web-design-standards.git"
	// "https://github.com/epeli/underscore.string"
	re := regexp.MustCompile(`.*github\.com[:|/](.*?)(?:\.git)?$`)
	return re.ReplaceAllString(gitURL, "$1")
}

// GetGitHubStars uses the GitHub API to get the star count for a
// particular GitHub repository.
func GetGitHubStars(gitURL string) Stars {
	gitHubRepository := getRepo(gitURL)
	resp, err := http.Get(util.GetProtocol() + "://api.github.com/repos/" + gitHubRepository)
	util.Check(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	util.Check(err)

	var stars Stars
	util.Check(json.Unmarshal(body, &stars))
	return stars
}

// GetClient gets a GitHub client to interact with its API.
func GetClient() *githubapi.Client {
	ctx := context.Background()
	conf := &oauth2.Config{
		Endpoint: githuboauth2.Endpoint,
	}

	tc := conf.Client(ctx, &oauth2.Token{AccessToken: GH_TOKEN})

	return githubapi.NewClient(tc)
}

type GetVersionsRes struct {
	Data struct {
		Repository struct {
			Refs struct {
				Nodes []GitHubVersion `json:"nodes"`
			} `json:"refs"`
		} `json:"repository"`
	} `json:"data"`
}

type GitHubVersion struct {
	Name   string `json:"name"`
	Target struct {
		CommittedDate string `json:"committedDate"`
	} `json:"target"`
}

type GraphQLRequest struct {
	Query string `json:"query"`
}

// GetVersions gets all of the versions associated with a git repo,
// as well as the latest version.
func GetVersions(ctx context.Context, gitURL string) ([]Version, error) {
	repo := getRepo(gitURL)
	parts := strings.Split(repo, "/")
	query := GraphQLRequest{Query: fmt.Sprintf(`
query {
  repository(name: "%s", owner: "%s") {
    refs(refPrefix: "refs/tags/", last: 100) {
      nodes {
        name
        target {
          ... on Tag {
            target {
              ... on Commit {
                committedDate
              }
            }
          }
        }
      }
    }
  }
}
	`, parts[1], parts[0])}
	log.Println(query)

	var res GetVersionsRes

	body, err := json.Marshal(query)
	if err != nil {
		return nil, errors.Wrap(err, "could not construct query")
	}

	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewReader(body))
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve tags")
	}

	req.Header.Set("Authorization", "bearer "+GH_TOKEN)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request")
	}

	if resp.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode response body")
		}
		return nil, errors.Errorf("GitHub GraphQL returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	// gitTags := Tags(ctx, packageGitcache)
	// util.Debugf(ctx, "found tags in git: %s\n", gitTags)

	// gitVersions := make([]Version, 0)
	// for _, tag := range gitTags {
	// 	version := strings.TrimPrefix(tag, "v")
	// 	gitVersions = append(gitVersions, Version{
	// 		Tag:       tag,
	// 		Version:   version,
	// 		TimeStamp: TimeStamp(ctx, packageGitcache, tag),
	// 	})
	// }

	// if latest := GetMostRecentVersion(gitVersions); latest != nil {
	// 	return gitVersions, &latest.Version
	// }
	spew.Dump(res)
	return nil, nil
}
