//=================================================================================
// Happy Media System
// one program including agents such as caster, server, player, monitor
// Author : Stoney Kang, sikang99@gmail.com
//=================================================================================
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"stoney/httpserver/src/base"
	pf "stoney/httpserver/src/protofile"
	ph "stoney/httpserver/src/protohttp"
	pt "stoney/httpserver/src/prototcp"
	pw "stoney/httpserver/src/protows"
	sr "stoney/httpserver/src/streamring"

	"github.com/bradfitz/http2"
	"golang.org/x/net/websocket"
)

//---------------------------------------------------------------------------
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

var hello_tmpl = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8" />
<script type="text/javascript" src="static/eventemitter2.min.js"></script>
<script type="text/javascript" src="static/mjpegcanvas.min.bak.js"></script>
  
<script type="text/javascript" type="text/javascript">
  function init() {
    var viewer = new MJPEGCANVAS.Viewer({
      divID : 'mjpeg',
      host : '{{ .Addr }}',
      port : {{ .Port }},
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

//---------------------------------------------------------------------------
const (
	Version   = "0.5.7"
	TCPClient = "Happy Media TCP Server"
	TCPServer = "Happy Media TCP Server"
	WSClient  = "Happy Media WS Server"
	WSServer  = "Happy Media WS Server"

	KBYTE = 1024
	MBYTE = 1024 * KBYTE

	LEN_MAX_LINE = 128
)

//---------------------------------------------------------------------------
var (
	NotSupportError = errors.New("Not supported protocol")

	fmode  = flag.String("m", "player", "Working mode of program [caster|server|player|reader|sender|receiver|shooter|catcher]")
	fhost  = flag.String("host", "localhost", "server host address")
	fport  = flag.String("port", "8000", "TCP port to be used for http")
	fports = flag.String("ports", "8001", "TCP port to be used for https")
	fport2 = flag.String("port2", "8002", "TCP port to be used for http2")
	furl   = flag.String("url", "http://localhost:8000/[index|hello|/stream]", "url to be accessed")
	froot  = flag.String("root", ".", "Define the root filesystem path")
	vflag  = flag.Bool("verbose", false, "Verbose display")
)

func init() {
	//log.SetOutput(os.Stdout)
	//log.SetPrefix("TRACE: ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// flag setting and parsing
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: %v [flags], v.%s\n\n", os.Args[0], Version)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		os.Exit(1)
	}

	flag.Parse()
}

type ServerConfig struct {
	Title        string
	Image        string
	Addr         string
	Host         string
	Port         string
	PortS        string
	Port2        string
	Mode         string
	ImageChannel chan []byte
	Player       io.Writer
	Ring         *sr.StreamRing
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		ImageChannel: make(chan []byte, 2)}
}

var conf = ServerConfig{
	Title:        "Happy Media System: MJPEG",
	Image:        "static/image/gophergun.png",
	Addr:         "http://localhost",
	Host:         *fhost,
	Port:         *fport,
	PortS:        *fports,
	Port2:        *fport2,
	Mode:         *fmode,
	ImageChannel: make(chan []byte, 2),
}

//---------------------------------------------------------------------------
// a single program including client and server in go style
//---------------------------------------------------------------------------
func main() {
	//flag.Parse()

	// check arguments
	if flag.NFlag() == 0 && flag.NArg() == 0 {
		flag.Usage()
	}

	url, err := url.ParseRequestURI(*furl)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("Happy Media System, v.%s\n", Version)
	fmt.Printf("Default config: %s %s\n", url.Scheme, url.Host)

	// determine the working type of the program
	fmt.Printf("Working mode: %s\n", *fmode)

	conf.Ring = ph.PrepareRing(3, MBYTE, "Server stream")

	// let's do working mode
	switch *fmode {
	// package protohttp
	case "reader":
		ActHttpReader(*furl, conf.Ring)
	case "player":
		ActHttpPlayer(*furl)
	case "caster":
		ActHttpCaster(*furl)
	case "monitor":
		ActHttpMonitor()
	case "server":
		ActHttpServer()

	// package prototcp
	case "sender":
		ts := pt.NewProtoTcp("localhost", "8087", "T-Tx")
		ts.ActSender()
	case "receiver":
		tr := pt.NewProtoTcp("localhost", "8087", "T-Rx")
		tr.ActReceiver(conf.Ring)

	// package protows
	case "shooter":
		ws := pw.NewProtoWs("localhost", "8087", "8443", "W-Tx")
		ws.ActShooter()
	case "catcher":
		wr := pw.NewProtoWs("localhost", "8087", "8443", "W-Rx")
		wr.ActCatcher()

	// package protofile
	case "filer":
		fr := pf.NewProtoFile("./static/image/*.jpg", "F-Tx")
		fr.ActReader(conf.Ring)

	default:
		fmt.Println("Unknown working mode")
		os.Exit(0)
	}
}

//---------------------------------------------------------------------------
// http monitor client
//---------------------------------------------------------------------------
func ActHttpMonitor() error {
	log.Printf("Happy Media Monitor\n")

	base.ShowNetInterfaces()

	return nil
}

//---------------------------------------------------------------------------
// http player client
//---------------------------------------------------------------------------
func ActHttpPlayer(url string) error {
	log.Printf("Happy Media Player for %s\n", url)

	hp := ph.NewProtoHttp("localhost", "8080")
	client := ph.NewClientConfig()
	return ph.SendRequestGet(client, hp)
}

//---------------------------------------------------------------------------
// http caster client
//---------------------------------------------------------------------------
func ActHttpCaster(url string) error {
	log.Printf("Happy Media Caster for %s\n", url)

	hp := ph.NewProtoHttp("localhost", "8080")
	client := ph.NewClientConfig()
	return ph.SendRequestPost(client, hp)
}

//---------------------------------------------------------------------------
//	multipart reader entry, mainly from camera
//---------------------------------------------------------------------------
func ActHttpReader(url string, sbuf *sr.StreamRing) {
	log.Printf("Happy Media Reader for %s\n", url)

	var err error
	var res *http.Response

	// WHY: different behavior?
	if strings.Contains(url, "axis") {
		res, err = http.Get(url)
	} else {
		client := ph.NewClientConfig()
		res, err = client.Get(url)
	}
	if err != nil {
		log.Fatalf("GET of %q: %v", url, err)
	}
	//log.Printf("Content-Type: %v", res.Header.Get("Content-Type"))

	sbuf.Boundary, err = ph.GetBoundary(res.Header.Get("Content-Type"))
	mr := multipart.NewReader(res.Body, sbuf.Boundary)

	err = ph.RecvMultipartToRing(mr, sbuf)
}

//---------------------------------------------------------------------------
// http server entry
//---------------------------------------------------------------------------
func ActHttpServer() error {
	log.Println("Happy Media Server")

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hello", helloHandler)   // view
	http.HandleFunc("/media", mediaHandler)   // on-demand
	http.HandleFunc("/stream", streamHandler) // live
	http.HandleFunc("/search", searchHandler) // server info
	http.HandleFunc("/status", statusHandler) // server status

	http.Handle("/websocket/", websocket.Handler(websocketHandler))

	// CAUTION: don't use /static not /static/ as the prefix
	http.Handle("/static/", http.StripPrefix("/static/", ph.FileServer("./static")))

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

	/*
		wg.Add(1)
		// WS server
		go serveWs(&wg)

		wg.Add(1)
		// WSS server
		go serveWss(&wg)
	*/

	//go ActHttpReader("http://imoment:imoment@192.168.0.91/axis-cgi/mjpg/video.cgi", conf.Ring)

	tr := pt.NewProtoTcp("localhost", "8087", "T-Rx")
	go tr.ActReceiver(conf.Ring)

	//fr := pf.NewProtoFile("./static/image/*.jpg", "F-Rx")
	//go fr.ActReader(conf.Ring)

	wg.Wait()

	return nil
}

//---------------------------------------------------------------------------
// for http access
//---------------------------------------------------------------------------
func serveHttp(wg *sync.WaitGroup) {
	log.Println("Starting HTTP server at http://" + *fhost + ":" + *fport)
	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + *fport,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

//---------------------------------------------------------------------------
// for https tls access
//---------------------------------------------------------------------------
func serveHttps(wg *sync.WaitGroup) {
	log.Println("Starting HTTPS server at https://" + *fhost + ":" + *fports)
	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + *fports,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServeTLS("sec/cert.pem", "sec/key.pem"))
}

//---------------------------------------------------------------------------
// for http2 tls access
//---------------------------------------------------------------------------
func serveHttp2(wg *sync.WaitGroup) {
	log.Println("Starting HTTP2 server at https://" + *fhost + ":" + *fport2)
	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + *fport2,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	http2.ConfigureServer(srv, &http2.Server{})
	log.Fatal(srv.ListenAndServeTLS("sec/cert.pem", "sec/key.pem"))
}

//---------------------------------------------------------------------------
// for ws access
// http://www.ajanicij.info/content/websocket-tutorial-go
//---------------------------------------------------------------------------
func serveWs(wg *sync.WaitGroup) {
	log.Println("Starting WS server at https://" + *fhost + ":" + *fport2)
	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + *fport2,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

//---------------------------------------------------------------------------
// for wss access
//---------------------------------------------------------------------------
func serveWss(wg *sync.WaitGroup) {
	log.Println("Starting WSS server at https://" + *fhost + ":" + *fport2)
	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + *fport2,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServeTLS("sec/cert.pem", "sec/key.pem"))
}

//---------------------------------------------------------------------------
// index file handler
//---------------------------------------------------------------------------
func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Index %s to %s\n", r.URL.Path, r.Host)

	if strings.Contains(r.URL.Path, "favicon.ico") {
		http.ServeFile(w, r, "static/favicon.ico")
		return
	}

	ph.SendTemplatePage(w, index_tmpl, conf)
}

