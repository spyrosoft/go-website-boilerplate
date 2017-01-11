package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
)

type StaticHandler struct {
	http.Dir
}

func loadSiteData() {
	rawSiteData, err := ioutil.ReadFile("private/site-data.json")
	panicOnErr(err)
	err = json.Unmarshal(rawSiteData, &siteData)
	panicOnErr(err)
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

func serve404OnErr(err error, responseWriter http.ResponseWriter) bool {
	if err != nil {
		serve404(responseWriter)
		return true
	}
	return false
}

func serve404(responseWriter http.ResponseWriter) {
	responseWriter.WriteHeader(http.StatusNotFound)
	template, err := ioutil.ReadFile(webRoot + "/error-templates/404.html")
	if err != nil {
		template = []byte("Error 404 - Page Not Found. Additionally a 404 page template could not be found.")
	}
	fmt.Fprint(responseWriter, string(template))
}

func serve500(responseWriter http.ResponseWriter) {
	responseWriter.WriteHeader(http.StatusNotFound)
	template, err := ioutil.ReadFile(webRoot + "/error-templates/500.html")
	if err != nil {
		template = []byte("Error 500 - Internal Server Error. Additionally a 500 page template could not be found.")
	}
	fmt.Fprint(responseWriter, string(template))
}

func (sh *StaticHandler) ServeHttp(responseWriter http.ResponseWriter, request *http.Request) {
	staticFilePath := staticFilePath(request)

	fileHandle, err := sh.Open(staticFilePath)
	if serve404OnErr(err, responseWriter) {
		return
	}
	defer fileHandle.Close()

	fileInfo, err := fileHandle.Stat()
	if serve404OnErr(err, responseWriter) {
		return
	}

	if fileInfo.IsDir() {
		if request.URL.Path[len(request.URL.Path)-1] != '/' {
			http.Redirect(responseWriter, request, request.URL.Path+"/", http.StatusFound)
			return
		}

		fileHandle, err = sh.Open(staticFilePath + "/index.html")
		if serve404OnErr(err, responseWriter) {
			return
		}
		defer fileHandle.Close()

		fileInfo, err = fileHandle.Stat()
		if serve404OnErr(err, responseWriter) {
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

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
