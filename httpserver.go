package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
)

var (
	port    = flag.String("port", "8000", "Define what TCP port to bind to")
	sport   = flag.String("sport", "8001", "Define what TCP port to bind to")
	root    = flag.String("root", ".", "Define the root filesystem path")
	version = flag.String("version", "0.1.0", "Version number")
)

var wg sync.WaitGroup

func main() {
	flag.Parse()

	http.HandleFunc("/hello", helloHandler)
	//http.HandleFunc("/static", fileHandler)
	http.Handle("/static", http.FileServer(http.Dir("./static/")))

	//fs := http.FileServer(http.Dir("./static"))
	//http.Handle("/static", http.StripPrefix("/static", fs))

	wg.Add(1)
	go serveHttp()

	wg.Add(1)
	go serveHttps()

	wg.Wait()
}

func serveHttp() {
	defer wg.Done()
	log.Println("Starting HTTP server at http://127.0.0.1:" + *port)
	//http.ListenAndServe(":"+*port, fileServer())
	http.ListenAndServe(":"+*port, nil)
}

func serveHttps() {
	defer wg.Done()
	log.Println("Starting HTTPS server at https://127.0.0.1:" + *sport)
	//http.ListenAndServeTLS(":"+*sport, "cert.pem", "key.pem", fileServer())
	http.ListenAndServeTLS(":"+*sport, "cert.pem", "key.pem", nil)
}

func fileServer() http.Handler {
	return http.FileServer(http.Dir(*root))
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	http.ServeFile(w, r, r.URL.Path[1:])
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there!")
}
