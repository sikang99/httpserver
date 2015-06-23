//=========================================================================
// Author : Stoney Kang, sikang99@gmail.com, 2015
// Protocol for HTTP streaming
// - http://www.sanarias.com/blog/1214PlayingwithimagesinHTTPresponseingolang
// - http://stackoverflow.com/questions/30552447/how-to-set-which-ip-to-use-for-a-http-request
// - http://giantmachines.tumblr.com/post/52184842286/golang-http-client-with-timeouts
//=========================================================================

package mediaconf

import (
	"bufio"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/websocket"

	"github.com/bradfitz/http2"

	ph "stoney/httpserver/src/protohttp"
	sb "stoney/httpserver/src/streambase"
	sr "stoney/httpserver/src/streamring"
)

//---------------------------------------------------------------------------
// http monitor client
//---------------------------------------------------------------------------
func (sc *ServerConfig) StreamMonitor(url string) error {
	log.Printf("%s for %s\n", ph.STR_HTTP_MONITOR, url)

	var err error
	var prestr string

	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")

		line, _, err := r.ReadLine()
		if err != nil {
			log.Println(err)
			break
		}

		cmdstr := strings.Replace(string(line), "\r", "", -1)
		if cmdstr == "" {
			continue
		}

		if strings.EqualFold(cmdstr, "quit") {
			fmt.Println("Bye bye.")
			break
		} else if cmdstr == "." {
			// previous command
			cmdstr = prestr
		}

		// remember command
		prestr = cmdstr

		err = sc.ParseCommand(cmdstr)
		if err != nil {
			log.Println(err)
			break
		}
	}

	return err
}

//---------------------------------------------------------------------------
// parse command
//---------------------------------------------------------------------------
func (sc *ServerConfig) ParseCommand(cmdstr string) error {
	var err error

	//fmt.Println(cmdstr)
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
			fmt.Printf("i will %s for %s shortly\n", res[0], res[1])
		case "ring":
			fmt.Println(sc.Ring)
		default:
			fmt.Printf("I can't %s for %s\n", res[0], res[1])
		}

	case "start":
		if len(res) < 2 {
			fmt.Printf("usage: start [read|write]\n")
			break
		}
		switch res[1] {
		case "read":
			go sc.StreamReader("http://imoment:imoment@192.168.0.91/axis-cgi/mjpg/video.cgi", sc.Ring)
		case "write":
			fallthrough
		default:
			fmt.Printf("I can't %s for %s\n", res[0], res[1])
		}

	case "help":
		if len(res) < 2 {
			fmt.Printf("usage: help [show|start]\n")
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
		fmt.Printf("usage: [show|start|help|quit]\n")
	}

	return err
}

//---------------------------------------------------------------------------
//	multipart reader entry, mainly from camera
//---------------------------------------------------------------------------
func (sc *ServerConfig) StreamReader(url string, ring *sr.StreamRing) error {
	log.Printf("%s for %s\n", ph.STR_HTTP_READER, url)

	var err error
	var res *http.Response

	// WHY: different behavior?
	if strings.Contains(url, "axis") {
		res, err = http.Get(url)
	} else {
		client := ph.NewClient()
		res, err = client.Get(url)
	}
	if err != nil {
		log.Fatalf("GET of %q: %v", url, err)
	}
	//log.Printf("Content-Type: %v", res.Header.Get("Content-Type"))

	ring.Boundary, err = ph.GetTypeBoundary(res.Header.Get("Content-Type"))
	mr := multipart.NewReader(res.Body, ring.Boundary)

	err = ph.ReadMultipartToRing(mr, ring)

	return err
}

//---------------------------------------------------------------------------
// http caster client
//---------------------------------------------------------------------------
func (sc *ServerConfig) StreamCaster(url string) error {
	log.Printf("%s for %s\n", ph.STR_HTTP_CASTER, url)

	hp := ph.NewProtoHttp()
	client := ph.NewClient()
	return hp.RequestPost(client)
}

