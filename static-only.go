package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"
)

type StaticHandler struct {
	http.Dir
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

func (sh *StaticHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {
	staticFilePath := staticFilePath(r)

	fileHandle, error := sh.Open(staticFilePath)
	if serve404OnErr(error, w) {
		return
	}
	defer fileHandle.Close()

	fileInfo, error := fileHandle.Stat()
	if serve404OnErr(error, w) {
		return
	}

	if fileInfo.IsDir() {
		if r.URL.Path[len(r.URL.Path)-1] != '/' {
			http.Redirect(w, r, r.URL.Path+"/", http.StatusFound)
			return
		}

		fileHandle, error = sh.Open(staticFilePath + "/index.html")
		if serve404OnErr(error, w) {
			return
		}
		defer fileHandle.Close()

		fileInfo, error = fileHandle.Stat()
		if serve404OnErr(error, w) {
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

func panicOnError(error error) {
	if error != nil {
		log.Panic(error)
	}
}
