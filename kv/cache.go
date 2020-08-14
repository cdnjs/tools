package kv

import (
	"fmt"

	"github.com/cdnjs/tools/util"
	"github.com/cloudflare/cloudflare-go"
)

// PurgeTags purges the zone's cache by tags.
func PurgeTags(tags []string) error {
	resp, err := api.PurgeCache(zoneID, cloudflare.PurgeCacheRequest{
		Tags: tags,
	})
	util.Check(err)
	fmt.Println(resp)
	return nil
}
