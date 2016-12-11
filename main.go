package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"

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

type StaticHandler struct {
	http.Dir
}

func loadSiteData() {
	rawSiteData, error := ioutil.ReadFile("private/site-data.json")
	panicOnError(error)
	error = json.Unmarshal(rawSiteData, &siteData)
	panicOnError(error)
	siteDataLoaded = true
}

func requestCatchAll(responseWriter http.ResponseWriter, request *http.Request) {
	if permanentRedirectOldURLs(request.URL.Path, responseWriter, request) {
		return
	}
	serveStaticFilesOr404(responseWriter, request)
}

func permanentRedirectOldURLs(currentURL string, responseWriter http.ResponseWriter, request *http.Request) bool {
	for oldURL, newURL := range siteData.URLPermanentRedirects {
		if currentURL == oldURL {
			http.Redirect(responseWriter, request, newURL, http.StatusMovedPermanently)
			return true
		}
	}
	return false
}

func serveStaticFilesOr404(responseWriter http.ResponseWriter, request *http.Request) {
	staticHandler := StaticHandler{http.Dir(webRoot)}
	staticHandler.ServeHttp(responseWriter, request)
}

func serve404OnError(error error, responseWriter http.ResponseWriter) bool {
	if error != nil {
		responseWriter.WriteHeader(http.StatusNotFound)
		errorTemplate404Content, error := ioutil.ReadFile(webRoot + "/error-templates/404.html")
		panicOnError(error)
		fmt.Fprint(responseWriter, string(errorTemplate404Content))
		return true
	}
	return false
}

func (sh *StaticHandler) ServeHttp(responseWriter http.ResponseWriter, request *http.Request) {
	staticFilePath := staticFilePath(request)

	fileHandle, error := sh.Open(staticFilePath)
	if serve404OnError(error, responseWriter) {
		return
	}
	defer fileHandle.Close()

	fileInfo, error := fileHandle.Stat()
	if serve404OnError(error, responseWriter) {
		return
	}

	if fileInfo.IsDir() {
		if request.URL.Path[len(request.URL.Path)-1] != '/' {
			http.Redirect(responseWriter, request, request.URL.Path+"/", http.StatusFound)
			return
		}

		fileHandle, error = sh.Open(staticFilePath + "/index.html")
		if serve404OnError(error, responseWriter) {
			return
		}
		defer fileHandle.Close()

		fileInfo, error = fileHandle.Stat()
		if serve404OnError(error, responseWriter) {
			return
		}
	}

	http.ServeContent(responseWriter, request, fileInfo.Name(), fileInfo.ModTime(), fileHandle)
}

func staticFilePath(request *http.Request) string {
	staticFilePath := request.URL.Path
	if !strings.HasPrefix(staticFilePath, "/") {
		staticFilePath = "/" + staticFilePath
		request.URL.Path = staticFilePath
	}
	return path.Clean(staticFilePath)
}

func panicOnError(error error) {
	if error != nil {
		log.Panic(error)
	}
}