//---------------------------------------------------------------------------
// http player client
//---------------------------------------------------------------------------
func (sc *ServerConfig) StreamPlayer(url string, ring *sr.StreamRing) error {
	log.Printf("%s for %s\n", ph.STR_HTTP_PLAYER, url)

	var err error
	var res *http.Response

	// WHY: different behavior?
	if strings.Contains(url, "axis") {
		res, err = http.Get(url)
	} else {
		client := ph.NewClient()
		res, err = client.Get(url)
	}
	if err != nil {
		log.Fatalf("GET of %q: %v", url, err)
	}
	//log.Printf("Content-Type: %v", res.Header.Get("Content-Type"))

	ring.Boundary, err = ph.GetTypeBoundary(res.Header.Get("Content-Type"))
	mr := multipart.NewReader(res.Body, ring.Boundary)

	err = ph.ReadMultipartToRing(mr, ring)

	return err
}

//---------------------------------------------------------------------------
// http server entry
//---------------------------------------------------------------------------
func (sc *ServerConfig) StreamServer(ring *sr.StreamRing) error {
	log.Printf("%s\n", ph.STR_HTTP_SERVER)

	http.HandleFunc("/", sc.IndexHandler)
	http.HandleFunc("/hello", sc.HelloHandler)   // view
	http.HandleFunc("/media", sc.MediaHandler)   // on-demand
	http.HandleFunc("/stream", sc.StreamHandler) // live
	http.HandleFunc("/search", sc.SearchHandler) // server info
	http.HandleFunc("/status", sc.StatusHandler) // server status

	http.Handle("/websocket/", websocket.Handler(sc.WebsocketHandler))

	// CAUTION: don't use /static not /static/ as the prefix
	http.Handle("/static/", http.StripPrefix("/static/", FileServer("./static")))

	//var wg sync.WaitGroup
	wg := sync.WaitGroup{}

	wg.Add(1)
	// HTTP server
	go sc.ServeHttp(&wg)

	wg.Add(1)
	// HTTPS server
	go sc.ServeHttps(&wg)

	wg.Add(1)
	// HTTP2 server
	go sc.ServeHttp2(&wg)

	/*
		wg.Add(1)
		// WS server
		go sc.ServeWs(&wg)

		wg.Add(1)
		// WSS server
		go sc.ServeWss(&wg)
	*/

	go sc.StreamReader("http://imoment:imoment@192.168.0.91/axis-cgi/mjpg/video.cgi", ring)
	//go pt.NewProtoTcp("localhost", "8087", "T-Rx").StreamServer(ring)
	//go pf.NewProtoFile("./static/image/*.jpg", "F-Rx").StreamCaster(ring)

	wg.Wait()

	return nil
}

//---------------------------------------------------------------------------
// index file handler
//---------------------------------------------------------------------------
func (sc *ServerConfig) IndexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Index %s to %s\n", r.URL.Path, r.Host)

	if strings.Contains(r.URL.Path, "favicon.ico") {
		http.ServeFile(w, r, "static/favicon.ico")
		return
	}

	ph.WriteTemplatePage(w, index_tmpl, sc)
}

//---------------------------------------------------------------------------
// hello file handler (default: hello.html)
//---------------------------------------------------------------------------
func (sc *ServerConfig) HelloHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Hello %s to %s\n", r.URL.Path, r.Host)

	hello_page := "static/hello.html"

	host := strings.Split(r.Host, ":")
	sc.Port = host[1]

	if r.TLS != nil {
		sc.Addr = "https://localhost"
	}

	_, err := os.Stat(hello_page)
	if err != nil {
		ph.WriteTemplatePage(w, hello_tmpl, sc)
		log.Printf("Hello serve %s\n", "hello_tmpl")
	} else {
		http.ServeFile(w, r, hello_page)
		log.Printf("Hello serve %s\n", hello_page)
	}
}

