// ---------------------------------------------------------------------------------
// one program including agents such as caster, server, player, monitor
// ---------------------------------------------------------------------------------
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
      host : 'https://localhost',
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
var (
	NotSupportError = errors.New("Not supported protocol")

	fmode  = flag.String("m", "player", "Working mode of program")
	fhost  = flag.String("host", "localhost", "server host address")
	fport  = flag.String("port", "8000", "TCP port to be used for http")
	fports = flag.String("ports", "8001", "TCP port to be used for https")
	fport2 = flag.String("port2", "8002", "TCP port to be used for http2")
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

type ServerConfig struct {
	Title        string
	Image        string
	Host         string
	Port         string
	Mode         string
	ImageChannel chan []byte
	Player       io.Writer
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		ImageChannel: make(chan []byte, 2)}
}

var conf = ServerConfig{
	Title:        "Simple MJPEG Proxy Server",
	Image:        "static/image/gophergun.jpg",
	Host:         *fhost,
	Port:         *fport,
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
	fmt.Printf("Default config: %s %s\n", url.Scheme, url.Host)

	// determine the working type of the program
	fmt.Printf("Working mode : %s\n", *fmode)

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
	base.ShowNetInterfaces()
	return nil
}

//---------------------------------------------------------------------------
// http player client
//---------------------------------------------------------------------------
func ActHttpPlayer(url string) error {
	log.Printf("httpPlayer %s\n", url)

	client := httpClientConfig()
	return httpClientGet(client, url)
}

//---------------------------------------------------------------------------
// http caster client
//---------------------------------------------------------------------------
func ActHttpCaster(url string) error {
	log.Printf("httpCaster %s\n", url)

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
// http POST to server
//---------------------------------------------------------------------------
func httpClientPost(client *http.Client, url string) error {
	b := new(bytes.Buffer)
	w := multipart.NewWriter(b)
	w.SetBoundary("myboundary")

	req, err := http.NewRequest("POST", url, b)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "multipart/x-mixed-replace; boundary=--myboundary")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	fmt.Println(res.Status)

	fdata, err := ioutil.ReadFile("static/image/gopher.jpg")
	fsize := len(fdata)

	part, _ := w.CreatePart(textproto.MIMEHeader{
		"Content-Type":   {"image/jpeg"},
		"Content-Length": {strconv.Itoa(fsize)},
	})

	b.Write(fdata)
	b.WriteTo(part)

	time.Sleep(time.Second)

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

	r := multipart.NewReader(res.Body, boundary)

	//recvMultipartDecodeJpeg(r)
	recvMultipartToBuffer(r)
}

//---------------------------------------------------------------------------
//	receive multipart data into buffer
//---------------------------------------------------------------------------
func recvMultipartToBuffer(r *multipart.Reader) {
	for {
		p, err := r.NextPart()
		if err != nil {
			log.Fatalf("NextPart: %v", err)
		}

		sl := p.Header.Get("Content-Length")
		nl, err := strconv.Atoi(sl)
		if err != nil {
			log.Fatalf("%s %s %d\n", p.Header, sl, nl)
		}
		//println(nl)
		data := make([]byte, nl)

		// implement like ReadFull() in jpeg.Decode()
		var tn int
		for tn < nl {
			n, err := p.Read(data[tn:])
			if err != nil {
				log.Println(err)
				return
			}
			tn += n
		}

		fmt.Printf("%s %d/%d [%02x %02x - %02x %02x]\n", p.Header.Get("Content-Type"), tn, nl, data[0], data[1], data[nl-2], data[nl-1])
		//conf.ImageChannel <- data[:n]
	}
}

//---------------------------------------------------------------------------
//	receive multipart data and decode jpeg
//---------------------------------------------------------------------------
func recvMultipartDecodeJpeg(r *multipart.Reader) {
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

//---------------------------------------------------------------------------
// receive byte data
//---------------------------------------------------------------------------
func recvStreamData(r io.Reader) {
	b := make([]byte, 1024*1204*1204)
	n, err := r.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	println(n)
}

//---------------------------------------------------------------------------
// http server entry
//---------------------------------------------------------------------------
func ActHttpServer() error {
	log.Println("Happy Media Server mode")

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hello", helloHandler)   // view
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

//---------------------------------------------------------------------------
// for http access
//---------------------------------------------------------------------------
func serveHttp(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Starting HTTP server at http://" + *fhost + ":" + *fport)

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
	defer wg.Done()
	log.Println("Starting HTTPS server at https://" + *fhost + ":" + *fports)

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
	defer wg.Done()
	log.Println("Starting HTTP2 server at https://" + *fhost + ":" + *fport2)

	srv := &http.Server{
		Addr: ":" + *fport2,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	http2.ConfigureServer(srv, &http2.Server{})
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
	log.Println("Media %s to %s\n", r.URL.Path, r.Host)

	_, err := os.Stat(r.URL.Path[1:])
	if err != nil {
		Responder(w, r, 404, r.URL.Path+" is Not Found")
	} else {
		Responder(w, r, 200, r.URL.Path)
	}
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
// 51 send files with the given format(extension) in the directory
//---------------------------------------------------------------------------
func sendStreamDir(w io.Writer, dir, ext string, loop bool) error {
	var err error

	files, err := filepath.Glob(dir + "*" + ext)
	if err != nil {
		log.Println(err)
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
// 21 handle /stream access
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

// ------------------------------E-----N-----D--------------------------------------
