package main

import (
	"flag"
	"log"
	"net/http"
	"sync"
)

var (
	port  = flag.String("port", "8000", "Define what TCP port to bind to")
	sport = flag.String("sport", "8001", "Define what TCP port to bind to")
	root  = flag.String("root", ".", "Define the root filesystem path")
)

func main() {
	flag.Parse()

	var wg sync.WaitGroup

	wg.Add(1)
	go serveHttp()

	wg.Add(1)
	go serveHttps()

	wg.Wait()
}

func serveHttp() {
	log.Println("Starting web server at http://0.0.0.0:" + *port)
	http.ListenAndServe(":"+*port, fileHandler())
}

func serveHttps() {
	log.Println("Starting web server at https://0.0.0.0:" + *sport)
	http.ListenAndServeTLS(":"+*sport, "cert.pem", "key.pem", fileHandler())
}

func fileHandler() http.Handler {
	return http.FileServer(http.Dir(*root))
}
