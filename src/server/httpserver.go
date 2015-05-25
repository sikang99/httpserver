// ---------------------------------------------------------------------------------
// one program including agents such as caster, server, player, monitor
// ---------------------------------------------------------------------------------
package main

import (
	"bufio"
	"bytes"
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
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"stoney/httpserver/src/base"

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
	Version   = "0.4.2"
	TCPClient = "Happy Media TCP Server"
	TCPServer = "Happy Media TCP Server"
	WSClient  = "Happy Media WS Server"
	WSServer  = "Happy Media WS Server"
)

//---------------------------------------------------------------------------
var (
	NotSupportError = errors.New("Not supported protocol")

	fmode  = flag.String("m", "player", "Working mode of program")
	fhost  = flag.String("host", "localhost", "server host address")
	fport  = flag.String("port", "8000", "TCP port to be used for http")
	fports = flag.String("ports", "8001", "TCP port to be used for https")
	fport2 = flag.String("port2", "8002", "TCP port to be used for http2")
	furl   = flag.String("url", "http://localhost:8000/hello", "url to be accessed")
	froot  = flag.String("root", ".", "Define the root filesystem path")
	vflag  = flag.Bool("version", false, Version)
)

func init() {
	//log.SetOutput(os.Stdout)
	//log.SetPrefix("TRACE: ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// parse command options
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
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		ImageChannel: make(chan []byte, 2)}
}

var conf = ServerConfig{
	Title:        "Happy Media System",
	Image:        "static/image/gophergun.jpg",
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

	url, err := url.ParseRequestURI(*furl)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("Happy Media System, v.%s\n", Version)
	fmt.Printf("Default config: %s %s\n", url.Scheme, url.Host)

	// determine the working type of the program
	fmt.Printf("Working mode: %s\n", *fmode)

	// let's do working mode
	switch *fmode {
	case "reader":
		ActHttpReader(*furl)
	case "player":
		ActHttpPlayer(*furl)
	case "caster":
		ActHttpCaster(*furl)
	case "monitor":
		ActHttpMonitor()
	case "server":
		ActHttpServer()
	case "sender":
		ActTcpSender(*fhost, *fport)
	case "receiver":
		ActTcpReceiver(*fport)
	case "shooter":
		ActWsShooter(*fhost, *fport)
	case "catcher":
		ActWsCatcher(*fport)
	default:
		fmt.Println("Unknown mode")
		os.Exit(0)
	}
}

//---------------------------------------------------------------------------
// config transport with timeout
//---------------------------------------------------------------------------
var timeout = time.Duration(3 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

func httpClientConfig() *http.Client {
	// simple timeout and tls setting
	tp := &http.Transport{
		Dial:            dialTimeout,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &http.Client{Transport: tp, Timeout: timeout}
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

	client := httpClientConfig()
	return httpClientGet(client, url)
}

//---------------------------------------------------------------------------
// http caster client
//---------------------------------------------------------------------------
func ActHttpCaster(url string) error {
	log.Printf("Happy Media Caster for %s\n", url)

	client := httpClientConfig()
	return httpClientPost(client, url)
}

//---------------------------------------------------------------------------
// http GET to server
//---------------------------------------------------------------------------
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

	printHttpResponse(res, body)

	return err
}

//---------------------------------------------------------------------------
// 21 http POST to server
// http://matt.aimonetti.net/posts/2013/07/01/golang-multipart-file-upload-example/
// http://www.tagwith.com/question_781711_golang-adding-multiple-files-to-a-http-multipart-request
//---------------------------------------------------------------------------
func httpClientPost(client *http.Client, url string) error {
	// send multipart data
	outer := new(bytes.Buffer)

	mw := multipart.NewWriter(outer)
	mw.SetBoundary("myboundary")
	b := new(bytes.Buffer)

	fdata, err := ioutil.ReadFile("static/image/gopher.jpg")
	fsize := len(fdata)

	for i := 0; i < 3; i++ {
		part, _ := mw.CreatePart(textproto.MIMEHeader{
			"Content-Type":   {"image/jpeg"},
			"Content-Length": {strconv.Itoa(fsize)},
		})

		b.Write(fdata)
		b.WriteTo(part)
	}

	// prepare a connection
	req, err := http.NewRequest("POST", url, outer)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "multipart/x-mixed-replace; boundary=--myboundary")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	fmt.Println(res.Status)

	return err
}

