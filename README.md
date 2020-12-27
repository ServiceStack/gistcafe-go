Useful utils for [gist.cafe](https://gist.cafe) Go Apps.

## Usage

Simple usage example:

```go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"

	"github.com/servicestack/gistcafe-go/inspect"
)

type GithubRepo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Homepage    string `json:"homepage"`
	Lang        string `json:"language"`
	Watchers    int    `json:"watchers_count"`
	Forks       int    `json:"forks"`
}

func main() {
	orgName := "golang"
	url := fmt.Sprintf("https://api.github.com/orgs/%s/repos", orgName)
	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", "gist.cafe")
	res, getErr := client.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}
	defer res.Body.Close()

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	var orgRepos []GithubRepo
	jsonErr := json.Unmarshal(body, &orgRepos)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	sort.Slice(orgRepos, func(i, j int) bool {
		return orgRepos[i].Watchers > orgRepos[j].Watchers
	})

	inspect.PrintDump(orgRepos[0:3])

	inspect.TableOptions{Headers: []string{"Name", "Lang", "Watchers", "Forks"}}.PrintDumpTable(orgRepos[0:10])

	inspect.Vars(map[string]interface{}{"orgRepos": orgRepos})
}
```

Which outputs:

```
[
    {
        name: go,
        description: The Go programming language,
        homepage: https://golang.org,
        language: Go,
        watchers_count: 80228,
        forks: 11669
    },
    {
        name: groupcache,
        description: groupcache is a caching and cache-filling library, intended as a replacement for memcached in many cases.,
        homepage: ,
        language: Go,
        watchers_count: 9556,
        forks: 1086
    },
    {
        name: protobuf,
        description: Go support for Google's protocol buffers,
        homepage: ,
        language: Go,
        watchers_count: 7311,
        forks: 1348
    }
]
+------------+------+----------+-------+
|    NAME    | LANG | WATCHERS | FORKS |
+------------+------+----------+-------+
| go         | Go   |    80228 | 11669 |
| groupcache | Go   |     9556 |  1086 |
| protobuf   | Go   |     7311 |  1348 |
| mock       | Go   |     4996 |   395 |
| tools      | Go   |     4824 |  1637 |
| mobile     | Go   |     4377 |   560 |
| lint       | Go   |     3764 |   504 |
| oauth2     | Go   |     3423 |   715 |
| glog       | Go   |     2634 |   720 |
| net        | Go   |     2254 |   915 |
+------------+------+----------+-------+
```

Whilst `inspect.vars()` lets you view variables in [gist.cafe](https://gist.cafe) viewer:

![](https://raw.githubusercontent.com/ServiceStack/gist-cafe/main/docs/images/vars-orgRepos-nodejs.png)

View and execute Dart gists with [gist.cafe](https://gist.cafe), e.g: [gist.cafe/58d4e2d53d8982ae108198e91fee4a69](https://gist.cafe/58d4e2d53d8982ae108198e91fee4a69).

## Features and bugs

Please file feature requests and bugs at the [issue tracker](https://github.com/ServiceStack/gistcafe-node/issues).