//---------------------------------------------------------------------------
// media file handler
//---------------------------------------------------------------------------
func (sc *ServerConfig) MediaHandler(w http.ResponseWriter, r *http.Request) {
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
func (sc *ServerConfig) WebsocketHandler(ws *websocket.Conn) {
	log.Printf("Websocket \n")

	err := websocket.Message.Send(ws, "Not yet implemented")
	if err != nil {
		log.Println(err)
	}
}

//---------------------------------------------------------------------------
// handle /search access
//---------------------------------------------------------------------------
func (sc *ServerConfig) SearchHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Search %s for %s to %s\n", r.Method, r.URL.Path, r.Host)

	var err error

	err = ph.WriteResponseMessage(w, 200, "/search: Not yet implemented")
	if err != nil {
		log.Println(err)
	}

	return
}

//---------------------------------------------------------------------------
// handle /search access
//---------------------------------------------------------------------------
func (sc *ServerConfig) StatusHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Search %s for %s to %s\n", r.Method, r.URL.Path, r.Host)

	var err error
	ring := sc.Ring

	switch r.Method {
	case "GET":
		str := fmt.Sprint(ring)
		err = ph.WriteResponseMessage(w, 200, str)
		if err != nil {
			log.Println(err)
			return
		}
	case "POST":
		fallthrough
	default:
		err = ph.WriteResponseMessage(w, 200, "/status: Not yet implemented")
		if err != nil {
			log.Println(err)
			return
		}
	}
}

//---------------------------------------------------------------------------
// handle /stream access
//---------------------------------------------------------------------------
func (sc *ServerConfig) StreamHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Stream %s for %s to %s\n", r.Method, r.URL.Path, r.Host)

	var err error
	ring := sc.Ring

	switch r.Method {
	case "POST": // for Caster
		ring.Boundary, err = ph.GetTypeBoundary(r.Header.Get("Content-Type"))

		err = ph.ResponsePost(w, ring.Boundary)
		if err != nil {
			log.Println(err)
			break
		}

		mr := multipart.NewReader(r.Body, ring.Boundary)

		err = ph.ReadMultipartToRing(mr, ring)
		if err != nil {
			log.Println(err)
			break
		}

	case "GET": // for Player
		err = ph.ResponseGet(w, ring.Boundary)
		if err != nil {
			log.Println(err)
			break
		}

		err = ph.WriteRingInMultipart(w, ring)
		if err != nil {
			log.Println(err)
			break
		}

	default:
		log.Println("Unknown request method: ", r.Method)
	}

	return
}

//---------------------------------------------------------------------------
// for http access
//---------------------------------------------------------------------------
func (sc *ServerConfig) ServeHttp(wg *sync.WaitGroup) {
	log.Println("Starting HTTP server at http://" + sc.Host + ":" + sc.Port)
	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + sc.Port,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

//---------------------------------------------------------------------------
// for https tls access
//---------------------------------------------------------------------------
func (sc *ServerConfig) ServeHttps(wg *sync.WaitGroup) {
	log.Println("Starting HTTPS server at https://" + sc.Host + ":" + sc.PortS)
	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + sc.PortS,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServeTLS("sec/cert.pem", "sec/key.pem"))
}

//---------------------------------------------------------------------------
// for http2 tls access
//---------------------------------------------------------------------------
func (sc *ServerConfig) ServeHttp2(wg *sync.WaitGroup) {
	log.Println("Starting HTTP2 server at https://" + sc.Host + ":" + sc.Port2)
	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + sc.Port2,
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
func (sc *ServerConfig) ServeWs(wg *sync.WaitGroup) {
	log.Println("Starting WS server at https://" + sc.Host + ":" + sc.Port)
	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + sc.Port,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

//---------------------------------------------------------------------------
// for wss access
//---------------------------------------------------------------------------
func (sc *ServerConfig) ServeWss(wg *sync.WaitGroup) {
	log.Println("Starting WSS server at https://" + sc.Host + ":" + sc.PortS)
	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + sc.PortS,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServeTLS("sec/cert.pem", "sec/key.pem"))
}

//---------------------------------------------------------------------------
// static file server handler
//---------------------------------------------------------------------------
func FileServer(path string) http.Handler {
	log.Println("File server for " + path)
	return http.FileServer(http.Dir(path))
}

// ---------------------------------E-----N-----D--------------------------------