//---------------------------------------------------------------------------
// print response headers and body (text case) for debugging
//---------------------------------------------------------------------------
func printHttpResponse(res *http.Response, body []byte) {
	h := res.Header
	for k, v := range h {
		fmt.Println("key:", k, "value:", v)
	}

	ct := res.Header.Get("Content-Type")

	println("")
	if strings.Contains(ct, "text") == true {
		fmt.Printf("%s\n", string(body))
	} else {
		fmt.Printf("[binary data]\n")
	}
	println("")
}

//---------------------------------------------------------------------------
//	multipart reader entry, mainly from camera
//---------------------------------------------------------------------------
func ActHttpReader(url string) {
	log.Printf("Happy Media Reader for %s\n", url)

	var err error
	var res *http.Response

	// WHY: different behavior?
	if strings.Contains(url, "axis") {
		res, err = http.Get(url)
	} else {
		client := httpClientConfig()
		res, err = client.Get(url)
	}
	if err != nil {
		log.Fatalf("GET of %q: %v", url, err)
	}
	log.Printf("Content-Type: %v", res.Header.Get("Content-Type"))

	mt, params, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	fmt.Printf("%v %v\n", params, res.Header.Get("Content-Type"))
	if err != nil {
		log.Fatalf("ParseMediaType: %s %v", mt, err)
	}

	boundary := params["boundary"]
	if !strings.HasPrefix(boundary, "--") {
		log.Printf("expected boundary to start with --, got %q", boundary)
	}

	mr := multipart.NewReader(res.Body, boundary)

	//recvMultipartDecodeJpeg(mr)
	recvMultipartToBuffer(mr)
}

//---------------------------------------------------------------------------
//	receive multipart data into buffer
//---------------------------------------------------------------------------
func recvMultipartToBuffer(r *multipart.Reader) error {
	var err error

	for {
		p, err := r.NextPart()
		if err != nil {
			log.Println(err)
			return err
		}

		sl := p.Header.Get("Content-Length")
		nl, err := strconv.Atoi(sl)
		if err != nil {
			log.Printf("%s %s %d\n", p.Header, sl, nl)
		}
		//println(nl)
		data := make([]byte, nl)

		// implement like ReadFull() in jpeg.Decode()
		var tn int
		for tn < nl {
			n, err := p.Read(data[tn:])
			if err != nil {
				log.Println(err)
				return err
			}
			tn += n
		}

		fmt.Printf("%s %d/%d [%02x %02x - %02x %02x]\n", p.Header.Get("Content-Type"), tn, nl, data[0], data[1], data[nl-2], data[nl-1])
		//conf.ImageChannel <- data[:n]
	}

	return err
}

//---------------------------------------------------------------------------
//	receive multipart data and decode jpeg
//---------------------------------------------------------------------------
func recvMultipartDecodeJpeg(r *multipart.Reader) {
	for {
		print(".")
		p, err := r.NextPart()
		if err != nil {
			log.Println(err)
		}

		_, err = jpeg.Decode(p)
		if err != nil {
			log.Println(err)
		}
	}
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

	http.Handle("/websocket/", websocket.Handler(websocketHandler))

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

	wg.Add(1)
	// WS server
	go serveWs(&wg)

	wg.Add(1)
	// WSS server
	go serveWss(&wg)

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
// static file server handler
//---------------------------------------------------------------------------
func fileServer(path string) http.Handler {
	log.Println("File server for " + path)
	return http.FileServer(http.Dir(path))
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

	sendPage(w, index_tmpl)
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
		sendPage(w, hello_tmpl)
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
		Responder(w, r, 404, r.URL.Path+" is Not Found")
	} else {
		Responder(w, r, 200, r.URL.Path)
	}
}