//---------------------------------------------------------------------------
// hello file handler (default: hello.html)
//---------------------------------------------------------------------------
func helloHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Hello %s to %s\n", r.URL.Path, r.Host)

	hello_page := "static/hello.html"

	host := strings.Split(r.Host, ":")
	conf.Port = host[1]

	if r.TLS != nil {
		conf.Addr = "https://localhost"
	}

	_, err := os.Stat(hello_page)
	if err != nil {
		ph.SendTemplatePage(w, hello_tmpl, conf)
		log.Printf("Hello serve %s\n", "hello_tmpl")
	} else {
		http.ServeFile(w, r, hello_page)
		log.Printf("Hello serve %s\n", hello_page)
	}
}

//---------------------------------------------------------------------------
// media file handler
//---------------------------------------------------------------------------
func mediaHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Media %s to %s\n", r.URL.Path, r.Host)

	_, err := os.Stat(r.URL.Path[1:])
	if err != nil {
		ph.SendResponseMessage(w, 404, r.URL.Path+" is Not Found")
	} else {
		ph.SendResponseMessage(w, 200, r.URL.Path)
	}
}

//---------------------------------------------------------------------------
// media file handler
//---------------------------------------------------------------------------
func websocketHandler(ws *websocket.Conn) {
	log.Printf("Websocket \n")

	err := websocket.Message.Send(ws, "Not yet implemented")
	if err != nil {
		log.Println(err)
	}
}

