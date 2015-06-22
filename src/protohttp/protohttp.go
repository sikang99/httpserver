//=========================================================================
// Author : Stoney Kang, sikang99@gmail.com, 2015
// Protocol for HTTP streaming
// - http://www.sanarias.com/blog/1214PlayingwithimagesinHTTPresponseingolang
// - http://stackoverflow.com/questions/30552447/how-to-set-which-ip-to-use-for-a-http-request
// - http://giantmachines.tumblr.com/post/52184842286/golang-http-client-with-timeouts
//=========================================================================

package protohttp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"golang.org/x/net/websocket"

	"github.com/bradfitz/http2"

	sb "stoney/httpserver/src/streambase"
	si "stoney/httpserver/src/streamimage"
	sr "stoney/httpserver/src/streamring"
)

//---------------------------------------------------------------------------
const (
	STR_HTTP_CASTER  = "Happy Media HTTP Caster"
	STR_HTTP_SERVER  = "Happy Media HTTP Server"
	STR_HTTP_PLAYER  = "Happy Media HTTP Player"
	STR_HTTP_MONITOR = "Happy Media HTTP Monitor"
	STR_HTTP_READER  = "Happy Media HTTP Reader"
	STR_HTTP_WRITER  = "Happy Media HTTP Writer"
)

type ProtoHttp struct {
	Url      string
	Host     string
	Port     string
	PortTls  string
	Port2    string
	Method   string
	Boundary string
	Desc     string
}

//---------------------------------------------------------------------------
// string information for package struct
//---------------------------------------------------------------------------
func (ph *ProtoHttp) String() string {
	str := fmt.Sprintf("\tUrl: %s", ph.Url)
	str += fmt.Sprintf("\tHost: %s", ph.Host)
	str += fmt.Sprintf("\tPort: %s,%s,%s", ph.Port, ph.PortTls, ph.Port2)
	str += fmt.Sprintf("\tBoundary: %s", ph.Boundary)
	str += fmt.Sprintf("\tDesc: %s", ph.Desc)
	return str
}

//---------------------------------------------------------------------------
// make a new struct
//---------------------------------------------------------------------------
func NewProtoHttp(args ...string) *ProtoHttp {
	ph := &ProtoHttp{
		Host:     sb.STR_DEF_HOST,
		Port:     sb.STR_DEF_PORT,
		PortTls:  sb.STR_DEF_PTLS,
		Port2:    sb.STR_DEF_PORT2,
		Boundary: sb.STR_DEF_BDRY,
	}

	ph.Desc = "NewProtoHttp"

	for i, arg := range args {
		switch {
		case i == 0:
			ph.Boundary = arg
		case i == 1:
			ph.Desc = arg
		}
	}

	return ph
}

func NewProtoHttpWithPorts(args ...string) *ProtoHttp {
	ph := NewProtoHttp()

	ph.Desc = "NewProtoHttpWithPorts"

	for i, arg := range args {
		switch {
		case i == 0:
			ph.Port = arg
		case i == 1:
			ph.PortTls = arg
		case i == 2:
			ph.Port2 = arg
		}
	}
	return ph
}

