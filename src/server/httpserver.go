//=================================================================================
// Happy Media System
// one program including agents such as caster, server, player, monitor
// Author : Stoney Kang, sikang99@gmail.com
//=================================================================================
package main

import (
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
	pt "stoney/httpserver/src/prototcp"
	pw "stoney/httpserver/src/protows"
	si "stoney/httpserver/src/streamimage"
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

	conf.Ring = prepareStreamRing(3, MBYTE, "Server stream")

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
		ts := pt.NewProtoTcp("localhost", "8087", "T-Tx")
		ts.ActSender()
	case "receiver":
		tr := pt.NewProtoTcp("localhost", "8087", "T-Rx")
		tr.ActReceiver(conf.Ring)
	case "shooter":
		ws := pw.NewProtoWs("localhost", "8087", "8443", "W-Tx")
		ws.ActShooter()
	case "catcher":
		wr := pw.NewProtoWs("localhost", "8087", "8443", "W-Rx")
		wr.ActCatcher()
	case "filer":
		ActFileReader()
	default:
		fmt.Println("Unknown working mode")
		os.Exit(0)
	}
}

//==================================================================================
// HTTP
// - http://www.sanarias.com/blog/1214PlayingwithimagesinHTTPresponseingolang
//==================================================================================
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

	fdata, err := ioutil.ReadFile("static/image/gopher.jpg")
	fsize := len(fdata)

	for i := 0; i < 3; i++ {
		part, _ := mw.CreatePart(textproto.MIMEHeader{
			"Content-Type":   {"image/jpeg"},
			"Content-Length": {strconv.Itoa(fsize)},
		})

		b := new(bytes.Buffer)
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

	//recvMultipartToData(mr)
	recvMultipartToStreamRing(mr)
}

//---------------------------------------------------------------------------
//	multipart reader entry, mainly from camera
//---------------------------------------------------------------------------
func ActFileReader() {
	log.Printf("Happy Media File Reader\n")

	readDirToStreamRing("./static/image/*.jpg", true)
	fmt.Println(conf.Ring)
}

//---------------------------------------------------------------------------
//	receive a part to data
//---------------------------------------------------------------------------
func recvPartToData(r *multipart.Reader) ([]byte, error) {
	var err error

	p, err := r.NextPart()
	if err != nil { // io.EOF
		log.Println(err)
		return nil, err
	}

	sl := p.Header.Get("Content-Length")
	nl, err := strconv.Atoi(sl)
	if err != nil {
		log.Printf("%s %s %d\n", p.Header, sl, nl)
		return nil, err
	}

	data := make([]byte, nl)

	// implement like ReadFull() in jpeg.Decode()
	var tn int
	for tn < nl {
		n, err := p.Read(data[tn:])
		if err != nil {
			log.Println(err)
			return nil, err
		}
		tn += n
	}

	fmt.Printf("%s %d/%d [%0x - %0x]\n", p.Header.Get("Content-Type"), tn, nl, data[:2], data[nl-2:])

	return data[:nl], err
}

//---------------------------------------------------------------------------
//	receive a part to slot of buffer
//----------------------------------------s-----------------------------------
func recvPartToSlot(r *multipart.Reader, ss *sr.StreamSlot) error {
	var err error

	p, err := r.NextPart()
	if err != nil { // io.EOF
		log.Println(err)
		return err
	}

	sl := p.Header.Get("Content-Length")
	nl, err := strconv.Atoi(sl)
	if err != nil {
		log.Printf("%s %s %d\n", p.Header, sl, nl)
		return err
	}

	var tn int
	for tn < nl {
		n, err := p.Read(ss.Content[tn:])
		if err != nil {
			log.Println(err)
			return err
		}
		tn += n
	}

	ss.Length = nl
	ss.Type = p.Header.Get("Content-Type")
	//fmt.Println(ss)

	return err
}

//---------------------------------------------------------------------------
//	receive multipart data into buffer
//---------------------------------------------------------------------------
func recvMultipartToStreamRing(r *multipart.Reader) error {
	var err error

	//sbuf := prepareStreamRing(5, MBYTE, "AXIS Camera")
	sbuf := conf.Ring
	err = sbuf.SetStatusUsing()
	if err != nil {
		return sr.ErrStatus
	}

	// insert slots to the buffer
	for i := 0; true; i++ {
		//pre, pos := sbuf.ReadSlotIn()
		//fmt.Println("P", pos, pre)

		slot, pos := sbuf.GetSlotIn()
		//fmt.Println(i, pos, slot)

		err = recvPartToSlot(r, slot)
		if err != nil {
			log.Println(err)
			break
		}
		fmt.Println(i, pos, slot)

		sbuf.SetPosInByPos(pos + 1)
	}

	return err
}

