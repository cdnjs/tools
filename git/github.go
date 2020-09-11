package git

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/cdnjs/tools/util"

	githubapi "github.com/google/go-github/github"
	"golang.org/x/oauth2"
	githuboauth2 "golang.org/x/oauth2/github"
)

// Stars holds the number of stars for a GitHub repository.
type Stars struct {
	Stars uint `json:"stargazers_count"`
}

// GetGitHubStars uses the GitHub API to get the star count for a
// particular GitHub repository.
func GetGitHubStars(gitURL string) Stars {
	// Ex gitURL:
	// "git@github.com:chris-pearce/backpack.css.git"
	// "git+https://github.com/18F/web-design-standards.git"
	// "https://github.com/epeli/underscore.string"
	re := regexp.MustCompile(`.*github\.com[:|/](.*?)(?:\.git)?$`)
	gitHubRepository := re.ReplaceAllString(gitURL, "$1")

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
	token := util.GetEnv("GITHUB_REPO_API_KEY")
	ctx := context.Background()
	conf := &oauth2.Config{
		Endpoint: githuboauth2.Endpoint,
	}

	tc := conf.Client(ctx, &oauth2.Token{AccessToken: token})

	return githubapi.NewClient(tc)
}
