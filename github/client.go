package github

import (
	"context"

	"github.com/cdnjs/tools/util"

	githubapi "github.com/google/go-github/github"
	"golang.org/x/oauth2"
	githuboauth2 "golang.org/x/oauth2/github"
)

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
