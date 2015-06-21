//=================================================================================
// Happy Media System
// one program including agents such as caster, server, player, monitor
// Author : Stoney Kang, sikang99@gmail.com
//=================================================================================
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	pf "stoney/httpserver/src/protofile"
	ph "stoney/httpserver/src/protohttp"
	pt "stoney/httpserver/src/prototcp"
	pw "stoney/httpserver/src/protows"

	sb "stoney/httpserver/src/streambase"
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
	STR_MEDIA_VERSION = "0.8.7"
	STR_MEDIA_SYSTEM  = "Happy Media System"
)

//---------------------------------------------------------------------------
var (
	fmode  = flag.String("m", "player", "Working mode of program [caster|server|player|reader|sender|receiver|shooter|catcher]")
	fhost  = flag.String("host", "localhost", "server host address")
	fport  = flag.String("port", "8000", "TCP port to be used for http")
	fports = flag.String("ports", "8001", "TCP port to be used for https")
	fport2 = flag.String("port2", "8002", "TCP port to be used for http2")
	furl   = flag.String("url", "http://localhost:8000/[index|hello|/stream]", "url to be accessed")
	froot  = flag.String("root", ".", "Define the root filesystem path")
	vflag  = flag.Bool("verbose", false, "Verbose display")
)

//---------------------------------------------------------------------------
func init() {
	//log.SetOutput(os.Stdout)
	//log.SetPrefix("TRACE: ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// flag setting and parsing
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: %v [flags], v.%s\n\n", os.Args[0], STR_MEDIA_VERSION)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		os.Exit(1)
	}

	flag.Parse()
}

//---------------------------------------------------------------------------
type ServerConfig struct {
	Title        string
	Image        string
	Addr         string
	Host         string
	Port         string
	PortS        string
	Port2        string
	Mode         string
	Ring         *sr.StreamRing
	ImageChannel chan []byte
	// http://giantmachines.tumblr.com/post/52184842286/golang-http-client-with-timeouts
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
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

	fmt.Printf("%s, v.%s\n", STR_MEDIA_SYSTEM, STR_MEDIA_VERSION)
	fmt.Printf("Default config: %s %s\n", url.Scheme, url.Host)
	fmt.Printf("Working mode: %s\n", *fmode)

	conf.Ring = ph.PrepareRing(3, sb.MBYTE, "Server ring buffer")

	// let's do by the working mode
	switch *fmode {
	// package protohttp
	case "http_reader":
		ActHttpReader(*furl, conf.Ring)
	case "http_player":
		ActHttpPlayer(*furl, conf.Ring)
	case "http_caster":
		ActHttpCaster(*furl)
	case "http_monitor":
		ActHttpMonitor(*furl)
	case "http_server":
		ActHttpServer()

	// package prototcp
	case "tcp_caste":
		pt.NewProtoTcpWithParams("localhost", "8087", "T-Tx").ActCaster()
	case "tcp_server":
		tr := pt.NewProtoTcpWithParams("localhost", "8087", "T-Rx")
		tr.ActServer(conf.Ring)
	case "tcp_player":
		tr := pt.NewProtoTcpWithParams("localhost", "8087", "T-Rx")
		tr.ActPlayer(conf.Ring)

	// package protows
	case "ws_caster":
		ws := pw.NewProtoWs("localhost", "8087", "8443", "W-Tx")
		ws.StreamCaster()
	case "ws_server":
		wr := pw.NewProtoWs("localhost", "8087", "8443", "W-Rx")
		wr.StreamServer()
	case "ws_player":
		wr := pw.NewProtoWs("localhost", "8087", "8443", "W-Rx")
		wr.StreamPlayer()

	// package protofile
	case "file_reader":
		fr := pf.NewProtoFile("./static/image/*.jpg", "F-Rr")
		fr.StreamReader(conf.Ring)
	case "file_writer":
		fr := pf.NewProtoFile("./static/image/*.jpg", "F-Wr")
		fr.StreamWriter(conf.Ring)

	default:
		fmt.Println("Unknown working mode")
		os.Exit(0)
	}
}

