package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

var (
	fDaemon = flag.Bool("d", false, "Daemon server mode")
	fTest   = flag.Bool("t", false, "Test mode")
	fUrl    = flag.String("url", "http://localhost:8000/hello", "url to be accessed")
	port    = flag.String("port", "8000", "Define TCP port to be used for http")
	sport   = flag.String("sport", "8001", "Define TCP port to be used for https")
	root    = flag.String("root", ".", "Define the root filesystem path")
	version = flag.String("version", "0.1.1", "Version number")
)

// client and server in go style
func main() {
	flag.Parse()

	if *fTest == true {
		httpTester()
		os.Exit(0)
	}

	// determine the role of client and server
	if *fDaemon == true {
		httpServer()
	} else {
		httpClient(*fUrl)
	}
}

func httpTester() error {
	ShowNetInterfaces()
	return nil
}

// http client
func httpClient(url string) error {
	log.Printf("Client mode with %s\n", url)
	res, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return err
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}

	fmt.Printf("Header: %s\n", res.Header["Content-Type"])
	fmt.Printf("Code: %s\n", res.Status)
	fmt.Printf("Data: %s\n", string(data))

	return nil
}

// http server
func httpServer() error {
	log.Println("Server mode")
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/media", mediaHandler)

	// CAUTION: don't use /static not /static/
	http.Handle("/static/", http.StripPrefix("/static/", fileServer("./static")))

	var wg sync.WaitGroup

	wg.Add(1)
	go serveHttp(&wg)

	wg.Add(1)
	go serveHttps(&wg)

	wg.Wait()

	return nil
}

// for http access
func serveHttp(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Starting HTTP server at http://127.0.0.1:" + *port)
	//http.ListenAndServe(":"+*port, fileServer())
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

// for https access
func serveHttps(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Starting HTTPS server at https://127.0.0.1:" + *sport)
	//http.ListenAndServeTLS(":"+*sport, "cert.pem", "key.pem", fileServer())
	log.Fatal(http.ListenAndServeTLS(":"+*sport, "cert.pem", "key.pem", nil))
}

// static file server
func fileServer(path string) http.Handler {
	log.Println("File server " + path)
	return http.FileServer(http.Dir(path))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Index " + r.URL.Path)
	http.ServeFile(w, r, r.URL.Path[1:])

	/*
		err := renderFile(w, r.URL.Path, "html")
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
		}
	*/
}

/*
func renderFile(w http.ResponseWriter, filename, ext string) (err error) {
	file, err := ioutil.ReadFile(filename)
	defer file.Close()

	if ext != "" {
		w.Header().Set("Content-Type", mime.TypeByExtension(ext))
	}

	return err
}
*/

func helloHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Hello %s\n", r.URL.Path)
	//log.Printf("Hello %q\n", html.EscapeString(r.URL.Path))
	fmt.Fprintf(w, "Hi there! from Stoney")
}

func mediaHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Media " + r.URL.Path)
	fmt.Fprintf(w, "Media handler required!")
}
