package git

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/cdnjs/tools/util"
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