//---------------------------------------------------------------------------
// http monitor client
//---------------------------------------------------------------------------
func (ph *ProtoHttp) StreamMonitor(url string) error {
	log.Printf("%s for %s\n", STR_HTTP_MONITOR, url)

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
// parse command
//---------------------------------------------------------------------------
func ParseCommand(cmdstr string) error {
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
//	multipart reader entry, mainly from camera
//---------------------------------------------------------------------------
func (ph *ProtoHttp) StreamReader(url string, ring *sr.StreamRing) {
	log.Printf("%s for %s\n", STR_HTTP_READER, url)

	var err error
	var res *http.Response

	// WHY: different behavior?
	if strings.Contains(url, "axis") {
		res, err = http.Get(url)
	} else {
		client := NewClientConfig()
		res, err = client.Get(url)
	}
	if err != nil {
		log.Fatalf("GET of %q: %v", url, err)
	}
	//log.Printf("Content-Type: %v", res.Header.Get("Content-Type"))

	ring.Boundary, err = GetTypeBoundary(res.Header.Get("Content-Type"))
	mr := multipart.NewReader(res.Body, ring.Boundary)

	err = ReadMultipartToRing(mr, ring)
}

//---------------------------------------------------------------------------
// http caster client
//---------------------------------------------------------------------------
func (ph *ProtoHttp) StreamCaster(url string) error {
	log.Printf("%s for %s\n", STR_HTTP_CASTER, url)

	hp := NewProtoHttp()
	client := NewClientConfig()
	return hp.RequestPost(client)
}

//---------------------------------------------------------------------------
// http player client
//---------------------------------------------------------------------------
func (ph *ProtoHttp) StreamPlayer(url string, ring *sr.StreamRing) error {
	log.Printf("%s for %s\n", STR_HTTP_PLAYER, url)

	var err error
	var res *http.Response

	// WHY: different behavior?
	if strings.Contains(url, "axis") {
		res, err = http.Get(url)
	} else {
		client := NewClientConfig()
		res, err = client.Get(url)
	}
	if err != nil {
		log.Fatalf("GET of %q: %v", url, err)
	}
	//log.Printf("Content-Type: %v", res.Header.Get("Content-Type"))

	ring.Boundary, err = GetTypeBoundary(res.Header.Get("Content-Type"))
	mr := multipart.NewReader(res.Body, ring.Boundary)

	err = ReadMultipartToRing(mr, ring)

	return err
}

//---------------------------------------------------------------------------
// http server entry
//---------------------------------------------------------------------------
func (ph *ProtoHttp) StreamServer(ring *sr.StreamRing) error {
	log.Printf("%s\n", STR_HTTP_SERVER)

	http.HandleFunc("/", ph.IndexHandler)
	http.HandleFunc("/hello", ph.HelloHandler)   // view
	http.HandleFunc("/media", ph.MediaHandler)   // on-demand
	http.HandleFunc("/stream", ph.StreamHandler) // live
	http.HandleFunc("/search", ph.SearchHandler) // server info
	http.HandleFunc("/status", ph.StatusHandler) // server status

	http.Handle("/websocket/", websocket.Handler(ph.WebsocketHandler))

	// CAUTION: don't use /static not /static/ as the prefix
	http.Handle("/static/", http.StripPrefix("/static/", FileServer("./static")))

	//var wg sync.WaitGroup
	wg := sync.WaitGroup{}

	wg.Add(1)
	// HTTP server
	go ph.ServeHttp(&wg)

	wg.Add(1)
	// HTTPS server
	go ph.ServeHttps(&wg)

	wg.Add(1)
	// HTTP2 server
	go ph.ServeHttp2(&wg)

	/*
		wg.Add(1)
		// WS server
		go ph.ServeWs(&wg)

		wg.Add(1)
		// WSS server
		go ph.ServeWss(&wg)
	*/

	go ph.StreamReader("http://imoment:imoment@192.168.0.91/axis-cgi/mjpg/video.cgi", ring)
	//go pt.NewProtoTcp("localhost", "8087", "T-Rx").StreamServer(ring)
	//go pf.NewProtoFile("./static/image/*.jpg", "F-Rx").StreamCaster(ring)

	wg.Wait()

	return nil
}

//---------------------------------------------------------------------------
// index file handler
//---------------------------------------------------------------------------
func (ph *ProtoHttp) IndexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Index %s to %s\n", r.URL.Path, r.Host)

	if strings.Contains(r.URL.Path, "favicon.ico") {
		http.ServeFile(w, r, "static/favicon.ico")
		return
	}

	WriteTemplatePage(w, index_tmpl, conf)
}

//---------------------------------------------------------------------------
// hello file handler (default: hello.html)
//---------------------------------------------------------------------------
func (ph *ProtoHttp) HelloHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Hello %s to %s\n", r.URL.Path, r.Host)

	hello_page := "static/hello.html"

	host := strings.Split(r.Host, ":")
	conf.Port = host[1]

	if r.TLS != nil {
		conf.Addr = "https://localhost"
	}

	_, err := os.Stat(hello_page)
	if err != nil {
		WriteTemplatePage(w, hello_tmpl, conf)
		log.Printf("Hello serve %s\n", "hello_tmpl")
	} else {
		http.ServeFile(w, r, hello_page)
		log.Printf("Hello serve %s\n", hello_page)
	}
}