//---------------------------------------------------------------------------
// handle /search access
//---------------------------------------------------------------------------
func searchHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Search %s for %s to %s\n", r.Method, r.URL.Path, r.Host)

	err := ph.SendResponseMessage(w, 200, "/search: Not yet implemented")
	if err != nil {
		log.Println(err)
	}

	return
}

//---------------------------------------------------------------------------
// handle /search access
//---------------------------------------------------------------------------
func statusHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Search %s for %s to %s\n", r.Method, r.URL.Path, r.Host)

	err := ph.SendResponseMessage(w, 200, "/status: Not yet implemented")
	if err != nil {
		log.Println(err)
	}
}

//---------------------------------------------------------------------------
// handle /stream access
//---------------------------------------------------------------------------
func streamHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Stream %s for %s to %s\n", r.Method, r.URL.Path, r.Host)

	var err error
	sbuf := conf.Ring

	switch r.Method {
	case "POST": // for Caster
		sbuf.Boundary, err = ph.GetBoundary(r.Header.Get("Content-Type"))

		err = ph.SendResponsePost(w, sbuf.Boundary)
		if err != nil {
			log.Println(err)
			break
		}

		mr := multipart.NewReader(r.Body, sbuf.Boundary)

		err = ph.RecvMultipartToRing(mr, sbuf)
		if err != nil {
			log.Println(err)
			break
		}

	case "GET": // for Player
		err = ph.SendResponseGet(w, sbuf.Boundary)
		if err != nil {
			log.Println(err)
			break
		}

		err = ph.SendMultipartRing(w, sbuf)
		if err != nil {
			log.Println(err)
			break
		}

	default:
		log.Println("Unknown method: ", r.Method)
	}

	return
}

// ---------------------------------E-----N-----D--------------------------------
