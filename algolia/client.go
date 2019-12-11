package algolia

import (
	"github.com/xtuc/cdnjs-go/util"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
)

func GetClient() *search.Client {
	return search.NewClient("2QWLVLXZB6", util.GetEnv("ALGOLIA_WRITE_API_KEY"))
}

func GetTmpIndex(client *search.Client) *search.Index {
	return client.InitIndex("libraries.tmp")
}
