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
const (
	STR_ECHO_CLIENT = "Echo WS Client"
	STR_ECHO_SERVER = "Echo WS Server"

	STR_WS_CASTER = "Happy Media WS Caster"
	STR_WS_SERVER = "Happy Media WS Server"
	STR_WS_PLAYER = "Happy Media WS Player"
)

type ProtoWs struct {
	Host     string
	Port     string
	PortTls  string // for TLS
	Desc     string
	Method   string
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
// new ProtoWs struct, variadic argument
//---------------------------------------------------------------------------
func NewProtoWs(args ...string) *ProtoWs {
	pw := &ProtoWs{
		Host:     sb.STR_DEF_HOST,
		Port:     sb.STR_DEF_PORT,
		PortTls:  sb.STR_DEF_PTLS,
		Boundary: sb.STR_DEF_BDRY,
	}

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
	pw.Desc = "cleared"
	if pw.Conn != nil {
		pw.Conn.Close()
		pw.Conn = nil
	}
}

//---------------------------------------------------------------------------
// Echo client
// - http://www.websocket.org/echo.html
//---------------------------------------------------------------------------
func (pw *ProtoWs) EchoClient(smsg string) error {
	var err error
	log.Printf("%s\n", STR_ECHO_CLIENT)

	origin := fmt.Sprintf("http://%s/", pw.Host)
	url := fmt.Sprintf("ws://%s:%s/echo", pw.Host, pw.Port)

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Println(err)
		return err
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
		log.Println(err)
		return err
	}
	log.Println(rmsg)

	return err
}

//---------------------------------------------------------------------------
// Echo server
//---------------------------------------------------------------------------
func (pw *ProtoWs) EchoServer() error {
	var err error
	log.Printf("%s\n", STR_ECHO_SERVER)

	http.Handle("/echo", websocket.Handler(pw.EchoHandler))
	log.Fatal(http.ListenAndServe(":"+pw.Port, nil))

	return err
}

//---------------------------------------------------------------------------
// websocket echo handler
//---------------------------------------------------------------------------
func (pw *ProtoWs) EchoHandler(ws *websocket.Conn) {
	var err error

	mode := "change"

	switch mode {
	case "direct":
		io.Copy(ws, ws)

	case "change":
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

	default:
		log.Println("unknown mode")
	}
}

