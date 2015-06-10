//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for WebSocket(WS, WSS)
// - https://godoc.org/golang.org/x/net/websocket
// - https://github.com/golang-samples/websocket
// - http://www.ajanicij.info/content/websocket-tutorial-go
// - http://www.jonathan-petitcolas.com/2015/01/27/playing-with-websockets-in-go.html
//==================================================================================

package protows

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
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	sb "stoney/httpserver/src/streambase"
	sr "stoney/httpserver/src/streamring"

	"golang.org/x/net/websocket"
)

//---------------------------------------------------------------------------
type ProtoWs struct {
	Host     string
	Port     string
	PortTls  string // for TLS
	Desc     string
	Boundary string
	Conn     *websocket.Conn
}

//---------------------------------------------------------------------------
// string information of struct
//---------------------------------------------------------------------------
func (pw *ProtoWs) String() string {
	str := fmt.Sprintf("\tHost: %s", pw.Host)
	str += fmt.Sprintf("\tPort: %s,%s", pw.Port, pw.PortTls)
	str += fmt.Sprintf("\tConn: %v", pw.Conn)
	str += fmt.Sprintf("\tBoundary: %s", pw.Boundary)
	str += fmt.Sprintf("\tDesc: %s", pw.Desc)
	return str
}

//---------------------------------------------------------------------------
// info handling
//---------------------------------------------------------------------------
func (pw *ProtoWs) SetAddr(hname, hport, hptls, desc string) {
	pw.Host = hname
	pw.Port = hport
	pw.PortTls = hptls
	pw.Desc = desc
}

func (pw *ProtoWs) Reset() {
	pw.Host = sb.STR_DEF_HOST
	pw.Port = sb.STR_DEF_PORT
	pw.PortTls = sb.STR_DEF_PTLS
	pw.Boundary = sb.STR_DEF_BDRY
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
		Boundary: sb.STR_DEF_BDRY,
	}
}

//---------------------------------------------------------------------------
// Echo client
//---------------------------------------------------------------------------
func (pw *ProtoWs) EchoClient(smsg string) {
	log.Printf("Happy Media WS Echo Client\n")

	origin := fmt.Sprintf("http://%s/", pw.Host)
	url := fmt.Sprintf("ws://%s:%s/echo", pw.Host, pw.Port)

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Println(err)
		return
	}

	//smsg := "Hello World!"
	err = websocket.Message.Send(ws, smsg)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(smsg)

	var rmsg string
	err = websocket.Message.Receive(ws, &rmsg)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(rmsg)
}

//---------------------------------------------------------------------------
// Echo server
//---------------------------------------------------------------------------
func (pw *ProtoWs) EchoServer() {
	log.Printf("Happy Media WS Echo Server\n")

	fmt.Println("TODO")
}

//---------------------------------------------------------------------------
// shooter for test and debugging
//---------------------------------------------------------------------------
func (pw *ProtoWs) ActShooter() {
	log.Printf("Happy Media WS Shooter\n")

	origin := fmt.Sprintf("http://%s/", pw.Host)
	url := fmt.Sprintf("ws://%s:%s/stream", pw.Host, pw.Port)

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()

	err = SendRequestPost(ws, pw.Boundary)
	if err != nil {
		log.Println(err)
		return
	}

	err = SendMultipartFiles(ws, "../../static/image/*.jpg", pw.Boundary)
}

//---------------------------------------------------------------------------
// catcher server
//---------------------------------------------------------------------------
func (pw *ProtoWs) ActCatcher() {
	log.Printf("Happy Media WS Catcher on ws:%s and wss:%s\n", pw.Port, pw.PortTls)

	http.Handle("/echo", websocket.Handler(pw.EchoHandler))
	http.Handle("/stream", websocket.Handler(pw.StreamHandler))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	//log.Fatal(http.ListenAndServe(":"+pw.Port, nil))

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
// start server to receive requests
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

	mode := "change"

	switch mode {
	case "direct":
		io.Copy(ws, ws)

	default: // "change"
		for {
			var reply string

			err = websocket.Message.Receive(ws, &reply)
			if err != nil {
				log.Println(err)
				break
			}
			//log.Println("Received from client: " + reply)

			smsg := "Echo " + reply
			err = websocket.Message.Send(ws, smsg)
			if err != nil {
				log.Println(err)
				break
			}
			//log.Println("Sending to client: " + smsg)
		}
	}
}

//---------------------------------------------------------------------------
// WebSocket stream handler in the server
//---------------------------------------------------------------------------
func (pw *ProtoWs) StreamHandler(ws *websocket.Conn) {
	log.Printf("in from %s\n", ws.RemoteAddr())
	defer log.Printf("out %s\n", ws.RemoteAddr())

	err := HandleRequest(ws)
	if err != nil {
		log.Println(err)
		return
	}

	sbuf := sr.NewStreamRing(2, sb.MBYTE)

	//err = RecvMultipartData(ws, pw.Boundary)
	err = RecvMultipartToRing(ws, sbuf)
	if err != nil {
		log.Println(err)
		return
	}
}

//---------------------------------------------------------------------------
// WebSocket summit a request to the server
//---------------------------------------------------------------------------
func SendRequestPost(ws *websocket.Conn, boundary string) error {
	var err error

	// send POST request
	smsg := "POST /stream HTTP/1.1\r\n"
	smsg += fmt.Sprintf("Content-Type: multipart/x-mixed-replace; boundary=--%s\r\n", boundary)
	smsg += "User-Agent: Happy Media WS Client\r\n"
	smsg += "\r\n"

	n, err := ws.Write([]byte(smsg))
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Printf(">> Send(%d):\n%s", n, smsg)

	// recv response
	rmsg := make([]byte, 512)

	n, err = ws.Read(rmsg)
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Printf("<< Recv(%d):\n%s", n, rmsg[:n])

	return err
}

