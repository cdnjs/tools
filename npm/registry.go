package npm

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/cdnjs/tools/util"
)

type MonthlyDownload struct {
	Downloads uint `json:"downloads"`
}

func Exists(name string) bool {
	resp, err := http.Get("https://registry.npmjs.org/" + name)
	util.Check(err)
	return resp.StatusCode == http.StatusOK
}

func GetMonthlyDownload(name string) MonthlyDownload {
	resp, err := http.Get("https://api.npmjs.org/downloads/point/last-month/" + name)
	util.Check(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	util.Check(err)

	var counts MonthlyDownload
	util.Check(json.Unmarshal(body, &counts))
	return counts
}