//---------------------------------------------------------------------------
// caster for test and debugging
//---------------------------------------------------------------------------
func (pw *ProtoWs) ActCaster() error {
	var err error
	log.Printf("%s\n", STR_WS_CASTER)

	origin := fmt.Sprintf("http://%s/", pw.Host)
	url := fmt.Sprintf("ws://%s:%s/stream", pw.Host, pw.Port)

	ws, err := websocket.Dial(url, "", origin)
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
// websocket server
//---------------------------------------------------------------------------
func (pw *ProtoWs) ActServer() error {
	var err error
	log.Printf("%s on ws:%s and wss:%s\n", STR_WS_SERVER, pw.Port, pw.PortTls)

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
// websocket player
//---------------------------------------------------------------------------
func (pw *ProtoWs) ActPlayer() error {
	var err error
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
	err = pw.ReadMultipartToData(r)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//--------------------------------------------------------------------------r
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
// WebSocket stream handler in the server
//---------------------------------------------------------------------------
func (pw *ProtoWs) StreamHandler(ws *websocket.Conn) {
	log.Printf("in from %s\n", ws.RemoteAddr())
	defer log.Printf("out %s\n", ws.RemoteAddr())
	defer ws.Close()

	r := bufio.NewReader(ws)
	w := bufio.NewWriter(ws)

	ring := sr.NewStreamRing(2, sb.MBYTE)

	err := pw.HandleRequest(r, w, ring)
	if err != nil {
		log.Println(err)
		return
	}

	//err = pw.ReadMultipartToData(r, pw.Boundary)
	err = pw.ReadMultipartToRing(r, ring)
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
	req += fmt.Sprintf("User-Agent: %s\r\n", STR_WS_PLAYER)
	req += "\r\n"

	_, err = w.Write([]byte(req))
	if err != nil {
		log.Println(err)
		return err
	}
	w.Flush()

	log.Printf("SEND [%d]\n%s", len(req), req)

	// recv response
	err = pw.ReadMessage(r)
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
	req += fmt.Sprintf("Content-Type: multipart/x-mixed-replace; boundary=%s\r\n", pw.Boundary)
	req += fmt.Sprintf("User-Agent: %s\r\n", STR_WS_CASTER)
	req += "\r\n"

	_, err = w.Write([]byte(req))
	if err != nil {
		log.Println(err)
		return err
	}
	w.Flush()

	log.Printf("SEND [%d]\n%s", len(req), req)

	// recv response
	err = pw.ReadMessage(r)
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
	res += fmt.Sprintf("Server: %s\r\n", STR_WS_SERVER)
	res += "\r\n"

	defer log.Printf("SEND [%d]\n%s", len(res), res)

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
	res += fmt.Sprintf("Server: %s\r\n", STR_WS_SERVER)
	res += fmt.Sprintf("Content-Type: multipart/x-mixed-replace; boundary=%s\r\n", pw.Boundary)
	res += "\r\n"

	defer log.Printf("SEND [%d]\n%s", len(res), res)

	_, err = w.Write([]byte(res))
	if err != nil {
		log.Println(err)
		return err
	}
	w.Flush()

	return err
}

//---------------------------------------------------------------------------
// WebSocket handle a client request in the server
//---------------------------------------------------------------------------
func (pw *ProtoWs) HandleRequest(r *bufio.Reader, w *bufio.Writer, ring *sr.StreamRing) error {
	var err error

	// recv request and parse it
	err = pw.ReadMessage(r)
	if err != nil {
		log.Println(err)
		return err
	}

	// send response and multipart
	switch pw.Method {
	case "POST":
		err = pw.ResponsePost(w)
		if err != nil {
			log.Println(err)
			return err
		}
		err = pw.ReadMultipartToRing(r, ring)
		if err != nil {
			log.Println(err)
			return err
		}
	case "GET":
		err = pw.ResponseGet(w)
		if err != nil {
			log.Println(err)
			return err
		}
		err = pw.WriteRingInMultipart(w, ring)
		//err = pt.WriteDataToStream(w, ring)
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
// read http message
//---------------------------------------------------------------------------
func (pw *ProtoWs) ReadMessage(r *bufio.Reader) error {
	var err error

	headers, err := pw.ReadMessageHeader(r)
	if err != nil {
		log.Println(err)
		return err
	}

	value, ok := headers[sb.STR_HDR_CONTENT_TYPE]
	if ok {
		pw.Boundary, err = GetTypeBoundary(value)
	}

	clen := 0
	value, ok = headers[sb.STR_HDR_CONTENT_LENGTH]
	if ok {
		clen, _ = strconv.Atoi(value)
	}

	if clen > 0 {
		_, err = pw.ReadBodyToData(r, clen)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return err
}

//---------------------------------------------------------------------------
// read header of message and return a map
//---------------------------------------------------------------------------
func (pw *ProtoWs) ReadMessageHeader(r *bufio.Reader) (map[string]string, error) {
	var err error

	result := make(map[string]string)

	// parse a request line
	line, _, err := r.ReadLine()
	if err != nil {
		log.Println(err)
		return result, err
	}

	res := strings.Fields(string(line))
	pw.Method = res[0]

	// parse header lines
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			log.Println(err)
			break
		}
		//fmt.Println(string(line))

		if string(line) == "" {
			break
		}

		keyvalue := strings.SplitN(string(line), ":", 2)
		if len(keyvalue) > 1 {
			result[keyvalue[0]] = strings.TrimSpace(keyvalue[1])
		}
	}

	//fmt.Println(result)
	return result, err
}

//---------------------------------------------------------------------------
// read(recv) body of message to data
//---------------------------------------------------------------------------
func (pw *ProtoWs) ReadBodyToData(r *bufio.Reader, clen int) ([]byte, error) {
	var err error

	data := make([]byte, clen)

	tn := 0
	for tn < clen {
		n, err := r.Read(data[tn:])
		if err != nil {
			log.Println(err)
			break
		}
		tn += n
	}

	//fmt.Printf(string(data[:tn]))
	fmt.Printf("[DATA] (%d/%d)\n\n", tn, clen)
	return data, err
}

//---------------------------------------------------------------------------
// WebSocket parse request
//---------------------------------------------------------------------------
func ParseRequest(msg string) error {
	var err error

	req, err := GetHttpRequest(msg)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = GetTypeBoundary(req.Header.Get(sb.STR_HDR_CONTENT_TYPE))
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// get http request from message
//---------------------------------------------------------------------------
func GetHttpRequest(msg string) (*http.Request, error) {
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
// get http header from message
//---------------------------------------------------------------------------
func GetHttpHeader(msg string) (http.Header, error) {
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
// get boundary string from request
//---------------------------------------------------------------------------
func GetTypeBoundary(ctype string) (string, error) {
	var err error

	mt, params, err := mime.ParseMediaType(ctype)
	//fmt.Printf("%v %v\n", params, req.Header.Get("Content-Type"))
	if err != nil {
		log.Printf("ParseMediaType: %s %v", mt, err)
		return "", err
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
func (pw *ProtoWs) ReadMultipartToData(r *bufio.Reader) error {
	var err error

	mr := multipart.NewReader(r, pw.Boundary)

	for {
		err = ReadPartToData(mr)
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
func (pw *ProtoWs) ReadMultipartToRing(r *bufio.Reader, ring *sr.StreamRing) error {
	var err error

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
func (pw *ProtoWs) WriteRingInMultipart(w *bufio.Writer, ring *sr.StreamRing) error {
	var err error

	if !ring.IsUsing() {
		log.Println("ErrStatus")
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
	slot.Timestamp = sb.GetTimestamp()
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
func WriteSlotInPart(mw *multipart.Writer, ss *sr.StreamSlot) error {
	var err error

	buf := new(bytes.Buffer)

	part, err := mw.CreatePart(textproto.MIMEHeader{
		sb.STR_HDR_CONTENT_TYPE:   {ss.Type},
		sb.STR_HDR_CONTENT_LENGTH: {strconv.Itoa(ss.Length)},
	})
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = buf.Write(ss.Content[:ss.Length])
	_, err = buf.WriteTo(part)

	return err
}

//---------------------------------------------------------------------------
// recv a part of multipart
//---------------------------------------------------------------------------
func ReadPartToData(mr *multipart.Reader) error {
	var err error

	p, nl, err := ReadPartHeader(mr)
	if err != nil {
		log.Println(err)
		return err
	}

	ss := sr.NewStreamSlotBySize(nl)

	err = ReadPartBodyToSlot(p, nl, ss)
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Printf("%d [%0x-%0x]\n", nl, ss.Content[:2], ss.Content[nl-2:nl])
	return err
}

//---------------------------------------------------------------------------
// recv a part to slot of ring
//---------------------------------------------------------------------------
func ReadPartToSlot(mr *multipart.Reader, ss *sr.StreamSlot) error {
	var err error

	p, nl, err := ReadPartHeader(mr)
	if err != nil {
		log.Println(err)
		return err
	}

	err = ReadPartBodyToSlot(p, nl, ss)
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
func ReadPartHeader(mr *multipart.Reader) (*multipart.Part, int, error) {
	var err error

	p, err := mr.NextPart()
	if err != nil { // io.EOF
		log.Println(err)
		return p, 0, err
	}

	sl := p.Header.Get(sb.STR_HDR_CONTENT_LENGTH)
	nl, err := strconv.Atoi(sl)
	if err != nil {
		log.Printf("%s %s -> %d\n", p.Header, sl, nl)
		return p, nl, err
	}

	return p, nl, err
}

//---------------------------------------------------------------------------
// read a part body to slot
//---------------------------------------------------------------------------
func ReadPartBodyToSlot(p *multipart.Part, nl int, ss *sr.StreamSlot) error {
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