//---------------------------------------------------------------------------
// media file handler
//---------------------------------------------------------------------------
func websocketHandler(ws *websocket.Conn) {
	log.Printf("Websocket \n")

}

//---------------------------------------------------------------------------
// send a file
//---------------------------------------------------------------------------
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

//---------------------------------------------------------------------------
// send a page
//---------------------------------------------------------------------------
func sendPage(w http.ResponseWriter, page string) error {
	t, err := template.New("mjpeg").Parse(page)
	if err != nil {
		log.Println(err)
		return err
	}

	return t.Execute(w, conf)
}

//---------------------------------------------------------------------------
// send some jpeg files in multipart format
//---------------------------------------------------------------------------
func sendStreamTest(w io.Writer, loop bool) error {
	var err error

	files := []string{"static/image/arducar.jpg", "static/image/gopher.jpg"}

	for {
		for i := range files {
			err = sendStreamFile(w, files[i])
			if err != nil {
				log.Println(err)
				break
			}
			time.Sleep(1 * time.Second)
		}
		if !loop {
			break
		}
	}

	return err
}

//---------------------------------------------------------------------------
// send files with the given format(extension) in the directory
//---------------------------------------------------------------------------
func sendStreamDir(w io.Writer, dir, ext string, loop bool) error {
	var err error

	// direct pattern matching
	files, err := filepath.Glob(dir + "*" + ext)
	if err != nil {
		log.Println(err)
		return err
	}
	if files == nil {
		log.Println("no matched file")
		return err
	}

	for {
		for i := range files {
			err = sendStreamFile(w, files[i])
			if err != nil {
				return err
			}
			time.Sleep(time.Second)
		}

		if !loop {
			break
		}
	}

	/*
		// read dir and compare extension
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			log.Println(err)
			return err
		}

		for {
			for _, f := range files {
				fpath := dir + f.Name()
				if filepath.Ext(fpath) == ext {
					//fmt.Println(fpath)
					err = sendStreamFile(w, fpath)
					if err != nil {
						return err
					}
					time.Sleep(time.Second)
				}
			}

			if !loop {
				break
			}
		}
	*/
	return err
}

//---------------------------------------------------------------------------
// send any file in multipart format with boundary (standard style)
//---------------------------------------------------------------------------
func sendStreamFile(w io.Writer, file string) error {
	mw := multipart.NewWriter(w)
	mw.SetBoundary("myboundary")

	buf := new(bytes.Buffer)

	fdata, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println(err)
		return err
	}
	fsize := len(fdata)

	// check the extension at first and its content
	ctype := mime.TypeByExtension(filepath.Ext(file))
	if ctype == "" {
		ctype = http.DetectContentType(fdata)
	}

	part, err := mw.CreatePart(textproto.MIMEHeader{
		"Content-Type":   {ctype},
		"Content-Length": {strconv.Itoa(fsize)},
	})
	if err != nil {
		log.Println(err)
		return err
	}

	buf.Write(fdata)  // prepare data in the buffer
	buf.WriteTo(part) // output the part with buffer in multipart format

	return err
}

//---------------------------------------------------------------------------
// send an jpeg file in multipart format with boundary (brute force style)
//---------------------------------------------------------------------------
const boundary = "myboundary"
const frameheader = "\r\n" +
	"--" + boundary + "\r\n" +
	"Content-Type: video/mjpeg\r\n" +
	"Content-Length: %d\r\n" +
	"X-Timestamp: 0.000000\r\n" +
	"\r\n"

