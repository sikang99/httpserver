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
	"crypto/tls"
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
	"time"

	sb "stoney/httpserver/src/streambase"
	sr "stoney/httpserver/src/streamring"

	"github.com/fatih/color"
	"golang.org/x/net/websocket"
)

//---------------------------------------------------------------------------
const (
	STR_ECHO_CLIENT = "Echo WebSocket Client"
	STR_ECHO_SERVER = "Echo WebSocket Server"

	STR_WS_CASTER = "Happy Media WebSocket Caster"
	STR_WS_SERVER = "Happy Media WebSocket Server"
	STR_WS_PLAYER = "Happy Media WebSocket Player"
)

type ProtoWs struct {
	Mode     string // normal or TLS secure mode
	Host     string
	Port     string
	PortTls  string // for TLS
	Port2    string // for HTTP/2
	Desc     string
	Method   string
	Boundary string
	Conn     *websocket.Conn
	Ring     *sr.StreamRing
}

//---------------------------------------------------------------------------
// string information of struct
//---------------------------------------------------------------------------
func (pw *ProtoWs) String() string {
	str := fmt.Sprintf("\tMode: %s", pw.Mode)
	str += fmt.Sprintf("\tHost: %s", pw.Host)
	str += fmt.Sprintf("\tPort: %s,%s,%s", pw.Port, pw.PortTls, pw.Port2)
	str += fmt.Sprintf("\tConn: %v", pw.Conn)
	str += fmt.Sprintf("\tMethod: %s", pw.Method)
	str += fmt.Sprintf("\tBoundary: %s", pw.Boundary)
	str += fmt.Sprintf("\tDesc: %s", pw.Desc)
	return str
}

//---------------------------------------------------------------------------
// new ProtoWs struct, variadic argument
//---------------------------------------------------------------------------
func NewProtoWs(args ...string) *ProtoWs {
	pw := &ProtoWs{
		Mode:     sb.STR_DEF_MODE,
		Host:     sb.STR_DEF_HOST,
		Port:     sb.STR_DEF_PORT,
		PortTls:  sb.STR_DEF_PTLS,
		Port2:    sb.STR_DEF_PORT2,
		Boundary: sb.STR_DEF_BDRY,
	}

	pw.Desc = "NewProtoWs"

	for i, arg := range args {
		if i == 0 {
			pw.Host = arg
		} else if i == 1 {
			pw.Port = arg
		} else if i == 2 {
			pw.PortTls = arg
		} else if i == 3 {
			pw.Desc = arg
		}
	}

	return pw
}

func NewProtoWsWithPorts(args ...string) *ProtoWs {
	pw := NewProtoWs()

	pw.Desc = "NewProtoWsWithPorts"

	for i, arg := range args {
		switch {
		case i == 0:
			pw.Port = arg
		case i == 1:
			pw.PortTls = arg
		case i == 2:
			pw.Port2 = arg
		}
	}

	return pw
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
	pw.Mode = sb.STR_DEF_MODE
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
	pw.Desc = "cleared"
	if pw.Conn != nil {
		pw.Conn.Close()
		pw.Conn = nil
	}
}

//---------------------------------------------------------------------------
// Echo client
// - http://www.websocket.org/echo.html
// - http://41j.com/blog/2014/12/simple-websocket-example-golang/
//---------------------------------------------------------------------------
func (pw *ProtoWs) EchoClient(smsg string) error {
	var err error
	log.Printf("%s to %s:%s,%s\n", STR_ECHO_CLIENT, pw.Host, pw.Port, pw.PortTls)

	ws, err := pw.Connect("echo", "sec")
	//ws, err := pw.Connect("/echo")
	if err != nil {
		log.Println(err)
		return err
	}

	// if testing state?
	if sb.IsTerminal() {
		err = InputSendReceive(ws)
	} else {
		err = SendReceive(ws, smsg)
	}

	return err
}

//---------------------------------------------------------------------------
// send and receive message
//---------------------------------------------------------------------------
func SendReceive(ws *websocket.Conn, smsg string) error {
	var err error

	err = websocket.Message.Send(ws, smsg)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(color.BlueString(smsg))

	var rmsg string
	err = websocket.Message.Receive(ws, &rmsg)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(color.GreenString(rmsg))

	return err
}

