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

// Gets the protocol, either http or https.
func getProtocol() string {
	if util.HasHTTPProxy() {
		return "http"
	}
	return "https"
}

// GetGitHubStars uses the GitHub API to get the star count for a
// particular GitHub repository.
func GetGitHubStars(gitUrl string) Stars {
	// Ex gitUrl:
	// "git@github.com:chris-pearce/backpack.css.git"
	// "git+https://github.com/18F/web-design-standards.git"
	re := regexp.MustCompile(`.*github.com[:|/](.*)\.git$`)
	gitHubRepository := re.ReplaceAllString(gitUrl, "$1")

	resp, err := http.Get(getProtocol() + "://api.github.com/repos/" + gitHubRepository)
	util.Check(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	util.Check(err)

	var stars Stars
	util.Check(json.Unmarshal(body, &stars))
	return stars

}
