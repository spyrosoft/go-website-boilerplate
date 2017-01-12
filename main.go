package main

import (
	"log"
	"net/http"

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
func exampleAJAXFunction(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}
