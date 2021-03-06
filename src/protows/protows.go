//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for WebSocket(WS, WSS)
// - https://godoc.org/golang.org/x/net/websocket
// - https://github.com/golang-samples/websocket
// - http://www.ajanicij.info/content/websocket-tutorial-go
// - http://www.jonathan-petitcolas.com/2015/01/27/playing-with-websockets-in-go.html
// - https://plus.google.com/+FumitoshiUkai/posts/ERN6zYozENV
//==================================================================================

package protows

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	pb "stoney/httpserver/src/protobase"
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
	Base     *pb.ProtoBase
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
	base := pb.NewProtoBase()

	pw := &ProtoWs{
		Mode:     sb.STR_DEF_MODE,
		Host:     sb.STR_DEF_HOST,
		Port:     sb.STR_DEF_PORT,
		PortTls:  sb.STR_DEF_PTLS,
		Port2:    sb.STR_DEF_PORT2,
		Boundary: sb.STR_DEF_BDRY,
		Base:     base,
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

	//ws, err := pw.Connect("echo", "sec")
	ws, err := pw.Connect("/echo")
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
				log.Println(sb.RedString(err))
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
				log.Println(sb.RedString(err))
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

	err = pw.RequestPost(ws)
	if err != nil {
		log.Println(err)
		return err
	}

	err = WriteStreamFiles(ws, "../../static/image/*.jpg", pw.Boundary, true)

	return err
}

//---------------------------------------------------------------------------
// Server
//---------------------------------------------------------------------------
func (pw *ProtoWs) StreamServer() (err error) {
	log.Printf("%s on ws:%s and wss:%s\n", STR_WS_SERVER, pw.Port, pw.PortTls)

	pw.Ring = sr.NewStreamRingWithSize(2, sb.MBYTE)

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

	err = pw.RequestGet(ws)
	if err != nil {
		log.Println(err)
		return err
	}

	// recv multipart stream from server
	err = ReadStreamToData(ws, pw.Boundary)
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
	log.Printf("in for %s\n", ws.LocalAddr())
	defer log.Printf("out for %s\n", ws.RemoteAddr())
	defer ws.Close()

	var err error

	err = pw.HandleRequest(ws, pw.Ring)
	if err != nil {
		log.Println(err)
		return
	}
}

