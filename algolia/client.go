package algolia

import (
	"github.com/cdnjs/tools/util"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/opt"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
)

var (
	tmpIndex  = "libraries.tmp"
	prodIndex = "libraries"
)

func GetClient() *search.Client {
	return search.NewClient("2QWLVLXZB6", util.GetEnv("ALGOLIA_WRITE_API_KEY"))
}

func GetTmpIndex(client *search.Client) *search.Index {
	index := client.InitIndex(tmpIndex)
	_, err := index.SetSettings(search.Settings{
		SearchableAttributes: opt.SearchableAttributes(
			"unordered(name)",
			"unordered(alternativeNames)",
			"unordered(github.repo)",
			"unordered(description)",
			"unordered(keywords)",
			"unordered(filename)",
			"unordered(repositories.url)",
			"unordered(github.user)",
			"unordered(maintainers.name)",
		),
		CustomRanking: opt.CustomRanking(
			"desc(github.stargazers_count)", "asc(name)",
		),
		AttributesForFaceting: opt.AttributesForFaceting(
			"fileType", "keywords",
		),
		OptionalWords: opt.OptionalWords(
			"js", "css",
		),
	})
	util.Check(err)
	return index
}

func PromoteIndex(client *search.Client) {
	_, err := client.MoveIndex(tmpIndex, prodIndex)
	util.Check(err)
}
