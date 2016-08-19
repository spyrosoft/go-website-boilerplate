package main

import (
	"fmt"
	"net/http"
	"log"
	"strings"
	"path"
	"io/ioutil"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
)

type Credentials struct {
	LiveOrDev string `json:"live-or-dev"`
}

type StaticHandler struct {
	http.Dir
}

var (
	webRoot = "awestruct/_site"
	credentials = Credentials{}
	credentialsHaveBeenLoaded = false
)

func loadCredentials() {
	rawCredentials, error := ioutil.ReadFile("private/credentials.json")
	panicOnError(error)
	error = json.Unmarshal(rawCredentials, &credentials)
	panicOnError(error)
	credentialsHaveBeenLoaded = true
}

func (sh *StaticHandler) ServeHttp(responseWriter http.ResponseWriter, request *http.Request) {
	staticFilePath := staticFilePath(request)
	
	fileHandle, error := sh.Open(staticFilePath)
	if serve404OnError(error, responseWriter) { return }
	defer fileHandle.Close()
	
	fileInfo, error := fileHandle.Stat()
	if serve404OnError(error, responseWriter) { return }
	
	if fileInfo.IsDir() {
		fileHandle, error = sh.Open(staticFilePath + "/index.html")
		if serve404OnError(error, responseWriter) { return }
		defer fileHandle.Close()
		
		fileInfo, error = fileHandle.Stat()
		if serve404OnError(error, responseWriter) { return }
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

func panicOnError(error error) { if error != nil { log.Panic(error) } }

//TODO: Place this function in its proper place
func exampleAJAXFunction(responseWriter http.ResponseWriter, request *http.Request, requestParameters httprouter.Params) {}

func main() {
	loadCredentials()
	router := httprouter.New()
	router.POST("/example-ajax-uri", exampleAJAXFunction)
	router.NotFound = http.HandlerFunc(serveStaticFilesOr404)
	log.Fatal(http.ListenAndServe(":8080", router))
}