func sendStreamJpeg(w io.Writer, file string) error {
	var err error

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

//---------------------------------------------------------------------------
// send response for /stream with multipart format
//---------------------------------------------------------------------------
func sendStreamRequest(w http.ResponseWriter) error {
	return nil
}

// for Caster
func sendStreamOK(w http.ResponseWriter) error {
	w.Header().Set("Server", "Happy Media Server")
	w.WriteHeader(http.StatusOK)

	return nil
}

// for Player
func sendStreamResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=--"+boundary)
	w.Header().Set("Server", "Happy Media Server")
	w.WriteHeader(http.StatusOK)

	return nil
}

//---------------------------------------------------------------------------
// 22 handle /stream access
//---------------------------------------------------------------------------
func streamHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Stream %s for %s to %s\n", r.Method, r.URL.Path, r.Host)

	switch r.Method {
	case "POST": // for Caster
		/*
			err := sendStreamOK(w)
			if err != nil {
				log.Println(err)
				break
			}
		*/
		mt, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		fmt.Printf("%v %v\n", r.Header.Get("Content-Type"), params)
		if err != nil {
			log.Fatalf("ParseMediaType: %s %v", mt, err)
		}

		boundary := params["boundary"]
		if !strings.HasPrefix(boundary, "--") {
			log.Printf("expected boundary to start with --, got %q", boundary)
		}

		mr := multipart.NewReader(r.Body, boundary)

		recvMultipartToBuffer(mr)

	case "GET": // for Player
		err := sendStreamResponse(w)
		if err != nil {
			log.Println(err)
			break
		}

		err = sendStreamDir(w, "./static/image/", ".jpg", true)
		if err != nil {
			log.Println(err)
			break
		}

	default:
		log.Println("Unknown method: ", r.Method)
	}

	return
}

//---------------------------------------------------------------------------
// respond message simply
//---------------------------------------------------------------------------
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

//==================================================================================
// TCP Socket
//  - http://stackoverflow.com/questions/25090690/how-to-write-a-proxy-in-go-golang-using-tcp-connections
//  - https://github.com/nf/gohttptun - A tool to tunnel TCP over HTTP, written in Go
//==================================================================================
//---------------------------------------------------------------------------
// act TCP sender for test and debugging
//---------------------------------------------------------------------------
func ActTcpSender(hname, hport string) {
	log.Printf("Happy Media TCP Sender\n")

	addr, _ := net.ResolveTCPAddr("tcp", hname+":"+hport)
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	err = conn.SetNoDelay(true)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Connecting to %s\n", addr)

	headers, err := TcpSummitRequest(conn)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(headers)

	return
}

//---------------------------------------------------------------------------
// TCP receiver for debugging
//---------------------------------------------------------------------------
func ActTcpReceiver(hport string) {
	log.Printf("Happy Media TCP Receiver\n")

	l, err := net.Listen("tcp", ":"+hport)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer l.Close()

	log.Printf("Listening on :%s\n", hport)

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		go TcpHandleRequest(conn)
	}
}

//---------------------------------------------------------------------------
// 31 summit TCP request
//---------------------------------------------------------------------------
func TcpSummitRequest(conn net.Conn) (map[string]string, error) {
	var err error

	// send POST request
	req := "POST /stream HTTP/1.1\r\n"
	req += fmt.Sprintf("Content-Type: multipart/x-mixed-replace; boundary=%s\r\n", "myboundary")
	req += "User-Agent: Happy Media TCP Client\r\n"
	req += "\r\n"

	fmt.Printf("SEND [%d]\n%s", len(req), req)

	_, err = conn.Write([]byte(req))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	_, err = TcpReadMessage(conn)

	// send multipart stream, ex) jpg files
	err = TcpSendMultipartFiles(conn, "static/image/*.jpg")

	return nil, err
}