//---------------------------------------------------------------------------
// input a line and send, receive message
//---------------------------------------------------------------------------
func InputSendReceive(ws *websocket.Conn) error {
	var err error

	r := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		line, _, err := r.ReadLine()
		if err != nil {
			log.Println(err)
			break
		}

		input := strings.Replace(string(line), "\r", "", -1)

		if strings.EqualFold(input, "quit") {
			fmt.Println("Bye bye.")
			break
		}

		err = SendReceive(ws, input)
		if err != nil {
			log.Println(err)
			break
		}
	}

	return err
}

//---------------------------------------------------------------------------
// Echo server
//---------------------------------------------------------------------------
func (pw *ProtoWs) EchoServer() (err error) {
	log.Printf("%s\n", STR_ECHO_SERVER)

	http.Handle("/echo", websocket.Handler(pw.EchoHandler))
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

	return err
}

//---------------------------------------------------------------------------
// echo handler
//---------------------------------------------------------------------------
func (pw *ProtoWs) EchoHandler(ws *websocket.Conn) {
	var err error

	log.Printf("in for %s\n", ws.LocalAddr())
	defer log.Printf("out for %s\n", ws.RemoteAddr())
	defer ws.Close()

	mode := "change"

	switch mode {
	case "direct":
		io.Copy(ws, ws)

	case "change":
		for {
			var rmsg string

			err = websocket.Message.Receive(ws, &rmsg)
			if err != nil {
				sb.LogPrintln(err)
				break
			}
			//log.Println("Received from client: " + rmsg)

			if strings.EqualFold(rmsg, "quit") {
				log.Println("quit")
				os.Exit(0)
			}

			smsg := "Echo " + rmsg
			err = websocket.Message.Send(ws, smsg)
			if err != nil {
				sb.LogPrintln(err)
				break
			}
			//log.Println("Sending to client: " + smsg)
		}

	default:
		log.Println("unknown echo mode")
	}
}

//---------------------------------------------------------------------------
// Caster for test and debugging
//---------------------------------------------------------------------------
func (pw *ProtoWs) StreamCaster(sec ...string) (err error) {
	log.Printf("%s\n", STR_WS_CASTER)

	ws, err := pw.Connect("/stream")
	if err != nil {
		log.Println(err)
		return err
	}
	defer ws.Close()

	r := bufio.NewReader(ws)
	w := bufio.NewWriter(ws)

	err = pw.RequestPost(r, w)
	if err != nil {
		log.Println(err)
		return err
	}

	err = WriteMultipartFiles(w, "../../static/image/*.jpg", pw.Boundary, true)

	return err
}

//---------------------------------------------------------------------------
// Server
//---------------------------------------------------------------------------
func (pw *ProtoWs) StreamServer() (err error) {
	log.Printf("%s on ws:%s and wss:%s\n", STR_WS_SERVER, pw.Port, pw.PortTls)

	pw.Ring = sr.NewStreamRing(2, sb.MBYTE)

	http.Handle("/stream", websocket.Handler(pw.StreamHandler))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	wg := sync.WaitGroup{}

	wg.Add(1)
	// HTTP server
	go pw.serveHttp(&wg)

	wg.Add(1)
	// HTTPS server
	go pw.serveHttps(&wg)

	wg.Wait()

	return err
}

//---------------------------------------------------------------------------
// Player
//---------------------------------------------------------------------------
func (pw *ProtoWs) StreamPlayer() (err error) {
	log.Printf("%s\n", STR_WS_PLAYER)

	origin := fmt.Sprintf("http://%s/", pw.Host)
	url := fmt.Sprintf("ws://%s:%s/stream", pw.Host, pw.Port)

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Println(err)
		return err
	}

	r := bufio.NewReader(ws)
	w := bufio.NewWriter(ws)

	err = pw.RequestGet(r, w)
	if err != nil {
		log.Println(err)
		return err
	}

	// recv multipart stream from server
	err = ReadMultipartToData(r, pw.Boundary)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// connect
