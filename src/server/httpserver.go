package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	//"../base"
	"stoney/httpserver/src/base"

	"github.com/bradfitz/http2"
)

var index_tmpl = `<!DOCTYPE html>
<html>
<head>
</head>
<body>
<center>
<h2>Hello! from Stoney Kang, a Novice Gopher</h2>.
<img src="{{ .Image }}">Gopher with a gun</img>
</center>
</body>
</html>
`

var mjpeg_tmpl = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8" />
<script type="text/javascript" src="static/eventemitter2.min.js"></script>
<script type="text/javascript" src="static/mjpegcanvas.min.js"></script>
  
<script type="text/javascript" type="text/javascript">
  function init() {
    var viewer = new MJPEGCANVAS.Viewer({
      divID : 'mjpeg',
      host : 'localhost',
      port : 8080,
      width : 1024,
      height : 768,
      topic : 'agilecam'
    });
  }
</script>
</head>

<body onload="init()">
<center>
  <h1>{{ .Title }}</h1>
  <div id="mjpeg"></div>
</center>
</body>
</html>
`

type Config struct {
	Title string
	Image string
	Host  string
	Port  string
	Mode  string
}

var (
	NotSupportError = errors.New("Not supported protocol")

	fmode  = flag.String("m", "player", "Working mode of program")
	fhost  = flag.String("host", "localhost", "server host address")
	fport  = flag.String("port", "8000", "Define TCP port to be used for http")
	fports = flag.String("ports", "8001", "Define TCP port to be used for https")
	fport2 = flag.String("port2", "8002", "Define TCP port to be used for http2")
	furl   = flag.String("url", "http://localhost:8000/hello", "url to be accessed")
	froot  = flag.String("root", ".", "Define the root filesystem path")
	vflag  = flag.Bool("version", false, "0.2.0")
)

func init() {
	//log.SetOutput(os.Stdout)
	//log.SetPrefix("TRACE: ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// parse command options
	flag.Parse()
}

// a single program including client and server in go style
func main() {
	//flag.Parse()

	url, err := url.ParseRequestURI(*furl)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("%s %s\n", url.Scheme, url.Host)

	// determine the working type of the program
	fmt.Printf("Working mode : %s\n", *fmode)

	switch *fmode {
	case "reader":
		streamReader(*furl)
	case "player":
		httpPlayer(*furl)
	case "caster":
		httpCaster(*furl)
	case "monitor":
		httpMonitor()
	case "server":
		httpServer()
	default:
		fmt.Println("Unknown mode")
		os.Exit(0)
	}
}

// http monitor
func httpMonitor() error {
	base.ShowNetInterfaces()
	return nil
}

// http player client
func httpPlayer(url string) error {
	log.Printf("httpPlayer %s\n", url)

	// simple tls setting
	tp := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tp}

	err := httpClientGet(client, url)
	return err
}

// http caster client
func httpCaster(url string) error {
	log.Printf("httpCaster %s\n", url)

	// simple tls setting
	tp := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tp}

	err := httpClientPost(client, url)
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
	mjpg := base.MjpegNew()
	header_type := fmt.Sprint("multipart/x-mixed-replace; boundary=--myboundary")

	println("1")
	res, err := client.Post(url, header_type, io.Reader(mjpg))
	//res, err := client.Post(url, header_type, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	println("2")
	// Content-Type: video/mjpeg

	//sendStreamData(client.Conn)

	//body := new(bytes.Buffer)
	//writer := multipart.NewWriter(body)
	return err
}

func printHttpHeader(h http.Header) {
	for k, v := range h {
		fmt.Println("key:", k, "value:", v)
	}
}

func streamReader(url string) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatalf("GET of %q: %v", url, err)
	}
	log.Printf("Content: %v", res.Header.Get("Content-Type"))

	_, params, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	fmt.Printf("%v %v\n", params, res.Header.Get("Content-Type"))
	if err != nil {
		log.Fatalf("ParseMediaType: %v", err)
	}

	boundary := params["boundary"]
	if !strings.HasPrefix(boundary, "--") {
		log.Printf("expected boundary to start with --, got %q", boundary)
	}
	r := multipart.NewReader(res.Body, boundary)
	decodeFrames(r)

}

func decodeFrames(r *multipart.Reader) {
	for {
		print(".")
		p, err := r.NextPart()
		if err != nil {
			log.Fatalf("NextPart: %v", err)
		}
		_, err = jpeg.Decode(p)
		if err != nil {
			log.Fatalf("jpeg Decode: %v", err)
		}
	}
}

// http server
func httpServer() error {
	log.Println("Server mode")
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/media", mediaHandler)   // on-demand
	http.HandleFunc("/stream", streamHandler) // live

	// CAUTION: don't use /static not /static/ as the prefix
	http.Handle("/static/", http.StripPrefix("/static/", fileServer("./static")))

	//var wg sync.WaitGroup
	wg := sync.WaitGroup{}

	wg.Add(1)
	// HTTP server
	go serveHttp(&wg)

	wg.Add(1)
	// HTTPS server
	go serveHttps(&wg)

	wg.Add(1)
	// HTTP2 server
	go serveHttp2(&wg)

	wg.Wait()

	return nil
}

// for http access
func serveHttp(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Starting HTTP server at http://" + *fhost + ":" + *fport)
	log.Fatal(http.ListenAndServe(":"+*fport, nil))
}

// for https tls access
func serveHttps(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Starting HTTPS server at https://" + *fhost + ":" + *fports)
	log.Fatal(http.ListenAndServeTLS(":"+*fports, "sec/cert.pem", "sec/key.pem", nil))
}

// for http2 tls access
func serveHttp2(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Starting HTTP2 server at https://" + *fhost + ":" + *fport2)

	var srv http.Server
	srv.Addr = ":" + *fport2
	http2.ConfigureServer(&srv, &http2.Server{})
	log.Fatal(srv.ListenAndServeTLS("sec/cert.pem", "sec/key.pem"))
}

// static file server
func fileServer(path string) http.Handler {
	log.Println("File server for " + path)
	return http.FileServer(http.Dir(path))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Index " + r.URL.Path)

	if strings.Contains(r.URL.Path, "favicon.ico") {
		http.ServeFile(w, r, "static/favicon.ico")
		return
	}

	sendPage(w, index_tmpl)
}

var conf = Config{
	Title: "Simple MJPEG Canvas Player",
	Image: "static/image/gophergun.jpg",
	Host:  *fhost,
	Port:  *fport,
	Mode:  *fmode,
}

func sendFile(w http.ResponseWriter, file string) error {
	w.Header().Set("Content-Type", "image/icon")
	w.Header().Set("Server", "Happy Media Server")
	body, err := ioutil.ReadFile("static/favicon.ico")
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Fprint(w, string(body))

	return nil
}

func sendPage(w http.ResponseWriter, page string) error {
	t, err := template.New("mjpeg").Parse(page)
	if err != nil {
		log.Println(err)
		return err
	}

	return t.Execute(w, conf)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Hello %s\n", r.URL.Path)

	mjpeg_page := "static/mjpeg_canvas.html"

	_, err := os.Stat(mjpeg_page)
	if err != nil {
		sendPage(w, mjpeg_tmpl)
		log.Printf("Hello %s\n", "mjpeg_tmpl")
	} else {
		http.ServeFile(w, r, mjpeg_page)
		log.Printf("Hello %s\n", mjpeg_page)

	}
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

const boundary = "myboundary"
const frameheader = "\r\n" +
	"--" + boundary + "\r\n" +
	"Content-Type: video/mjpeg\r\n" +
	"Content-Length: %d\r\n" +
	"X-Timestamp: 0.000000\r\n" +
	"\r\n"

//func sendStreamFile(w http.ResponseWriter, file string) error {
func sendStreamFile(w io.Writer, file string) error {
	jpeg, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println(err)
		return err
	}
	//fmt.Printf("%s [%d] : %0x%0x - %0x%0x\n", file, len(jpeg), jpeg[0], jpeg[1], jpeg[len(jpeg)-2], jpeg[len(jpeg)-1])

	header := fmt.Sprintf(frameheader, len(jpeg))
	frame := make([]byte, (len(header) + len(jpeg)))

	copy(frame, header)
	copy(frame[len(header):], jpeg)

	_, err = w.Write(frame)
	//n, err := w.Write(frame)
	//fmt.Printf("%s [%d=%d] : \n", file, len(header)+len(jpeg), n)

	return err
}

func sendStreamRequest(w http.ResponseWriter) error {
	return nil
}

func sendStreamResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=--"+boundary)
	w.Header().Set("Server", "Happy Media Server")
	w.WriteHeader(http.StatusOK)

	return nil
}

func sendStreamData(w io.Writer) error {
	var err error

	for {
		err = sendStreamFile(w, "static/image/arducar.jpg")
		if err != nil {
			log.Println(err)
			break
		}
		time.Sleep(1 * time.Second)

		err = sendStreamFile(w, "static/image/gopher.jpg")
		if err != nil {
			log.Println(err)
			break
		}
		time.Sleep(1 * time.Second)
	}

	return err
}

func streamHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Stream %s for %s\n", r.Method, r.URL.Path)

	switch r.Method {
	case "POST":
		log.Println("POST")

	case "GET":
		err := sendStreamResponse(w)
		if err != nil {
			log.Println(err)
			break
		}

		err = sendStreamData(w)
		if err != nil {
			log.Println(err)
		}

	default:
		log.Println("Unknown method")
	}

	return
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
