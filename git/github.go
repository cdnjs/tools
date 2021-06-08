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
	"time"

	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
	"github.com/cdnjs/tools/version"

	githubapi "github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	githuboauth2 "golang.org/x/oauth2/github"
)

var (
	GH_TOKEN = os.Getenv("GH_TOKEN")
)

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
		Target struct {
			TarballUrl    string `json:"tarballUrl"`
			CommittedDate string `json:"committedDate"`
			AuthoredDate  string `json:"authoredDate"`
		} `json:"target"`
	} `json:"target"`
}

type GraphQLRequest struct {
	Query string `json:"query"`
}

// GetVersions gets all of the versions associated with a git repo,
// as well as the latest version.
func GetVersions(ctx context.Context, config *packages.Autoupdate) ([]version.Version, error) {
	name := *config.Target
	repo := getRepo(*config.Target)
	parts := strings.Split(repo, "/")
	query := GraphQLRequest{Query: fmt.Sprintf(`
query {
  repository(name: "%s", owner: "%s") {
    refs(refPrefix: "refs/tags/", last: 30) {
      nodes {
        name
        target {
          ... on Tag {
            target {
              ... on Commit {
			    tarballUrl
                committedDate
                authoredDate
              }
            }
          }
        }
      }
    }
  }
}
	`, parts[1], parts[0])}

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

	versions := make([]version.Version, 0)

	for _, githubVersion := range res.Data.Repository.Refs.Nodes {
		var err error
		date := time.Time{}
		if githubVersion.Target.Target.CommittedDate != "" {
			date, err = time.Parse(time.RFC3339, githubVersion.Target.Target.CommittedDate)
		} else if githubVersion.Target.Target.AuthoredDate != "" {
			date, err = time.Parse(time.RFC3339, githubVersion.Target.Target.AuthoredDate)
		}
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse tag date")
		}

		if !version.IsVersionIgnored(config, githubVersion.Name) {
			versionName := githubVersion.Name
			if versionName[0:1] == "v" {
				versionName = versionName[1:]
			}

			versions = append(versions, version.Version{
				Version: versionName,
				Tarball: githubVersion.Target.Target.TarballUrl,
				Date:    date,
				Source:  "git",
			})
		} else {
			log.Printf("%s: version %s is ignored\n", name, githubVersion.Name)
		}
	}

	return versions, nil
}