// - https://code.google.com/p/go/source/browse/websocket/client.go?repo=net&r=d7ff1d8f275c5693a5fc301cbfdf0a6f89e4f57c
// - http://andrewwdeane.blogspot.kr/2013/01/gobing-down-secure-websockets.html
//---------------------------------------------------------------------------
func (pw *ProtoWs) Connect(hand string, secure ...string) (ws *websocket.Conn, err error) {
	if secure == nil {
		origin := fmt.Sprintf("http://%s:%s", pw.Host, pw.Port)
		url := fmt.Sprintf("ws://%s:%s%s", pw.Host, pw.Port, hand)

		ws, err = websocket.Dial(url, "", origin)
	} else {
		pw.Mode = "secure"
		origin := fmt.Sprintf("http://%s:%s", pw.Host, pw.PortTls)
		url := fmt.Sprintf("wss://%s:%s%s", pw.Host, pw.PortTls, hand)

		wconf, err := websocket.NewConfig(url, origin)
		cert, err := tls.LoadX509KeyPair("sec/cert.pem", "sec/key.pem")
		if err != nil {
			log.Println(err)
			return nil, err
		}
		tconf := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
		wconf.TlsConfig = &tconf

		ws, err = websocket.DialConfig(wconf)
	}

	if err == nil {
		log.Printf("connected in %s mode", pw.Mode)
	}

	return ws, err
}

//---------------------------------------------------------------------------
// start server to receive requests
//---------------------------------------------------------------------------
func (pw *ProtoWs) serveHttp(wg *sync.WaitGroup) {
	log.Println("Starting WS server at http://" + pw.Host + ":" + pw.Port)
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
	log.Println("Starting WSS server at https://" + pw.Host + ":" + pw.PortTls)
	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + pw.PortTls,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServeTLS("sec/cert.pem", "sec/key.pem"))
}

//---------------------------------------------------------------------------
// stream handler in the server
//---------------------------------------------------------------------------
func (pw *ProtoWs) StreamHandler(ws *websocket.Conn) {
	var err error

	log.Printf("in for %s\n", ws.LocalAddr())
	defer log.Printf("out for %s\n", ws.RemoteAddr())
	defer ws.Close()

	r := bufio.NewReader(ws)
	w := bufio.NewWriter(ws)

	err = pw.HandleRequest(r, w, pw.Ring)
	if err != nil {
		log.Println(err)
		return
	}
}

