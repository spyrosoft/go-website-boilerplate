# Boilerplate

Abstracted website boilerplate to be included in all of my Golang websites.

Example main.go:

```go
package main

import (
	"log"
	"net/http"

	"github.com/spyrosoft/go-website-boilerplate"
	"github.com/julienschmidt/httprouter"
)

type SiteData struct {
	LiveOrDev             string            `json:"live-or-dev"`
	URLPermanentRedirects map[string]string `json:"url-permanent-redirects"`
}

var (
	webRoot        = "awestruct/_site"
	siteData       = SiteData{}
	siteDataLoaded = false
)

func main() {
	loadSiteData()
	router := httprouter.New()
	router.POST("/example-ajax-uri", exampleAJAXFunction)
	router.NotFound = http.HandlerFunc(requestCatchAll)
	log.Fatal(http.ListenAndServe(":8080", router))
}

//TODO: Rename this function
func exampleAJAXFunction(responseWriter http.ResponseWriter, request *http.Request, requestParameters httprouter.Params) {
}
```