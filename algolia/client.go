// Package algolia contains functions used to interact with the
// Algolia Seach API to update the search index used for autocomplete on
// the cdnjs website.
package algolia

import (
	"github.com/cdnjs/tools/util"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/opt"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
)

var (
	// The Algolia temporary index was used by the previous bot
	// and I just copied it over because I didn't know how it worked.
	// I don't see the value of building a temporary index and at the
	// end renaming it to the production index. I would rather maintain
	// a single index and only update entries when the bot pushed a new update.
	tmpIndex  = "libraries.tmp" // temp Algolia index
	prodIndex = "libraries"     // production Algolia index
)

// GetClient instantiates a new client to interact with the Algolia Search API
// using an Application ID and API key.
func GetClient() *search.Client {
	return search.NewClient("2QWLVLXZB6", util.GetEnv("ALGOLIA_WRITE_API_KEY"))
}

// GetProdIndex gets the Algolia production index.
func GetProdIndex(client *search.Client) *search.Index {
	return client.InitIndex(prodIndex)
}

// GetTmpIndex instantiates and configures a new temporary Algolia Search index.
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

// PromoteIndex promotes the temporary Algolia Search index to production.
func PromoteIndex(client *search.Client) {
	_, err := client.MoveIndex(tmpIndex, prodIndex)
	util.Check(err)
}
