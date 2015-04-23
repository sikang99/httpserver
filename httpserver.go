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

func main() {
	flag.Parse()

	log.Println("Preparing http/https service ...")
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