//---------------------------------------------------------------------------
// parse command
//---------------------------------------------------------------------------
func ParseCommand(cmdstr string) error {
	var err error

	//fmt.Println(cmdstr)
	/*
		r := bufio.NewReader(cmdstr)
		var s scanner.Scanner
		s.Init(cmdstr)
		tok := s.Scan()
		for tok != scanner.EOF {
			fmt.Print(tok)
			tok = s.Scan()
		}
	*/

	res := strings.Fields(cmdstr)
	if len(res) < 1 {
		return err
	}
	//fmt.Println(res)

	switch res[0] {
	case "show":
		if len(res) < 2 {
			fmt.Printf("usage: show [network|channel|ring]\n")
			break
		}
		switch res[1] {
		case "network":
			sb.ShowNetInterfaces()
		case "channel":
			fallthrough
		case "ring":
			fmt.Printf("i will %s for %s shortly\n", res[0], res[1])
		default:
			fmt.Printf("I can't %s for %s\n", res[0], res[1])
		}
	case "help":
		if len(res) < 2 {
			fmt.Printf("usage: help [show|act]\n")
			break
		}
		switch res[1] {
		case "act":
			fallthrough
		case "show":
			fmt.Printf("i will %s for %s shortly\n", res[0], res[1])
		default:
			fmt.Printf("I can't %s for %s\n", res[0], res[1])
		}
	default:
		fmt.Printf("usage: [show|help|quit]\n")
	}

	return err
}

//---------------------------------------------------------------------------
// http monitor client
//---------------------------------------------------------------------------
func ActHttpMonitor(url string) error {
	log.Printf("%s for %s\n", ph.STR_HTTP_MONITOR, url)

	var err error

	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")

		line, _, err := r.ReadLine()
		if err != nil {
			continue
		}

		cmdstr := strings.Replace(string(line), "\r", "", -1)

		if strings.EqualFold(cmdstr, "quit") {
			fmt.Println("Bye bye.")
			return err
		}

		err = ParseCommand(cmdstr)
		if err != nil {
			log.Println(err)
			break
		}
	}

	return err
}

//---------------------------------------------------------------------------
// http player client
//---------------------------------------------------------------------------
func ActHttpPlayer(url string, sbuf *sr.StreamRing) error {
	log.Printf("%s for %s\n", ph.STR_HTTP_PLAYER, url)

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

	sbuf.Boundary, err = ph.GetTypeBoundary(res.Header.Get("Content-Type"))
	mr := multipart.NewReader(res.Body, sbuf.Boundary)

	err = ph.ReadMultipartToRing(mr, sbuf)

	return err
}

//---------------------------------------------------------------------------
// http caster client
//---------------------------------------------------------------------------
func ActHttpCaster(url string) error {
	log.Printf("%s for %s\n", ph.STR_HTTP_CASTER, url)

	hp := ph.NewProtoHttp("localhost", "8080")
	client := ph.NewClientConfig()
	return hp.RequestPost(client)
}

//---------------------------------------------------------------------------
//	multipart reader entry, mainly from camera
//---------------------------------------------------------------------------
func ActHttpReader(url string, sbuf *sr.StreamRing) {
	log.Printf("%s for %s\n", ph.STR_HTTP_READER, url)

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

	sbuf.Boundary, err = ph.GetTypeBoundary(res.Header.Get("Content-Type"))
	mr := multipart.NewReader(res.Body, sbuf.Boundary)

	err = ph.ReadMultipartToRing(mr, sbuf)
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

	go ActHttpReader("http://imoment:imoment@192.168.0.91/axis-cgi/mjpg/video.cgi", conf.Ring)
	//go pt.NewProtoTcp("localhost", "8087", "T-Rx").ActServer(conf.Ring)
	//go pf.NewProtoFile("./static/image/*.jpg", "F-Rx").ActCaster(conf.Ring)

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

	ph.WriteTemplatePage(w, index_tmpl, conf)
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
		ph.WriteTemplatePage(w, hello_tmpl, conf)
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
		ph.WriteResponseMessage(w, 404, r.URL.Path+" is Not Found")
	} else {
		ph.WriteResponseMessage(w, 200, r.URL.Path)
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

	err := ph.WriteResponseMessage(w, 200, "/search: Not yet implemented")
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

	err := ph.WriteResponseMessage(w, 200, "/status: Not yet implemented")
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
		sbuf.Boundary, err = ph.GetTypeBoundary(r.Header.Get("Content-Type"))

		err = ph.ResponsePost(w, sbuf.Boundary)
		if err != nil {
			log.Println(err)
			break
		}

		mr := multipart.NewReader(r.Body, sbuf.Boundary)

		err = ph.ReadMultipartToRing(mr, sbuf)
		if err != nil {
			log.Println(err)
			break
		}

	case "GET": // for Player
		err = ph.ResponseGet(w, sbuf.Boundary)
		if err != nil {
			log.Println(err)
			break
		}

		err = ph.WriteRingInMultipart(w, sbuf)
		if err != nil {
			log.Println(err)
			break
		}

	default:
		log.Println("Unknown request method: ", r.Method)
	}

	return
}

// ---------------------------------E-----N-----D--------------------------------