//---------------------------------------------------------------------------
// WebSocket handle a client request in the server
//---------------------------------------------------------------------------
func HandleRequest(ws *websocket.Conn) error {
	var err error

	// recv POST request
	rmsg := make([]byte, 512)

	n, err := ws.Read(rmsg)
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Printf("<< Recv(%d):\n%s", n, rmsg[:n])

	// parse request
	err = ParseRequest(string(rmsg[:n]))
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
	fmt.Printf(">> Send(%d):\n%s", n, smsg)

	return err
}

//---------------------------------------------------------------------------
// WebSocket parse request
//---------------------------------------------------------------------------
func ParseRequest(msg string) error {
	var err error

	req, err := GetRequest(msg)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = GetBoundary(req.Header.Get("Content-Type"))
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// WebSocket get http request from message
//---------------------------------------------------------------------------
func GetRequest(msg string) (*http.Request, error) {
	var err error

	reader := bufio.NewReader(strings.NewReader(msg))
	req, err := http.ReadRequest(reader)
	if err != nil {
		log.Println(err)
		return req, err
	}
	//fmt.Println(req.Header)

	return req, err
}

//---------------------------------------------------------------------------
// WebSocket get boundary string from request
//---------------------------------------------------------------------------
func GetBoundary(ctype string) (string, error) {
	var err error

	mt, params, err := mime.ParseMediaType(ctype)
	//fmt.Printf("%v %v\n", params, req.Header.Get("Content-Type"))
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
// WebSocket get header from message
//---------------------------------------------------------------------------
func GetHeader(msg string) (http.Header, error) {
	var err error

	reader := bufio.NewReader(strings.NewReader(msg))
	tp := textproto.NewReader(reader)

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	httpHeader := http.Header(mimeHeader)
	//fmt.Println(httpHeader)

	return httpHeader, err
}

//---------------------------------------------------------------------------
// send files in multipart
//---------------------------------------------------------------------------
func SendMultipartFiles(ws *websocket.Conn, pattern string, boundary string) error {
	var err error

	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Println(err)
		return err
	}
	if files == nil {
		log.Printf("no file for '%s'\n", pattern)
		return sb.ErrNull
	}

	//mw := multipart.NewWriter(os.Stdout)	// for debug
	mw := multipart.NewWriter(ws)
	mw.SetBoundary(boundary)

	for i := range files {
		err = SendPartFile(mw, files[i])
		if err != nil {
			if err == sb.ErrSize { // ignore file of size error
				continue
			}
			log.Println(err)
			return err
		}
		time.Sleep(time.Millisecond)
	}

	return err
}

//---------------------------------------------------------------------------
// send a file in part
//---------------------------------------------------------------------------
func SendPartFile(mw *multipart.Writer, file string) error {
	var err error

	fmt.Println(file)
	fdata, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println(err)
		return err
	}

	if len(fdata) == 0 {
		fmt.Printf(">> ignore '%s'\n", file)
		return sb.ErrSize
	}

	ctype := mime.TypeByExtension(file)
	if ctype == "" {
		ctype = http.DetectContentType(fdata)
	}

	err = SendPartData(mw, fdata, len(fdata), ctype)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// recv multipart (for debugging)
//---------------------------------------------------------------------------
func RecvMultipartData(ws *websocket.Conn, boundary string) error {
	var err error

	mr := multipart.NewReader(ws, boundary)

	for {
		err = RecvPartToData(mr)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return err
}

//---------------------------------------------------------------------------
// recv multipart to ring buffer
//---------------------------------------------------------------------------
func RecvMultipartToRing(ws *websocket.Conn, sbuf *sr.StreamRing) error {
	var err error

	mr := multipart.NewReader(ws, sbuf.Boundary)

	for {
		slot, pos := sbuf.GetSlotIn()

		err = RecvPartToSlot(mr, slot)
		if err != nil {
			log.Println(err)
			return err
		}

		sbuf.SetPosInByPos(pos + 1)
	}

	return err
}

//---------------------------------------------------------------------------
// send a part of multipart
//---------------------------------------------------------------------------
func SendPartData(mw *multipart.Writer, data []byte, dsize int, dtype string) error {
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
// recv a part of multipart
//---------------------------------------------------------------------------
func RecvPartToData(mr *multipart.Reader) error {
	var err error

	p, nl, err := RecvPartHeader(mr)
	if err != nil {
		log.Println(err)
		return err
	}

	data := make([]byte, nl)

	err = RecvPartBodyToData(p, nl, data)
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Printf("%d [%0x-%0x]\n", nl, data[:2], data[nl-2:])
	return err
}

//---------------------------------------------------------------------------
// recv a part to slot of ring
//---------------------------------------------------------------------------
func RecvPartToSlot(mr *multipart.Reader, ss *sr.StreamSlot) error {
	var err error

	p, nl, err := RecvPartHeader(mr)
	if err != nil {
		log.Println(err)
		return err
	}

	err = RecvPartBodyToSlot(p, nl, ss)
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Printf("%d [%0x-%0x]\n", nl, ss.Content[:2], ss.Content[nl-2:nl])
	return err
}

//---------------------------------------------------------------------------
// read a part header and parse it
//---------------------------------------------------------------------------
func RecvPartHeader(mr *multipart.Reader) (*multipart.Part, int, error) {
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
func RecvPartBodyToData(p *multipart.Part, nl int, data []byte) error {
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
func RecvPartBodyToSlot(p *multipart.Part, nl int, ss *sr.StreamSlot) error {
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