//---------------------------------------------------------------------------
// a GET request and get its response
//---------------------------------------------------------------------------
func (pw *ProtoWs) RequestGet(r *bufio.Reader, w *bufio.Writer) error {
	var err error

	// send GET request
	req := "GET /stream HTTP/1.1\r\n"
	req += fmt.Sprintf("%s: %s\r\n", sb.STR_HDR_USER_AGENT, STR_WS_PLAYER)
	req += "\r\n"

	_, err = w.Write([]byte(req))
	if err != nil {
		log.Println(err)
		return err
	}
	w.Flush()

	log.Printf("SEND [%d]\n%s", len(req), color.GreenString(req))

	// recv response
	err = pw.ReadResponse(r)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// a POST request to the server
//---------------------------------------------------------------------------
func (pw *ProtoWs) RequestPost(r *bufio.Reader, w *bufio.Writer) error {
	var err error

	// send POST request
	req := "POST /stream HTTP/1.1\r\n"
	req += fmt.Sprintf("%s: multipart/x-mixed-replace; boundary=%s\r\n", sb.STR_HDR_CONTENT_TYPE, pw.Boundary)
	req += fmt.Sprintf("%s: %s\r\n", sb.STR_HDR_USER_AGENT, STR_WS_CASTER)
	req += "\r\n"

	_, err = w.Write([]byte(req))
	if err != nil {
		log.Println(err)
		return err
	}
	w.Flush()

	log.Printf("SEND [%d]\n%s", len(req), color.GreenString(req))

	// recv response
	err = pw.ReadResponse(r)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// send response for POST request
//---------------------------------------------------------------------------
func (pw *ProtoWs) ResponsePost(w *bufio.Writer) error {
	var err error

	res := "HTTP/1.1 200 Ok\r\n"
	res += fmt.Sprintf("%s: %s\r\n", sb.STR_HDR_SERVER, STR_WS_SERVER)
	res += "\r\n"

	defer log.Printf("SEND [%d]\n%s", len(res), color.CyanString(res))

	_, err = w.Write([]byte(res))
	if err != nil {
		log.Println(err)
		return err
	}
	w.Flush()

	return err
}

//---------------------------------------------------------------------------
// send response for GET request
//---------------------------------------------------------------------------
func (pw *ProtoWs) ResponseGet(w *bufio.Writer) error {
	var err error

	res := "HTTP/1.1 200 Ok\r\n"
	res += fmt.Sprintf("%s: multipart/x-mixed-replace; boundary=%s\r\n", sb.STR_HDR_CONTENT_TYPE, pw.Boundary)
	res += fmt.Sprintf("%s: %s\r\n", sb.STR_HDR_SERVER, STR_WS_SERVER)
	res += "\r\n"

	defer log.Printf("SEND [%d]\n%s", len(res), color.CyanString(res))

	_, err = w.Write([]byte(res))
	if err != nil {
		log.Println(err)
		return err
	}
	w.Flush()

	return err
}

//---------------------------------------------------------------------------
// handle a client request in the server
//---------------------------------------------------------------------------
func (pw *ProtoWs) HandleRequest(r *bufio.Reader, w *bufio.Writer, ring *sr.StreamRing) error {
	var err error

	// recv request and parse it
	err = pw.ReadRequest(r)
	if err != nil {
		log.Println(err)
		return err
	}

	// send response and multipart
	switch pw.Method {
	case "POST":
		ring.Boundary = pw.Boundary

		err = pw.ResponsePost(w)
		if err != nil {
			log.Println(err)
			return err
		}
		err = ReadMultipartToRing(r, ring)
		if err != nil {
			log.Println(err)
			return err
		}
	case "GET":
		pw.Boundary = ring.Boundary

		err = pw.ResponseGet(w)
		if err != nil {
			log.Println(err)
			return err
		}
		err = WriteRingInMultipart(w, ring)
		if err != nil {
			log.Println(err)
			return err
		}
	default:
		err = sb.ErrSupport
	}

	return err
}

//---------------------------------------------------------------------------
// read http request message
//---------------------------------------------------------------------------
func (pw *ProtoWs) ReadRequest(r *bufio.Reader) error {
	var err error

	req, err := http.ReadRequest(r)
	if err != nil {
		log.Println(err)
		return err
	}
	//fmt.Println(req)

	pw.Method = req.Method
	ctype := req.Header.Get(sb.STR_HDR_CONTENT_TYPE)
	if ctype != "" {
		pw.Boundary, err = GetTypeBoundary(req.Header.Get(sb.STR_HDR_CONTENT_TYPE))
		if err != nil {
			log.Println(err)
			return err
		}
	}
	//fmt.Println(pw)

	return err
}

//---------------------------------------------------------------------------
// read http response message
//---------------------------------------------------------------------------
func (pw *ProtoWs) ReadResponse(r *bufio.Reader) error {
	var err error

	res, err := http.ReadResponse(r, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	//fmt.Println(res)

	if res.StatusCode != 200 {
		log.Println(res.Status)
		return sb.ErrStatus
	}

	return err
}

//---------------------------------------------------------------------------
// get boundary string from request
//---------------------------------------------------------------------------
func GetTypeBoundary(ctype string) (string, error) {
	var err error

	mt, params, err := mime.ParseMediaType(ctype)
	//fmt.Printf("%v %v\n", params, req.Header.Get("Content-Type"))
	if err != nil {
		log.Printf("ParseMediaType: %s %v", mt, err)
		return "", sb.ErrParse
	}

	boundary := params["boundary"]
	if !strings.HasPrefix(boundary, "--") {
		log.Printf("expected to start with --, got %q", boundary)
	}

	return boundary, err
}

//---------------------------------------------------------------------------
// write files in multipart
//---------------------------------------------------------------------------
func WriteMultipartFiles(w *bufio.Writer, pattern string, boundary string, loop bool) error {
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

	mw := multipart.NewWriter(w)
	mw.SetBoundary(boundary)

	for {
		for i := range files {
			err = WriteFileInPart(mw, files[i])
			if err != nil {
				if err == sb.ErrSize { // ignore file of size error
					continue
				}
				log.Println(err)
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
// write a file in part
//---------------------------------------------------------------------------
func WriteFileInPart(mw *multipart.Writer, file string) error {
	var err error

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

	err = WriteDataInPart(mw, fdata, len(fdata), ctype)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// recv multipart (for debugging)
//---------------------------------------------------------------------------
func ReadMultipartToData(r *bufio.Reader, boundary string) error {
	var err error

	slot := sr.NewStreamSlot()
	mr := multipart.NewReader(r, boundary)

	for {
		err = ReadPartToSlot(mr, slot)
		if err != nil {
			log.Println(err)
			return err
		}
		fmt.Println(">4", slot)
	}

	return err
}

//---------------------------------------------------------------------------
// recv multipart to ring buffer
//---------------------------------------------------------------------------
func ReadMultipartToRing(r *bufio.Reader, ring *sr.StreamRing) error {
	var err error

	err = ring.SetStatusUsing()
	if err != nil {
		return sb.ErrStatus
	}
	defer ring.Reset()

	mr := multipart.NewReader(r, ring.Boundary)

	for {
		slot, pos := ring.GetSlotIn()

		err = ReadPartToSlot(mr, slot)
		if err != nil {
			log.Println(err)
			return err
		}

		ring.SetPosInByPos(pos + 1)
	}

	return err
}

//---------------------------------------------------------------------------
// recv multipart to ring buffer
//---------------------------------------------------------------------------
func WriteRingInMultipart(w *bufio.Writer, ring *sr.StreamRing) error {
	var err error

	if !ring.IsUsing() {
		sb.LogPrintln("ErrStatus")
		return sb.ErrStatus
	}

	mw := multipart.NewWriter(w)
	mw.SetBoundary(ring.Boundary)

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

		err = WriteSlotInPart(mw, slot)
		if err != nil {
			log.Println(err)
			break
		}
		fmt.Println("3>", slot)

		pos = npos
	}

	return err
}

//---------------------------------------------------------------------------
// send a part of multipart
//---------------------------------------------------------------------------
func WriteDataInPart(mw *multipart.Writer, data []byte, dsize int, dtype string) error {
	var err error

	// prepare a slot and its data
	slot := sr.NewStreamSlotBySize(dsize)
	defer fmt.Println("1>", slot)

	slot.Type = dtype
	slot.Length = dsize
	slot.Timestamp = sb.GetTimestampNow()
	copy(slot.Content, data)

	// send the slot
	err = WriteSlotInPart(mw, slot)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// send a part of multipart
//---------------------------------------------------------------------------
func WriteSlotInPart(mw *multipart.Writer, slot *sr.StreamSlot) error {
	var err error

	buf := new(bytes.Buffer)

	part, err := mw.CreatePart(textproto.MIMEHeader{
		sb.STR_HDR_CONTENT_TYPE:   {slot.Type},
		sb.STR_HDR_CONTENT_LENGTH: {strconv.Itoa(slot.Length)},
		sb.STR_HDR_TIMESTAMP:      {strconv.FormatInt(slot.Timestamp, 10)},
	})
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = buf.Write(slot.Content[:slot.Length])
	_, err = buf.WriteTo(part)

	return err
}

//---------------------------------------------------------------------------
// recv a part to slot of ring
//---------------------------------------------------------------------------
func ReadPartToSlot(mr *multipart.Reader, slot *sr.StreamSlot) error {
	var err error

	// read a part
	p, err := mr.NextPart()
	if err != nil { // io.EOF
		log.Println(err)
		return err
	}

	nl, err := ParsePartHeaderToSlot(p, slot)
	if err != nil {
		log.Println(err)
		return err
	}

	err = ReadPartBodyToSlot(p, slot, nl)
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Println(">2", slot)
	return err
}

//---------------------------------------------------------------------------
// read a part header and parse it
//---------------------------------------------------------------------------
func ParsePartHeaderToSlot(p *multipart.Part, slot *sr.StreamSlot) (int, error) {
	var err error

	// get header information of each part
	stamp := p.Header.Get(sb.STR_HDR_TIMESTAMP)
	slot.Timestamp = sb.GetTimestampFromString(stamp)

	slot.Type = p.Header.Get(sb.STR_HDR_CONTENT_TYPE)

	sl := p.Header.Get(sb.STR_HDR_CONTENT_LENGTH)
	nl, err := strconv.Atoi(sl)
	if err != nil {
		log.Printf("%s %s -> %d\n", p.Header, sl, nl)
		return nl, err
	}

	return nl, err
}

//---------------------------------------------------------------------------
// read a part body to slot
//---------------------------------------------------------------------------
func ReadPartBodyToSlot(p *multipart.Part, slot *sr.StreamSlot, nl int) error {
	var err error

	// to prevent from handling incomplete data
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

	return err
}

// ---------------------------------E-----N-----D--------------------------------