//---------------------------------------------------------------------------
// a GET request and get its response
//---------------------------------------------------------------------------
func (pw *ProtoWs) RequestGet(ws *websocket.Conn) error {
	var err error

	// send GET request
	req := "GET /stream HTTP/1.1\r\n"
	req += fmt.Sprintf("%s: %s\r\n", sb.STR_HDR_USER_AGENT, STR_WS_PLAYER)
	req += "\r\n"

	err = websocket.Message.Send(ws, req)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("SEND [%d]\n%s", len(req), color.GreenString(req))

	// recv response
	err = pw.ReadResponse(ws)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// a POST request to the server
//---------------------------------------------------------------------------
func (pw *ProtoWs) RequestPost(ws *websocket.Conn) error {
	var err error

	// send POST request
	req := "POST /stream HTTP/1.1\r\n"
	req += fmt.Sprintf("%s: multipart/x-mixed-replace; boundary=%s\r\n", sb.STR_HDR_CONTENT_TYPE, pw.Boundary)
	req += fmt.Sprintf("%s: %s\r\n", sb.STR_HDR_USER_AGENT, STR_WS_CASTER)
	req += "\r\n"

	err = websocket.Message.Send(ws, req)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("SEND [%d]\n%s", len(req), color.GreenString(req))

	// recv response
	err = pw.ReadResponse(ws)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// send response for POST request
//---------------------------------------------------------------------------
func (pw *ProtoWs) ResponsePost(ws *websocket.Conn) error {
	var err error

	res := "HTTP/1.1 200 Ok\r\n"
	res += fmt.Sprintf("%s: %s\r\n", sb.STR_HDR_SERVER, STR_WS_SERVER)
	res += "\r\n"

	defer log.Printf("SEND [%d]\n%s", len(res), color.CyanString(res))

	err = websocket.Message.Send(ws, res)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// send response for GET request
//---------------------------------------------------------------------------
func (pw *ProtoWs) ResponseGet(ws *websocket.Conn) error {
	var err error

	res := "HTTP/1.1 200 Ok\r\n"
	res += fmt.Sprintf("%s: multipart/x-mixed-replace; boundary=%s\r\n", sb.STR_HDR_CONTENT_TYPE, pw.Boundary)
	res += fmt.Sprintf("%s: %s\r\n", sb.STR_HDR_SERVER, STR_WS_SERVER)
	res += "\r\n"

	defer log.Printf("SEND [%d]\n%s", len(res), color.CyanString(res))

	err = websocket.Message.Send(ws, res)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// handle a client request in the server
//---------------------------------------------------------------------------
func (pw *ProtoWs) HandleRequest(ws *websocket.Conn, ring *sr.StreamRing) error {
	var err error

	// recv request and parse it
	err = pw.ReadRequest(ws)
	if err != nil {
		log.Println(err)
		return err
	}

	// send response and multipart
	switch pw.Method {
	case "POST":
		ring.Boundary = pw.Boundary

		err = pw.ResponsePost(ws)
		if err != nil {
			log.Println(err)
			return err
		}

		err = ReadStreamToRing(ws, ring, pw.Boundary)
		if err != nil {
			log.Println(err)
			return err
		}
	case "GET":
		pw.Boundary = ring.Boundary

		err = pw.ResponseGet(ws)
		if err != nil {
			log.Println(err)
			return err
		}

		err = WriteRingInStream(ws, ring)
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
func (pw *ProtoWs) ReadRequest(ws *websocket.Conn) error {
	var err error

	var rmsg string
	err = websocket.Message.Receive(ws, &rmsg)
	fmt.Print(color.GreenString(rmsg))

	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(rmsg)))
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
func (pw *ProtoWs) ReadResponse(ws *websocket.Conn) error {
	var err error

	var rmsg string
	err = websocket.Message.Receive(ws, &rmsg)
	fmt.Print(color.GreenString(rmsg))

	res, err := http.ReadResponse(bufio.NewReader(strings.NewReader(rmsg)), nil)
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
func WriteStreamFiles(ws *websocket.Conn, pattern string, boundary string, loop bool) error {
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

	for {
		for i := range files {
			err = WriteFileInFrame(ws, files[i], boundary)
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
func WriteFileInFrame(ws *websocket.Conn, file string, boundary string) error {
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

	err = WriteDataInFrame(ws, fdata, len(fdata), ctype, boundary)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// recv multipart (for debugging)
//---------------------------------------------------------------------------
func ReadStreamToData(ws *websocket.Conn, boundary string) error {
	var err error

	slot := sr.NewStreamSlot()

	for {
		err = ReadFrameToSlot(ws, slot, boundary)
		if err != nil {
			log.Println(err)
			return err
		}
		fmt.Println(">4", slot)
	}

	return err
}

//---------------------------------------------------------------------------
// read stream to ring buffer
//---------------------------------------------------------------------------
func ReadStreamToRing(ws *websocket.Conn, ring *sr.StreamRing, boundary string) error {
	var err error

	err = ring.SetStatusUsing()
	if err != nil {
		return sb.ErrStatus
	}
	defer ring.Reset()

	for ring.IsUsing() {
		slot, pos := ring.GetSlotIn()

		err = ReadFrameToSlot(ws, slot, boundary)
		if err != nil {
			log.Println(err)
			return err
		}

		ring.SetPosInByPos(pos + 1)
	}

	return err
}

//---------------------------------------------------------------------------
// read frame to slot
//---------------------------------------------------------------------------
func ReadFrameToSlot(ws *websocket.Conn, slot *sr.StreamSlot, boundary string) error {
	var err error

	nl, err := ReadFrameHeaderToSlot(ws, slot)
	if err != nil {
		log.Println(err)
		return err
	}

	err = ReadFrameBodyToSlot(ws, slot, nl)
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Println(">2", slot)
	return err
}

//---------------------------------------------------------------------------
// read a frame header and parse it
//---------------------------------------------------------------------------
func ReadFrameHeaderToSlot(ws *websocket.Conn, slot *sr.StreamSlot) (int, error) {
	var err error

	var rmsg string
	err = websocket.Message.Receive(ws, &rmsg)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	//fmt.Println(color.BlueString(string(rmsg)))

	r := bufio.NewReader(strings.NewReader(rmsg))
	nl := ParseFrameHeader(r, slot)

	return nl, err
}

//---------------------------------------------------------------------------
// parse frame header
//---------------------------------------------------------------------------
func ParseFrameHeader(r *bufio.Reader, slot *sr.StreamSlot) int {
	var nl int

	res := make(map[string]string)

	fstart := false
	for i := 0; ; i++ {
		line, _, err := r.ReadLine()
		if err != nil {
			log.Println(err)
			return nl
		}

		// find header start of part
		if string(line) == "" {
			if fstart {
				break
			} else {
				continue
			}
		}
		if !fstart {
			// for compatibility with agilecam shooter
			if strings.Contains(string(line), "--") || strings.Contains(string(line), "POST") {
				fstart = true
			}
		}

		// if maybe invalid header data, TODO: stop or ignore?
		if len(line) > sb.LEN_MAX_LINE {
			break
		}
		//fmt.Println(string(line))

		keyvalue := strings.SplitN(string(line), ":", 2)
		if len(keyvalue) > 1 {
			res[keyvalue[0]] = strings.TrimSpace(keyvalue[1])
		}
	}

	slot.Timestamp = sb.GetTimestampNow()

	slot.Type = res[sb.STR_HDR_CONTENT_TYPE]
	sl := res[sb.STR_HDR_CONTENT_LENGTH]
	nl, _ = strconv.Atoi(sl)

	return nl
}

//---------------------------------------------------------------------------
// read a frame header and parse it
//---------------------------------------------------------------------------
func ReadFrameBodyToSlot(ws *websocket.Conn, slot *sr.StreamSlot, nl int) error {
	var err error

	// ignore if 0 length of content
	if nl == 0 {
		return nil
	}

	err = websocket.Message.Receive(ws, &slot.Content)
	if err != nil {
		log.Println(err)
		return err
	}

	slot.Length = nl

	return err
}

//---------------------------------------------------------------------------
// recv multipart to ring buffer
//---------------------------------------------------------------------------
func WriteRingInStream(ws *websocket.Conn, ring *sr.StreamRing) error {
	var err error

	if !ring.IsUsing() {
		log.Println(sb.RedString("ErrStatus"))
		return sb.ErrStatus
	}

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

		err = WriteSlotInFrame(ws, slot, ring.Boundary)
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
// send a frame of stream
//---------------------------------------------------------------------------
func WriteDataInFrame(ws *websocket.Conn, data []byte, dsize int, dtype string, boundary string) error {
	var err error

	// prepare a slot and its data
	slot := sr.NewStreamSlotBySize(dsize)
	defer fmt.Println("1>", slot)

	slot.Type = dtype
	slot.Length = dsize
	slot.Timestamp = sb.GetTimestampNow()
	copy(slot.Content, data)

	// send the slot
	err = WriteSlotInFrame(ws, slot, boundary)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// send a frame of stream
//---------------------------------------------------------------------------
func WriteSlotInFrame(ws *websocket.Conn, slot *sr.StreamSlot, boundary string) error {
	var err error

	smsg := fmt.Sprintf("\r\n--%s\r\n", boundary)
	smsg += fmt.Sprintf("Content-Type: %s\r\n", slot.Type)
	smsg += fmt.Sprintf("Content-Length: %d\r\n", slot.Length)
	//smsg += fmt.Sprintf("X-Audio-Format: format=pcm_16; channel=1; frequency=44100\r\n")
	smsg += "\r\n"

	err = websocket.Message.Send(ws, smsg)
	if err != nil {
		log.Println(err)
		return err
	}

	err = websocket.Message.Send(ws, slot.Content[:slot.Length])
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//===========================================================================
// using multipart package, "mime/multipart"
//---------------------------------------------------------------------------
// recv multipart (for debugging)
//---------------------------------------------------------------------------
func ReadMultipartToData(ws *websocket.Conn, boundary string) error {
	var err error

	slot := sr.NewStreamSlot()
	mr := multipart.NewReader(ws, boundary)

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
func ReadMultipartToRing(ws *websocket.Conn, ring *sr.StreamRing) error {
	var err error

	err = ring.SetStatusUsing()
	if err != nil {
		return sb.ErrStatus
	}
	defer ring.Reset()

	mr := multipart.NewReader(ws, ring.Boundary)

	for ring.IsUsing() {
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

	//fmt.Println(">2", slot)
	return err
}

//---------------------------------------------------------------------------
// read a part header and parse it
//---------------------------------------------------------------------------
func ParsePartHeaderToSlot(p *multipart.Part, slot *sr.StreamSlot) (int, error) {
	var err error

	// get header information of each part
	stamp := p.Header.Get(sb.STR_HDR_TIMESTAMP)
	if stamp == "" {
		slot.Timestamp = sb.GetTimestampFromString(stamp)
	} else {
		slot.Timestamp = sb.GetTimestampNow()
	}

	slot.Type = p.Header.Get(sb.STR_HDR_CONTENT_TYPE)

	sl := p.Header.Get(sb.STR_HDR_CONTENT_LENGTH)
	nl, err := strconv.Atoi(sl)
	if err != nil {
		log.Printf("%s %s -> %d\n", p.Header, sl, nl)
		return nl, err
	}
	slot.Length = nl

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