//---------------------------------------------------------------------------
//	receive multipart data and decode jpeg
//---------------------------------------------------------------------------
func recvMultipartToData(r *multipart.Reader) error {
	var err error

	for {
		p, err := r.NextPart()
		if err != nil { // io.EOF
			log.Println(err)
			return err
		}

		sl := p.Header.Get("Content-Length")
		nl, err := strconv.Atoi(sl)
		if err != nil {
			log.Printf("%s %s %d\n", p.Header, sl, nl)
			continue
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

		fmt.Printf("%s %d/%d [%0x - %0x]\n", p.Header.Get("Content-Type"), tn, nl, data[:2], data[nl-2:])
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
		if err != nil { // io.EOF
			log.Println(err)
		}

		_, err = jpeg.Decode(p)
		if err != nil {
			log.Println(err)
		}
	}
}

//---------------------------------------------------------------------------
// prepare a stream buffer
//---------------------------------------------------------------------------
func prepareStreamRing(nb int, size int, desc string) *sr.StreamRing {

	sbuf := sr.NewStreamRing(nb, size)
	sbuf.Desc = desc
	fmt.Println(sbuf)

	return sbuf
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

	/*
		wg.Add(1)
		// WS server
		go serveWs(&wg)

		wg.Add(1)
		// WSS server
		go serveWss(&wg)
	*/

	//go ActFileReader()
	//go ActHttpReader("http://imoment:imoment@192.168.0.91/axis-cgi/mjpg/video.cgi")

	tr := pt.NewProtoTcp("localhost", "8087", "T-Rx")
	go tr.ActReceiver(conf.Ring)

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
// send image stream with the given format(extension) in the directory
//---------------------------------------------------------------------------
func sendStreamImage(w io.Writer, dtype string, loop bool) error {
	var err error

	for {
		err = sendPartImage(w, dtype)
		if err != nil {
			log.Println(err)
			break
		}

		time.Sleep(time.Second)

		if !loop {
			break
		}
	}

	return err
}

//---------------------------------------------------------------------------
// send image stream with the given format(extension) in the directory
//---------------------------------------------------------------------------
func sendStreamRing(w io.Writer) error {
	var err error

	sbuf := conf.Ring
	if !sbuf.IsUsing() {
		return sr.ErrStatus
	}

	var pos int
	for {
		slot, npos, err := sbuf.GetSlotNextByPos(pos)
		if slot == nil && err == sr.ErrEmpty {
			time.Sleep(time.Millisecond)
			continue
		}

		err = sendPartSlot(w, slot)
		if err != nil {
			log.Println(err)
			break
		}

		pos = npos
	}

	return err
}

//---------------------------------------------------------------------------
// send files with the given format(extension) in the directory
//---------------------------------------------------------------------------
func sendStreamDir(w io.Writer, pat string, loop bool) error {
	var err error

	// direct pattern matching
	files, err := filepath.Glob(pat)
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
			err = sendPartFile(w, files[i])
			if err != nil {
				return err
			}
			time.Sleep(time.Second)
		}

		if !loop {
			break
		}
	}

	return err
}

//---------------------------------------------------------------------------
// read files with the given pattern in the directory and put them to the buffer
//---------------------------------------------------------------------------
func readDirToStreamRing(pat string, loop bool) error {
	var err error

	// direct pattern matching
	files, err := filepath.Glob(pat)
	if err != nil {
		log.Println(err)
		return err
	}
	if files == nil {
		log.Printf("no matched file for %s\n", pat)
		return err
	}

	// use server stream buffer
	sbuf := conf.Ring
	err = sbuf.SetStatusUsing()
	if err != nil {
		return sr.ErrStatus
	}

	for {
		for i := range files {
			slot, pos := sbuf.GetSlotIn()
			err = readFileToSlot(files[i], slot)
			if err != nil {
				if err == sr.ErrEmpty {
					continue
				}
				log.Println(err)
				return err
			}
			fmt.Println(slot)
			sbuf.SetPosInByPos(pos + 1)
			time.Sleep(time.Second)
		}

		if !loop {
			break
		}
	}

	return err
}

