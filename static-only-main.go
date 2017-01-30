package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var (
	webRoot = "awestruct/_site"
)

func main() {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(serveStaticFilesOr404)
	log.Fatal(http.ListenAndServe(":8080", router))
}
