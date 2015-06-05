//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for WebSocket(WS, WSS)
// - https://godoc.org/golang.org/x/net/websocket
// - https://github.com/golang-samples/websocket
// - http://www.ajanicij.info/content/websocket-tutorial-go
//==================================================================================

package protows

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	sr "stoney/httpserver/src/streamring"

	"golang.org/x/net/websocket"
)

//---------------------------------------------------------------------------
const (
	KBYTE = 1024
	MBYTE = 1024 * KBYTE

	LEN_MAX_LINE = 128

	STR_DEF_HOST = "localhost"
	STR_DEF_PORT = "8080"
	STR_DEF_PTLS = "8443"
	STR_DEF_BDRY = "myboundary"
)

//---------------------------------------------------------------------------
type WsConfig struct {
	Ws       *websocket.Conn
	Boundary string
	Mr       *multipart.Reader
	Mw       *multipart.Writer
}

type DataFrame struct {
	Type string
	Size int
	Data []byte
}

type ProtoWs struct {
	Host     string
	Port     string
	PortTls  string // for TLS
	Desc     string
	Boundary string
	Conn     *websocket.Conn
}

//---------------------------------------------------------------------------
// WebSocket shooter for test and debugging
//---------------------------------------------------------------------------
func (pw *ProtoWs) String() string {
	str := fmt.Sprintf("\tHost: %s", pw.Host)
	str += fmt.Sprintf("\tPort: %s,%s", pw.Port, pw.PortTls)
	str += fmt.Sprintf("\tConn: %v", pw.Conn)
	str += fmt.Sprintf("\tBoundary: %s", pw.Boundary)
	str += fmt.Sprintf("\tDesc: %s", pw.Desc)
	return str
}

func (pw *ProtoWs) SetAddr(hname, hport, hptls, desc string) {
	pw.Host = hname
	pw.Port = hport
	pw.PortTls = hptls
	pw.Desc = desc
}

func (pw *ProtoWs) Reset() {
	pw.Host = STR_DEF_HOST
	pw.Port = STR_DEF_PORT
	pw.PortTls = STR_DEF_PTLS
	pw.Boundary = STR_DEF_BDRY
	pw.Desc = "reset"
	if pw.Conn != nil {
		pw.Conn.Close()
		pw.Conn = nil
	}
}

func (pw *ProtoWs) Clear() {
	pw.Host = ""
	pw.Port = ""
	pw.Desc = ""
	if pw.Conn != nil {
		pw.Conn.Close()
		pw.Conn = nil
	}
}

//---------------------------------------------------------------------------
// new ProtoWs struct
//---------------------------------------------------------------------------
func NewProtoWs(hname, hport, hptls, desc string) *ProtoWs {
	return &ProtoWs{
		Host:     hname,
		Port:     hport,
		PortTls:  hptls,
		Desc:     desc,
		Boundary: STR_DEF_BDRY,
	}
}

//---------------------------------------------------------------------------
// WebSocket shooter for test and debugging
//---------------------------------------------------------------------------
func ActWsShooter(pw *ProtoWs) {
	log.Printf("Happy Media WS Shooter\n")

	origin := fmt.Sprintf("http://%s/", pw.Host)
	url := fmt.Sprintf("ws://%s:%s/stream", pw.Host, pw.Port)

	// connect to the server
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Println(err)
		return
	}

	err = pw.SummitRequest(ws)
	if err != nil {
		log.Println(err)
		return
	}

	pw.SendMultipartFiles(ws, "static/image/*.jpg")
}

//---------------------------------------------------------------------------
// WebSocket catcher for test and debugging
//---------------------------------------------------------------------------
func ActWsCatcher(pw *ProtoWs) {
	log.Printf("Happy Media WS Catcher on ws:%s and wss:%s\n", pw.Port, pw.PortTls)

	http.Handle("/echo", websocket.Handler(pw.EchoHandler))
	http.Handle("/stream", websocket.Handler(pw.StreamHandler))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	//log.Fatal(http.ListenAndServe(":"+pw.Port, nil))

	//var wg sync.WaitGroup
	wg := sync.WaitGroup{}

	wg.Add(1)
	// HTTP server
	go pw.serveHttp(&wg)

	wg.Add(1)
	// HTTPS server
	go pw.serveHttps(&wg)

	wg.Wait()
}

//---------------------------------------------------------------------------
// for http access
//---------------------------------------------------------------------------
func (pw *ProtoWs) serveHttp(wg *sync.WaitGroup) {
	log.Println("Starting HTTP server at http://" + pw.Host + ":" + pw.Port)
	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + pw.Port,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

//---------------------------------------------------------------------------
// for https tls access
//---------------------------------------------------------------------------
func (pw *ProtoWs) serveHttps(wg *sync.WaitGroup) {
	log.Println("Starting HTTPS server at https://" + pw.Host + ":" + pw.PortTls)
	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + pw.PortTls,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServeTLS("sec/cert.pem", "sec/key.pem"))
}

//---------------------------------------------------------------------------
// WebSocket stream handler in the server
// - http://www.websocket.org/echo.html
// - http://jan.newmarch.name/go/websockets/chapter-websockets.html
//---------------------------------------------------------------------------
func (pw *ProtoWs) EchoHandler(ws *websocket.Conn) {
	var err error

	// simplest echo function
	//io.Copy(ws, ws)

	// slightly message change echo
	for {
		var reply string

		err = websocket.Message.Receive(ws, &reply)
		if err != nil {
			log.Println(err)
			break
		}

		fmt.Println("Received back from client: " + reply)

		msg := "Received:  " + reply
		fmt.Println("Sending to client: " + msg)

		err = websocket.Message.Send(ws, msg)
		if err != nil {
			log.Println(err)
			break
		}
	}
}