//---------------------------------------------------------------------------
// media file handler
//---------------------------------------------------------------------------
func (ph *ProtoHttp) MediaHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Media %s to %s\n", r.URL.Path, r.Host)

	_, err := os.Stat(r.URL.Path[1:])
	if err != nil {
		WriteResponseMessage(w, 404, r.URL.Path+" is Not Found")
	} else {
		WriteResponseMessage(w, 200, r.URL.Path)
	}
}

//---------------------------------------------------------------------------
// media file handler
//---------------------------------------------------------------------------
func (ph *ProtoHttp) WebsocketHandler(ws *websocket.Conn) {
	log.Printf("Websocket \n")

	err := websocket.Message.Send(ws, "Not yet implemented")
	if err != nil {
		log.Println(err)
	}
}

//---------------------------------------------------------------------------
// handle /search access
//---------------------------------------------------------------------------
func (ph *ProtoHttp) SearchHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Search %s for %s to %s\n", r.Method, r.URL.Path, r.Host)

	err := WriteResponseMessage(w, 200, "/search: Not yet implemented")
	if err != nil {
		log.Println(err)
	}

	return
}

//---------------------------------------------------------------------------
// handle /search access
//---------------------------------------------------------------------------
func (ph *ProtoHttp) StatusHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Search %s for %s to %s\n", r.Method, r.URL.Path, r.Host)

	err := WriteResponseMessage(w, 200, "/status: Not yet implemented")
	if err != nil {
		log.Println(err)
	}
}

//---------------------------------------------------------------------------
// handle /stream access
//---------------------------------------------------------------------------
func (ph *ProtoHttp) StreamHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Stream %s for %s to %s\n", r.Method, r.URL.Path, r.Host)

	var err error
	ring := conf.Ring

	switch r.Method {
	case "POST": // for Caster
		ring.Boundary, err = GetTypeBoundary(r.Header.Get("Content-Type"))

		err = ResponsePost(w, ring.Boundary)
		if err != nil {
			log.Println(err)
			break
		}

		mr := multipart.NewReader(r.Body, ring.Boundary)

		err = ReadMultipartToRing(mr, ring)
		if err != nil {
			log.Println(err)
			break
		}

	case "GET": // for Player
		err = ResponseGet(w, ring.Boundary)
		if err != nil {
			log.Println(err)
			break
		}

		err = WriteRingInMultipart(w, ring)
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
func (ph *ProtoHttp) ServeHttp(wg *sync.WaitGroup) {
	log.Println("Starting HTTP server at http://" + ph.Host + ":" + ph.Port)
	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + ph.Port,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

//---------------------------------------------------------------------------
// for https tls access
//---------------------------------------------------------------------------
func (ph *ProtoHttp) ServeHttps(wg *sync.WaitGroup) {
	log.Println("Starting HTTPS server at https://" + ph.Host + ":" + ph.PortTls)
	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + ph.PortTls,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServeTLS("sec/cert.pem", "sec/key.pem"))
}

//---------------------------------------------------------------------------
// for http2 tls access
//---------------------------------------------------------------------------
func (ph *ProtoHttp) ServeHttp2(wg *sync.WaitGroup) {
	log.Println("Starting HTTP2 server at https://" + ph.Host + ":" + ph.Port2)
	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + ph.Port2,
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
func (ph *ProtoHttp) ServeWs(wg *sync.WaitGroup) {
	log.Println("Starting WS server at https://" + ph.Host + ":" + ph.Port)
	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + ph.Port,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

//---------------------------------------------------------------------------
// for wss access
//---------------------------------------------------------------------------
func (ph *ProtoHttp) ServeWss(wg *sync.WaitGroup) {
	log.Println("Starting WSS server at https://" + ph.Host + ":" + ph.PortTls)
	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + ph.PortTls,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServeTLS("sec/cert.pem", "sec/key.pem"))
}

//---------------------------------------------------------------------------
// http GET to server
//---------------------------------------------------------------------------
func (ph *ProtoHttp) RequestGet(client *http.Client) error {
	var err error

	res, err := client.Get(ph.Url)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res, body)

	return err
}

//---------------------------------------------------------------------------
// http POST to server
//---------------------------------------------------------------------------
func (ph *ProtoHttp) RequestPost(client *http.Client) error {
	var err error

	// send multipart data
	outer := new(bytes.Buffer)

	mw := multipart.NewWriter(outer)
	mw.SetBoundary(ph.Boundary)

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
	req, err := http.NewRequest("POST", ph.Url, outer)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "multipart/x-mixed-replace; boundary=--"+ph.Boundary)

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	fmt.Println(res.Status)

	return err
}

