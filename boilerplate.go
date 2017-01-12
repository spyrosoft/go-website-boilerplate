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

func requestCatchAll(w http.ResponseWriter, r *http.Request) {
	if permanentRedirectOldURLs(r.URL.Path, w, r) {
		return
	}
	serveStaticFilesOr404(w, r)
}

func permanentRedirectOldURLs(currentURL string, w http.ResponseWriter, r *http.Request) bool {
	for oldURL, newURL := range siteData.URLPermanentRedirects {
		if currentURL == oldURL {
			http.Redirect(w, r, newURL, http.StatusMovedPermanently)
			return true
		}
	}
	return false
}

func serveStaticFilesOr404(w http.ResponseWriter, r *http.Request) {
	staticHandler := StaticHandler{http.Dir(webRoot)}
	staticHandler.ServeHttp(w, r)
}

func serve404OnErr(err error, w http.ResponseWriter) bool {
	if err != nil {
		serve404(w)
		return true
	}
	return false
}

func serve404(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	template, err := ioutil.ReadFile(webRoot + "/error-templates/404.html")
	if err != nil {
		template = []byte("Error 404 - Page Not Found. Additionally a 404 page template could not be found.")
	}
	fmt.Fprint(w, string(template))
}

func serve500(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	template, err := ioutil.ReadFile(webRoot + "/error-templates/500.html")
	if err != nil {
		template = []byte("Error 500 - Internal Server Error. Additionally a 500 page template could not be found.")
	}
	fmt.Fprint(w, string(template))
}

func (sh *StaticHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {
	staticFilePath := staticFilePath(r)

	fileHandle, err := sh.Open(staticFilePath)
	if serve404OnErr(err, w) {
		return
	}
	defer fileHandle.Close()

	fileInfo, err := fileHandle.Stat()
	if serve404OnErr(err, w) {
		return
	}

	if fileInfo.IsDir() {
		if r.URL.Path[len(r.URL.Path)-1] != '/' {
			http.Redirect(w, r, r.URL.Path+"/", http.StatusFound)
			return
		}

		fileHandle, err = sh.Open(staticFilePath + "/index.html")
		if serve404OnErr(err, w) {
			return
		}
		defer fileHandle.Close()

		fileInfo, err = fileHandle.Stat()
		if serve404OnErr(err, w) {
			return
		}
	}

	http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), fileHandle)
}

func staticFilePath(r *http.Request) string {
	staticFilePath := r.URL.Path
	if !strings.HasPrefix(staticFilePath, "/") {
		staticFilePath = "/" + staticFilePath
		r.URL.Path = staticFilePath
	}
	return path.Clean(staticFilePath)
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
