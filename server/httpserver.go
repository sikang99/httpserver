package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"../base"
	"github.com/bradfitz/http2"
)

var (
	NotSupportError = errors.New("Not supported protocol")

	fDaemon = flag.Bool("d", false, "Daemon server mode")
	fMon    = flag.Bool("m", false, "Monitor mode, especillay for web")
	fUrl    = flag.String("url", "http://localhost:8000/hello", "url to be accessed")
	port    = flag.String("port", "8000", "Define TCP port to be used for http")
	ports   = flag.String("ports", "8001", "Define TCP port to be used for https")
	port2   = flag.String("port2", "8002", "Define TCP port to be used for http2")
	root    = flag.String("root", ".", "Define the root filesystem path")
	version = flag.String("version", "0.1.1", "Version number")
)

func init() {
	log.SetOutput(os.Stdout)
	//log.SetPrefix("TRACE: ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// parse command options
	flag.Parse()
}

// a single program including client and server in go style
func main() {
	//flag.Parse()

	if *fMon == true {
		httpMonitor()
		os.Exit(0)
	}

	// determine the role of client and server
	if *fDaemon == true {
		httpServer()
	} else {
		httpClient(*fUrl)
	}
}

// http monitor
func httpMonitor() error {
	base.ShowNetInterfaces()
	return nil
}

// http client
func httpClient(url string) error {
	log.Printf("http.Client %s\n", url)

	// simple tls setting
	tp := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tp}

	err := httpClientGet(client, url)
	return err
}

func httpClientGet(client *http.Client, url string) error {
	res, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	//printHttpHeader(res.Header)
	ct := res.Header.Get("Content-Type")

	println("")
	fmt.Printf("Response Code: %s\n", res.Status)
	fmt.Printf("Content-Type: %s\n", res.Header["Content-Type"])
	if strings.Contains(ct, "text") == true {
		fmt.Printf("%s\n", string(body))
	} else {
		fmt.Printf("[binary data]\n")
	}
	println("")

	return err
}

func httpClientPost(client *http.Client, url string) error {
	res, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	// Content-Type: video/mjpeg

	return err
}

func printHttpHeader(h http.Header) {
	for k, v := range h {
		fmt.Println("key:", k, "value:", v)
	}
}

// http server
func httpServer() error {
	log.Println("Server mode")
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/media/", mediaHandler)

	// CAUTION: don't use /static not /static/ as the prefix
	http.Handle("/static/", http.StripPrefix("/static/", fileServer("./static")))

	//var wg sync.WaitGroup
	wg := sync.WaitGroup{}

	wg.Add(1)
	go serveHttp(&wg)

	wg.Add(1)
	go serveHttps(&wg)

	wg.Add(1)
	go serveHttp2(&wg)

	wg.Wait()

	return nil
}

// for http access
func serveHttp(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Starting HTTP server at http://localhost:" + *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

// for https tls access
func serveHttps(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Starting HTTPS server at https://localhost:" + *ports)
	log.Fatal(http.ListenAndServeTLS(":"+*ports, "sec/cert.pem", "sec/key.pem", nil))
}

// for http2 tls access
func serveHttp2(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Starting HTTP2 server at https://localhost:" + *port2)

	var srv http.Server
	srv.Addr = ":" + *port2
	http2.ConfigureServer(&srv, &http2.Server{})
	log.Fatal(srv.ListenAndServeTLS("sec/cert.pem", "sec/key.pem"))
}

// static file server
func fileServer(path string) http.Handler {
	log.Println("File server " + path)
	return http.FileServer(http.Dir(path))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Index " + r.URL.Path[1:])
	http.ServeFile(w, r, r.URL.Path[1:])
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	//log.Printf("Hello %s\n", r.URL.Path)
	log.Printf("Hello %q\n", html.EscapeString(r.URL.Path))
	fmt.Fprintf(w, "Hi there! from Stoney")
}

func mediaHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Media " + r.URL.Path)

	_, err := os.Stat(r.URL.Path[1:])
	if err != nil {
		Responder(w, r, 404, r.URL.Path+" is Not Found")
	} else {
		Responder(w, r, 200, r.URL.Path)
	}
}

func Responder(w http.ResponseWriter, r *http.Request, status int, message string) {
	/*
		w.Header().Set("Content-Type", mime.TypeByExtension(ext))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
	*/
	w.WriteHeader(status)
	log.Println(message)
	fmt.Fprintf(w, message)
}