//---------------------------------------------------------------------------
// handle request
//---------------------------------------------------------------------------
func TcpHandleRequest(conn net.Conn) error {
	var err error

	log.Printf("in %s\n", conn.RemoteAddr())
	defer log.Printf("out %s\n", conn.RemoteAddr())
	defer conn.Close()

	// recv POST request
	_, err = TcpReadMessage(conn)
	if err != nil {
		log.Println(err)
		return err
	}

	err = TcpSendResponse(conn)
	if err != nil {
		log.Println(err)
		return err
	}

	//  recv multipart stream
	for {
		_, err = TcpReadMessage(conn)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return err
}

//---------------------------------------------------------------------------
// send TCP response
//---------------------------------------------------------------------------
func TcpSendResponse(conn net.Conn) error {
	var err error

	res := "HTTP/1.1 200 Ok\r\n"
	res += "Server: Happy Media TCP Server\r\n"
	res += "\r\n"

	fmt.Printf("SEND [%d]\n%s", len(res), res)

	_, err = conn.Write([]byte(res))
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// read a message in http header style
//---------------------------------------------------------------------------
func TcpReadMessage(conn net.Conn) (map[string]string, error) {
	var err error

	reader := bufio.NewReader(conn)

	headers, err := TcpReadHeader(reader)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	clen := 0
	value, ok := headers["CONTENT-LENGTH"]
	if ok {
		clen, _ = strconv.Atoi(value)
	}

	if clen > 0 {
		_, err = TcpReadBody(reader, clen)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	return headers, err
}

//---------------------------------------------------------------------------
// read header of message
//---------------------------------------------------------------------------
func TcpReadHeader(reader *bufio.Reader) (map[string]string, error) {
	var err error

	result := make(map[string]string)

	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			log.Println(err)
			return result, err
		}
		if string(line) == "" {
			return result, err
		}

		fmt.Println(string(line))

		keyvalue := strings.SplitN(string(line), ":", 2)
		if len(keyvalue) > 1 {
			result[strings.ToUpper(keyvalue[0])] = strings.TrimSpace(keyvalue[1])
		}
	}

	fmt.Println(result)
	return result, err
}

//---------------------------------------------------------------------------
// read body of message
//---------------------------------------------------------------------------
func TcpReadBody(reader *bufio.Reader, clen int) ([]byte, error) {
	var err error

	buf := make([]byte, clen)

	tn := 0
	for tn < clen {
		n, err := reader.Read(buf[tn:])
		if err != nil {
			log.Println(err)
			break
		}
		tn += n
	}

	fmt.Printf("[DATA] (%d/%d)\n\n", tn, clen)
	return buf, err
}

//---------------------------------------------------------------------------
// send files in the multipart
//---------------------------------------------------------------------------
func TcpSendMultipartFiles(conn net.Conn, pattern string) error {
	var err error

	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Println(err)
		return err
	}
	if files == nil {
		log.Printf("no file for '%s'\n", pattern)
		return err
	}

	for i := range files {
		fdata, err := ioutil.ReadFile(files[i])
		if err != nil {
			log.Println(err)
			return err
		}

		if len(fdata) == 0 {
			fmt.Printf(">> ignore '%s'\n", files[i])
			continue
		}

		ctype := mime.TypeByExtension(files[i])
		if ctype == "" {
			ctype = http.DetectContentType(fdata)
		}

		err = TcpSendPart(conn, fdata, ctype)
		if err != nil {
			log.Println(err)
			return err
		}

		time.Sleep(time.Second)
	}

	return err
}