//---------------------------------------------------------------------------
// WebSocket stream handler in the server
//---------------------------------------------------------------------------
func (pw *ProtoWs) StreamHandler(ws *websocket.Conn) {
	log.Printf("in %s\n", ws.RemoteAddr())
	defer log.Printf("out %s\n", ws.RemoteAddr())

	err := pw.HandleRequest(ws)
	if err != nil {
		log.Println(err)
		return
	}

	err = pw.RecvMultipart(ws)
	if err != nil {
		log.Println(err)
		return
	}
}

//---------------------------------------------------------------------------
// WebSocket summit a request to the server
//---------------------------------------------------------------------------
func (pw *ProtoWs) SummitRequest(ws *websocket.Conn) error {
	var err error

	// send POST request
	smsg := "POST /stream HTTP/1.1\r\n"
	smsg += fmt.Sprintf("Content-Type: multipart/x-mixed-replace; boundary=%s\r\n", pw.Boundary)
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
// WebSocket handle a client request in the server
//---------------------------------------------------------------------------
func (pw *ProtoWs) HandleRequest(ws *websocket.Conn) error {
	var err error

	// recv POST request
	rmsg := make([]byte, 512)

	n, err := ws.Read(rmsg)
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Printf("Recv(%d):\n%s", n, rmsg[:n])

	// parse request
	err = pw.GetBoundary(string(rmsg[:n]))
	if err != nil {
		log.Println(err)
		return err
	}

	// send response
	smsg := "HTTP/1.1 200 Ok\r\n"
	smsg += "Server: Happy Media WS Server\r\n"
	smsg += "\r\n"

	n, err = ws.Write([]byte(smsg))
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Printf("Send(%d):\n%s", n, smsg)

	return err
}

//---------------------------------------------------------------------------
// WebSocket get boundary string from message
//---------------------------------------------------------------------------
func (pw *ProtoWs) GetBoundary(msg string) error {
	var err error

	req, err := pw.GetRequest(msg)

	mt, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	fmt.Printf("%v %v\n", params, req.Header.Get("Content-Type"))
	if err != nil {
		log.Printf("ParseMediaType: %s %v", mt, err)
		return err
	}

	boundary := params["boundary"]
	if !strings.HasPrefix(boundary, "--") {
		log.Printf("expected boundary to start with --, got %q", boundary)
	}

	pw.Boundary = boundary

	return err
}

//---------------------------------------------------------------------------
// WebSocket get http request from message
//---------------------------------------------------------------------------
func (pw *ProtoWs) GetRequest(msg string) (*http.Request, error) {
	var err error

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
// WebSocket get header from message
//---------------------------------------------------------------------------
func (pw *ProtoWs) GetHeader(msg string) (http.Header, error) {
	var err error

	reader := bufio.NewReader(strings.NewReader(msg))
	tp := textproto.NewReader(reader)

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	httpHeader := http.Header(mimeHeader)
	fmt.Println(httpHeader)

	return httpHeader, err
}

//---------------------------------------------------------------------------
// WebSocket send files in multipart
//---------------------------------------------------------------------------
func (pw *ProtoWs) SendMultipartFiles(ws *websocket.Conn, pattern string) error {
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
	mw.SetBoundary(pw.Boundary)

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

		err = pw.SendPartData(mw, fdata, len(fdata), ctype)
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
func (pw *ProtoWs) RecvMultipart(ws *websocket.Conn) error {
	var err error

	mr := multipart.NewReader(ws, pw.Boundary)

	for {
		err = pw.RecvPartToData(mr)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return err
}

//---------------------------------------------------------------------------
// WebSocket send a part of multipart
//---------------------------------------------------------------------------
func (pw *ProtoWs) SendPartData(mw *multipart.Writer, data []byte, dsize int, dtype string) error {
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
// WebSocket recv a part of multipart
//---------------------------------------------------------------------------
func (pw *ProtoWs) RecvPartToData(mr *multipart.Reader) error {
	var err error

	p, nl, err := pw.ReadPartHeader(mr)
	if err != nil {
		log.Println(err)
		return err
	}

	data := make([]byte, nl)

	err = pw.ReadPartBodyToData(p, nl, data)
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Printf("%d [%0x-%0x]\n", nl, data[2], data[nl-2:])
	return err
}

//---------------------------------------------------------------------------
// read a part header and parse it
//---------------------------------------------------------------------------
func (pw *ProtoWs) ReadPartHeader(mr *multipart.Reader) (*multipart.Part, int, error) {
	var err error

	p, err := mr.NextPart()
	if err != nil { // io.EOF
		log.Println(err)
		return p, 0, err
	}

	sl := p.Header.Get("Content-Length")
	nl, err := strconv.Atoi(sl)
	if err != nil {
		log.Printf("%s %s -> %d\n", p.Header, sl, nl)
		return p, nl, err
	}

	return p, nl, err
}

//---------------------------------------------------------------------------
// read a part body to data
//---------------------------------------------------------------------------
func (pw *ProtoWs) ReadPartBodyToData(p *multipart.Part, nl int, data []byte) error {
	var err error

	var tn int
	for tn < nl {
		n, err := p.Read(data[tn:])
		if err != nil {
			log.Println(err)
			return err
		}
		tn += n
	}

	return err
}

//---------------------------------------------------------------------------
// read a part body to slot
//---------------------------------------------------------------------------
func (pw *ProtoWs) ReadPartBodyToSlot(p *multipart.Part, nl int, ss *sr.StreamSlot) error {
	var err error

	ss.Length = 0

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

	return err
}

// ---------------------------------E-----N-----D--------------------------------