//---------------------------------------------------------------------------
//	receive a part to data
//---------------------------------------------------------------------------
func ReadPartToData(mr *multipart.Reader) ([]byte, error) {
	var err error

	p, err := mr.NextPart()
	if err != nil { // io.EOF
		log.Println(err)
		return nil, err
	}

	sl := p.Header.Get(sb.STR_HDR_CONTENT_LENGTH)
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

	//fmt.Printf("%s %d/%d [%0x - %0x]\n", p.Header.Get("Content-Type"), tn, nl, data[:2], data[nl-2:])
	return data[:nl], err
}

//---------------------------------------------------------------------------
//	receive a part to slot of buffer
//---------------------------------------------------------------------------
func ReadPartToSlot(mr *multipart.Reader, slot *sr.StreamSlot) error {
	var err error

	p, err := mr.NextPart()
	if err != nil { // io.EOF
		log.Println(err)
		return err
	}

	sl := p.Header.Get(sb.STR_HDR_CONTENT_LENGTH)
	nl, err := strconv.Atoi(sl)
	if err != nil {
		log.Printf("%s %s %d\n", p.Header, sl, nl)
		return err
	}

	slot.Length = 0

	var tn int
	for tn < nl {
		n, err := p.Read(slot.Content[tn:])
		if err != nil {
			log.Println(err)
			return err
		}
		tn += n
	}

	slot.Length = nl
	slot.Type = p.Header.Get(sb.STR_HDR_CONTENT_TYPE)
	slot.Timestamp = sb.GetTimestampNow()
	//fmt.Println(slot)

	return err
}

//---------------------------------------------------------------------------
//	receive multipart data into buffer
//---------------------------------------------------------------------------
func ReadMultipartToRing(mr *multipart.Reader, ring *sr.StreamRing) error {
	var err error

	err = ring.SetStatusUsing()
	if err != nil {
		return sb.ErrStatus
	}

	// insert slots to the buffer
	for i := 0; true; i++ {
		//pre, pos := ring.ReadSlotIn()
		//fmt.Println("P", pos, pre)

		slot, pos := ring.GetSlotIn()
		//fmt.Println(i, pos, slot)

		err = ReadPartToSlot(mr, slot)
		if err != nil {
			log.Println(err)
			break
		}
		//fmt.Println(i, pos, slot)

		ring.SetPosInByPos(pos + 1)
	}

	return err
}

//---------------------------------------------------------------------------
//	receive multipart data and decode jpeg
//---------------------------------------------------------------------------
func ReadMultipartToData(mr *multipart.Reader) error {
	var err error

	for {
		p, err := mr.NextPart()
		if err != nil { // io.EOF
			log.Println(err)
			return err
		}

		sl := p.Header.Get(sb.STR_HDR_CONTENT_LENGTH)
		nl, err := strconv.Atoi(sl)
		if sl == "" || nl == 0 {
			log.Printf("%s %s %d\n", p.Header, sl, nl)
			continue
		}
		if err != nil {
			log.Println(err)
			return err
		}

		if nl > 0 {
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

			//log.Printf("%s %d/%d [%0x - %0x]\n", p.Header.Get(sb.STR_HDR_CONTENT_LENGTH), tn, nl, data[:2], data[nl-2:])
		}
	}

	return err
}

//---------------------------------------------------------------------------
// prepare a stream buffer
//---------------------------------------------------------------------------
func PrepareRing(nb int, size int, desc string) *sr.StreamRing {

	ring := sr.NewStreamRing(nb, size)
	ring.Desc = desc
	fmt.Println(ring)

	return ring
}

//---------------------------------------------------------------------------
// static file server handler
//---------------------------------------------------------------------------
func FileServer(path string) http.Handler {
	log.Println("File server for " + path)
	return http.FileServer(http.Dir(path))
}