//---------------------------------------------------------------------------
// send a part
//---------------------------------------------------------------------------
func TcpSendPart(conn net.Conn, data []byte, ctype string) error {
	var err error

	clen := len(data)

	req := fmt.Sprintf("--%s\r\n", "myboundary")
	req += fmt.Sprintf("Content-Type: %s\r\n", ctype)
	req += fmt.Sprintf("Content-Length: %d\r\n", clen)
	req += "\r\n"

	defer fmt.Printf("SEND [%d,%d]\n%s", len(req), clen, req)

	_, err = conn.Write([]byte(req))
	if err != nil {
		log.Println(err)
		return err
	}

	if clen > 0 {
		_, err = conn.Write(data)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return err
}

//==================================================================================
// WebSocket(WS, WSS)
// - https://github.com/golang-samples/websocket
//==================================================================================
type WsConfig struct {
	Ws       *websocket.Conn
	Boundary string
	Mr       *multipart.Reader
	Mw       *multipart.Writer
}

//---------------------------------------------------------------------------
// WebSocket shooter for test and debugging
//---------------------------------------------------------------------------
func ActWsShooter(hname, hport string) {
	log.Printf("Happy Media WS Shooter\n")

	origin := fmt.Sprintf("http://%s/", hname)
	url := fmt.Sprintf("ws://%s:%s/stream", hname, hport)

	// connect to the server
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Println(err)
		return
	}

	err = WsSummitRequest(ws)
	if err != nil {
		log.Println(err)
		return
	}

	WsSendMultipartFiles(ws, "static/image/*.jpg")
}

//---------------------------------------------------------------------------
// WebSocket catcher for test and debugging
//---------------------------------------------------------------------------
func ActWsCatcher(hport string) {
	log.Printf("Happy Media WS Catcher\n")

	http.Handle("/stream", websocket.Handler(WsStreamHandler))
	log.Fatal(http.ListenAndServe(":"+hport, nil))
}

//---------------------------------------------------------------------------
// WebSocket stream handler in the server
//---------------------------------------------------------------------------
func WsStreamHandler(ws *websocket.Conn) {
	log.Printf("in %s\n", ws.RemoteAddr())
	defer log.Printf("out %s\n", ws.RemoteAddr())

	boundary, err := WsHandleRequest(ws)
	if err != nil {
		log.Println(err)
		return
	}

	err = WsRecvMultipart(ws, boundary)
	if err != nil {
		log.Println(err)
		return
	}
}

//---------------------------------------------------------------------------
// WebSocket summit requeest in the client
//---------------------------------------------------------------------------
func WsSummitRequest(ws *websocket.Conn) error {
	var err error

	boundary := "myboundary"

	// send POST request
	smsg := "POST /stream HTTP/1.1\r\n"
	smsg += fmt.Sprintf("Content-Type: multipart/x-mixed-replace; boundary=%s\r\n", boundary)
	smsg += "User-Agent: Happy Media WS Client\r\n"
	smsg += "\r\n"

	n, err := ws.Write([]byte(smsg))
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Printf("Send(%d):\n%s", n, smsg)

	// recv response
	rmsg := make([]byte, 512)

	n, err = ws.Read(rmsg)
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Printf("Recv(%d):\n%s", n, rmsg[:n])

	return err
}

//---------------------------------------------------------------------------
// WebSocket handle request in the server
//---------------------------------------------------------------------------
func WsHandleRequest(ws *websocket.Conn) (string, error) {
	var err error

	// recv POST request
	rmsg := make([]byte, 512)

	n, err := ws.Read(rmsg)
	if err != nil {
		log.Println(err)
		return "", err
	}
	fmt.Printf("Recv(%d):\n%s", n, rmsg[:n])

	// parse request
	boundary, err := WsGetBoundary(string(rmsg[:n]))
	if err != nil {
		log.Println(err)
		return boundary, err
	}

	// send response
	smsg := "HTTP/1.1 200 Ok\r\n"
	smsg += "Server: Happy Media WS Server\r\n"
	smsg += "\r\n"

	n, err = ws.Write([]byte(smsg))
	if err != nil {
		log.Println(err)
		return boundary, err
	}
	fmt.Printf("Send(%d):\n%s", n, smsg)

	return boundary, err
}

//---------------------------------------------------------------------------
// WebSocket get boundary string
//---------------------------------------------------------------------------
func WsGetBoundary(msg string) (string, error) {
	var err error

	req, err := WsGetRequest(msg)

	mt, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	fmt.Printf("%v %v\n", params, req.Header.Get("Content-Type"))
	if err != nil {
		log.Printf("ParseMediaType: %s %v", mt, err)
		return "", err
	}

	boundary := params["boundary"]
	if !strings.HasPrefix(boundary, "--") {
		log.Printf("expected boundary to start with --, got %q", boundary)
	}

	return boundary, err
}