//---------------------------------------------------------------------------
// read a file to stream buffer
//---------------------------------------------------------------------------
func readFileToSlot(file string, ss *sr.StreamSlot) error {
	var err error

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println(err)
		return err
	}
	dsize := len(data)

	if dsize == 0 {
		log.Printf("%s(%d) is empty\n", file, dsize)
		return sr.ErrEmpty
	}

	ctype := mime.TypeByExtension(filepath.Ext(file))
	if ctype == "" {
		ctype = http.DetectContentType(data)
	}

	copy(ss.Content, data)
	ss.Length = dsize
	ss.Type = ctype

	return err
}

//---------------------------------------------------------------------------
// send any file in multipart format with boundary (standard style)
//---------------------------------------------------------------------------
func sendPartFile(w io.Writer, file string) error {
	var err error

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println(err)
		return err
	}
	dsize := len(data)

	// check the extension at first and its content
	ctype := mime.TypeByExtension(filepath.Ext(file))
	if ctype == "" {
		ctype = http.DetectContentType(data)
	}

	mw := multipart.NewWriter(w)
	mw.SetBoundary("myboundary")

	part, err := mw.CreatePart(textproto.MIMEHeader{
		"Content-Type":   {ctype},
		"Content-Length": {strconv.Itoa(dsize)},
	})
	if err != nil {
		log.Println(err)
		return err
	}

	buf := new(bytes.Buffer)
	buf.Write(data)   // prepare data in the buffer
	buf.WriteTo(part) // output the part with buffer in multipart format

	return err
}

//---------------------------------------------------------------------------
// send image in multipart format with boundary (standard style)
//---------------------------------------------------------------------------
func sendPartImage(w io.Writer, dtype string) error {
	var err error

	//img := si.GenSpiralImage(1080, 768)
	//img := si.GenClockImage(800)
	img := si.GenRandomImageBlock(1080, 768)
	data, err := si.PutImageToBuffer(img, dtype, 90)
	if err != nil {
		log.Println(err)
		return err
	}
	dsize := len(data)

	ctype := http.DetectContentType(data)

	mw := multipart.NewWriter(w)
	mw.SetBoundary("myboundary")
	part, err := mw.CreatePart(textproto.MIMEHeader{
		"Content-Type":   {ctype},
		"Content-Length": {strconv.Itoa(dsize)},
	})
	if err != nil {
		log.Println(err)
		return err
	}

	buf := new(bytes.Buffer)
	buf.Write(data)
	buf.WriteTo(part)

	return err
}

//---------------------------------------------------------------------------
// send slot in multipart format with boundary (standard style)
//---------------------------------------------------------------------------
func sendPartSlot(w io.Writer, ss *sr.StreamSlot) error {
	var err error

	mw := multipart.NewWriter(w)
	mw.SetBoundary("myboundary")
	part, err := mw.CreatePart(textproto.MIMEHeader{
		"Content-Type":   {ss.Type},
		"Content-Length": {strconv.Itoa(ss.Length)},
	})
	if err != nil {
		log.Println(err)
		return err
	}

	buf := new(bytes.Buffer)
	buf.Write(ss.Content[:ss.Length])
	buf.WriteTo(part)

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

		recvMultipartToData(mr)

	case "GET": // for Player
		err := sendStreamResponse(w)
		if err != nil {
			log.Println(err)
			break
		}

		//err = sendStreamDir(w, "./static/image/*.jpg", true)
		//err = sendStreamImage(w, "jpg", true)
		err = sendStreamRing(w)
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
// handle /search access
//---------------------------------------------------------------------------
func searchHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Search %s for %s to %s\n", r.Method, r.URL.Path, r.Host)

	Responder(w, r, 200, "/search: Not yet implemented")

	return
}

//---------------------------------------------------------------------------
// handle /search access
//---------------------------------------------------------------------------
func statusHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Search %s for %s to %s\n", r.Method, r.URL.Path, r.Host)

	Responder(w, r, 200, "/status: Not yet implemented")
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

// ---------------------------------E-----N-----D--------------------------------