//---------------------------------------------------------------------------
// send a favicon
//---------------------------------------------------------------------------
func WriteFavicon(w http.ResponseWriter, file string) error {
	w.Header().Set("Content-Type", "image/icon")
	w.Header().Set("Server", STR_HTTP_SERVER)
	body, err := ioutil.ReadFile("static/favicon.ico")
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Fprint(w, string(body))

	return nil
}

//---------------------------------------------------------------------------
// send a template page
//---------------------------------------------------------------------------
func WriteTemplatePage(w http.ResponseWriter, page string, data interface{}) error {
	t, err := template.New("mjpeg").Parse(page)
	if err != nil {
		log.Println(err)
		return err
	}

	return t.Execute(w, data)
}

//---------------------------------------------------------------------------
// send image stream with the given format(extension) in the directory
//---------------------------------------------------------------------------
func WriteImagesInMultipart(w io.Writer, dtype string, loop bool) error {
	var err error

	for {
		err = WriteImageInPart(w, dtype, "myboundary")
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
func WriteRingInMultipart(w io.Writer, ring *sr.StreamRing) error {
	var err error

	if !ring.IsUsing() {
		return sb.ErrStatus
	}
	fmt.Println(ring)

	var pos int
	for {
		slot, npos, err := ring.GetSlotNextByPos(pos)
		if err != nil {
			if err == sb.ErrEmpty {
				time.Sleep(sb.TIME_DEF_WAIT)
				continue
			}
			log.Println(err)
			break
		}

		err = WriteSlotInPart(w, slot, ring.Boundary)
		if err != nil {
			log.Println(err)
			break
		}
		fmt.Println(slot)

		pos = npos
	}

	return err
}

//---------------------------------------------------------------------------
// send files with the given format(extension) in the directory
//---------------------------------------------------------------------------
func WriteDirInMultipart(w io.Writer, pat string, loop bool) error {
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
			err = WriteFileInPart(w, files[i], "myboundary")
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
// send any file in multipart format with boundary (standard style)
//---------------------------------------------------------------------------
func WriteFileInPart(w io.Writer, file string, boundary string) error {
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
	mw.SetBoundary(boundary)

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
func WriteImageInPart(w io.Writer, dtype string, boundary string) error {
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
	mw.SetBoundary(boundary)
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
func WriteSlotInPart(w io.Writer, slot *sr.StreamSlot, boundary string) error {
	var err error

	mw := multipart.NewWriter(w)
	mw.SetBoundary(boundary)
	part, err := mw.CreatePart(textproto.MIMEHeader{
		"Content-Type":   {slot.Type},
		"Content-Length": {strconv.Itoa(slot.Length)},
	})
	if err != nil {
		log.Println(err)
		return err
	}

	buf := new(bytes.Buffer)
	_, err = buf.Write(slot.Content[:slot.Length])
	_, err = buf.WriteTo(part)

	return err
}

//---------------------------------------------------------------------------
// get boundary string
//---------------------------------------------------------------------------
func GetTypeBoundary(ctype string) (string, error) {
	var err error

	mt, params, err := mime.ParseMediaType(ctype)
	//fmt.Printf("%v %v\n", params, ctype)
	if err != nil {
		log.Println("ParseMediaType: %s %v", mt, err)
		return "", err
	}

	boundary := params["boundary"]
	if !strings.HasPrefix(boundary, "--") {
		log.Printf("expected to start with --, got %q", boundary)
	}

	return boundary, err
}

//---------------------------------------------------------------------------
// send response for Player
//---------------------------------------------------------------------------
func ResponseGet(w http.ResponseWriter, boundary string) error {
	var err error

	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=--"+boundary)
	w.Header().Set("Server", STR_HTTP_SERVER)
	w.WriteHeader(http.StatusOK)

	return err
}

//---------------------------------------------------------------------------
// send response for Caster
//---------------------------------------------------------------------------
func ResponsePost(w http.ResponseWriter, boundary string) error {
	var err error

	w.Header().Set("Server", STR_HTTP_SERVER)
	w.WriteHeader(http.StatusOK)

	return err
}

//---------------------------------------------------------------------------
// send response message simply
//---------------------------------------------------------------------------
func WriteResponseMessage(w http.ResponseWriter, status int, message string) error {
	var err error

	w.WriteHeader(status)
	log.Println(message)
	fmt.Fprintf(w, message)

	return err
}

// ---------------------------------E-----N-----D--------------------------------