//---------------------------------------------------------------------------
// WebSocket get boundary string
//---------------------------------------------------------------------------
func WsGetRequest(msg string) (*http.Request, error) {
	var err error

	/*
		reader := bufio.NewReader(strings.NewReader(msg))
		tp := textproto.NewReader(reader)

		mimeHeader, err := tp.ReadMIMEHeader()
		if err != nil {
			log.Println(err)
		}

		httpHeader := http.Header(mimeHeader)
		fmt.Println(httpHeader)
	*/

	reader := bufio.NewReader(strings.NewReader(msg))
	req, err := http.ReadRequest(reader)
	if err != nil {
		log.Println(err)
		return req, err
	}
	fmt.Println(req.Header)

	return req, err
}

//---------------------------------------------------------------------------
// WebSocket send multipart
//---------------------------------------------------------------------------
func WsSendMultipartFiles(ws *websocket.Conn, pattern string) error {
	var err error

	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Println(err)
		return err
	}
	if files == nil {
		log.Printf("no file for '%s'\n", pattern)
		return err
	}

	//mw := multipart.NewWriter(os.Stdout)	// for debug
	mw := multipart.NewWriter(ws)
	mw.SetBoundary("myboundary")

	for i := range files {
		fmt.Println(files[i])
		fdata, err := ioutil.ReadFile(files[i])
		if err != nil {
			log.Println(err)
			return err
		}

		if len(fdata) == 0 {
			fmt.Printf(">> ignore '%s'\n", files[i])
			continue
		}

		ctype := mime.TypeByExtension(files[i])
		if ctype == "" {
			ctype = http.DetectContentType(fdata)
		}

		err = WsSendPart(mw, fdata, len(fdata), ctype)
		if err != nil {
			log.Println(err)
			return err
		}

		time.Sleep(time.Second)
	}

	return err
}

//---------------------------------------------------------------------------
// WebSocket recv multipart
//---------------------------------------------------------------------------
func WsRecvMultipart(ws *websocket.Conn, boundary string) error {
	var err error

	mr := multipart.NewReader(ws, boundary)

	for {
		err = WsRecvPart(mr)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return err
}

//---------------------------------------------------------------------------
// WebSocket send part
//---------------------------------------------------------------------------
func WsSendPart(mw *multipart.Writer, data []byte, dsize int, dtype string) error {
	var err error

	buf := new(bytes.Buffer)

	part, err := mw.CreatePart(textproto.MIMEHeader{
		"Content-Type":   {dtype},
		"Content-Length": {strconv.Itoa(dsize)},
	})
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = buf.Write(data)   // dn, prepare data in the buffer
	_, err = buf.WriteTo(part) // tn(int64), output the part with buffer in multipart format

	return err
}

//---------------------------------------------------------------------------
// WebSocket recv part
//---------------------------------------------------------------------------
func WsRecvPart(mr *multipart.Reader) error {
	var err error

	p, err := mr.NextPart()
	if err != nil {
		log.Println(err)
		return err
	}

	sl := p.Header.Get("Content-Length")
	nl, err := strconv.Atoi(sl)
	if err != nil {
		log.Printf("%s %s %d\n", p.Header, sl, nl)
		return err
	}

	data := make([]byte, nl)

	// implement like ReadFull() in jpeg.Decode()
	var tn int
	for tn < nl {
		n, err := p.Read(data[tn:])
		if err != nil {
			log.Println(err)
			return err
		}
		tn += n
	}

	//t.assert(nl == tn)
	fmt.Printf("%7s ->  %7d/%7d [%02x %02x - %02x %02x]\n", sl, nl, tn, data[0], data[1], data[nl-2], data[nl-1])

	return err
}

// ---------------------------------E-----N-----D--------------------------------
