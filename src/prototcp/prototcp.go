//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for TCP Socket
//  - http://stackoverflow.com/questions/25090690/how-to-write-a-proxy-in-go-golang-using-tcp-connections
//  - https://github.com/nf/gohttptun - A tool to tunnel TCP over HTTP, written in Go
//  - http://www.slideshare.net/feyeleanor/go-for-the-paranoid-network-programmer
//==================================================================================

package prototcp

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	ph "stoney/httpserver/src/protohttp"

	sb "stoney/httpserver/src/streambase"
	sr "stoney/httpserver/src/streamring"
)

//---------------------------------------------------------------------------
const (
	STR_TCP_CASTER = "Happy Media TCP Caster"
	STR_TCP_SERVER = "Happy Media TCP Server"
	STR_TCP_PLAYER = "Happy Media TCP Player"
)

//---------------------------------------------------------------------------
type ProtoTcp struct {
	Host     string
	Port     string
	Desc     string
	Method   string // POST or GET
	Boundary string
	Conn     net.Conn
}

//---------------------------------------------------------------------------
// string ProtoTcp information
//---------------------------------------------------------------------------
func (pt *ProtoTcp) String() string {
	str := fmt.Sprintf("\tHost: %s", pt.Host)
	str += fmt.Sprintf("\tPort: %s", pt.Port)
	str += fmt.Sprintf("\tConn: %v", pt.Conn)
	str += fmt.Sprintf("\tBoundary: %s", pt.Boundary)
	str += fmt.Sprintf("\tDesc: %s", pt.Desc)
	return str
}

//---------------------------------------------------------------------------
// info handling
//---------------------------------------------------------------------------
func (pt *ProtoTcp) SetAddr(hname, hport, desc string) {
	pt.Host = hname
	pt.Port = hport
	pt.Desc = desc
}

func (pt *ProtoTcp) Reset() {
	pt.Host = sb.STR_DEF_HOST
	pt.Port = sb.STR_DEF_PORT
	pt.Boundary = sb.STR_DEF_BDRY
	pt.Desc = "reset"
	if pt.Conn != nil {
		pt.Conn.Close()
		pt.Conn = nil
	}
}

func (pt *ProtoTcp) Clear() {
	pt.Host = ""
	pt.Port = ""
	pt.Desc = ""
	if pt.Conn != nil {
		pt.Conn.Close()
		pt.Conn = nil
	}
}

//---------------------------------------------------------------------------
// new ProtoTcp struct
//---------------------------------------------------------------------------
func NewProtoTcp(hname, hport, desc string) *ProtoTcp {
	return &ProtoTcp{
		Host:     hname,
		Port:     hport,
		Desc:     desc,
		Boundary: sb.STR_DEF_BDRY,
	}
}

//---------------------------------------------------------------------------
// act TCP sender for test and debugging
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ActCaster() {
	log.Printf("%s to %s:%s\n", STR_TCP_CASTER, pt.Host, pt.Port)

	addr, _ := net.ResolveTCPAddr("tcp", pt.Host+":"+pt.Port)
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	log.Printf("Caster> connected to %s\n", addr)

	err = conn.SetNoDelay(true)
	if err != nil {
		log.Println(err)
		return
	}

	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	err = pt.RequestPost(r, w)
	if err != nil {
		log.Println(err)
		return
	}

	// send multipart stream of files
	err = pt.WriteStreamFiles(w, "../../static/image/*", true)

	return
}

//---------------------------------------------------------------------------
// TCP receiver for debugging
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ActServer(ring *sr.StreamRing) {
	log.Printf("%s on :%s\n", STR_TCP_SERVER, pt.Port)

	l, err := net.Listen("tcp", ":"+pt.Port)
	if err != nil {
		log.Println(err)
		return
	}
	defer l.Close()

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			return
		}

		go pt.HandleRequest(conn, ring)
	}
}

//---------------------------------------------------------------------------
// TCP Player to receive data in multipart
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ActPlayer(ring *sr.StreamRing) {
	log.Printf("%s\n", STR_TCP_PLAYER)

	addr, _ := net.ResolveTCPAddr("tcp", pt.Host+":"+pt.Port)
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	log.Printf("Player> connected to %s\n", addr)

	err = conn.SetNoDelay(true)
	if err != nil {
		log.Println(err)
		return
	}

	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	err = pt.RequestGet(r, w)
	if err != nil {
		log.Println(err)
		return
	}
	//fmt.Println(headers)

	// recv multipart stream from server
	err = pt.ReadStreamToData(r)
	//err = RecvStreamToData(conn)
	if err != nil {
		log.Println(err)
		return
	}

	return
}

//---------------------------------------------------------------------------
// summit a GET request and get its response
//---------------------------------------------------------------------------
func (pt *ProtoTcp) RequestGet(r *bufio.Reader, w *bufio.Writer) error {
	var err error

	// send GET request
	req := "GET /stream HTTP/1.1\r\n"
	req += fmt.Sprintf("User-Agent: %s\r\n", STR_TCP_PLAYER)
	req += "\r\n"

	_, err = w.Write([]byte(req))
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("SEND [%d]\n%s", len(req), req)

	// recv response
	err = pt.ReadMessage(r)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// summit a POST request and get its response
//---------------------------------------------------------------------------
func (pt *ProtoTcp) RequestPost(r *bufio.Reader, w *bufio.Writer) error {
	var err error

	// send POST request
	req := "POST /stream HTTP/1.1\r\n"
	req += fmt.Sprintf("Content-Type: multipart/x-mixed-replace; boundary=%s\r\n", pt.Boundary)
	req += fmt.Sprintf("User-Agent: %s\r\n", STR_TCP_CASTER)
	req += "\r\n"

	_, err = w.Write([]byte(req))
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("SEND [%d]\n%s", len(req), req)

	// recv response
	err = pt.ReadMessage(r)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// handle request, please close socket after use
//---------------------------------------------------------------------------
func (pt *ProtoTcp) HandleRequest(conn net.Conn, ring *sr.StreamRing) error {
	var err error

	log.Printf("Server> in from %s\n", conn.RemoteAddr())
	defer log.Printf("Server> out from %s\n", conn.RemoteAddr())
	defer conn.Close()

	// change conn into bufio handler
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	// recv request and parse it
	err = pt.ReadMessage(r)
	if err != nil {
		log.Println(err)
		return err
	}

	// send response and multipart
	switch pt.Method {
	case "POST":
		err = pt.ResponsePost(w)
		if err != nil {
			log.Println(err)
			return err
		}
		err = pt.ReadStreamToRing(r, ring)
		if err != nil {
			log.Println(err)
			return err
		}
	case "GET":
		err = pt.ResponseGet(w)
		if err != nil {
			log.Println(err)
			return err
		}
		//err = pt.WriteRingToStream(w, ring)
		err = pt.WriteDataToStream(w, ring)
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
// send slot directly for debugging
//---------------------------------------------------------------------------
func (pt *ProtoTcp) WriteDataToStream(w *bufio.Writer, ring *sr.StreamRing) error {
	var err error

	for pos := 0; ; pos++ {
		slot, _ := ring.GetSlotByPos(pos)
		slot.Timestamp = sb.GetTimestamp()

		err = pt.WriteFrameSlot(w, slot)
		if err != nil {
			log.Println(err)
			break
		}
		log.Println(">", slot)

		time.Sleep(time.Second)
	}

	return err
}

//---------------------------------------------------------------------------
// send ring buffer to client in multipart
//---------------------------------------------------------------------------
func (pt *ProtoTcp) WriteRingToStream(w *bufio.Writer, ring *sr.StreamRing) error {
	var err error

	if !ring.IsUsing() {
		log.Println("ErrStatus")
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

		err = pt.WriteFrameSlot(w, slot)
		if err != nil {
			log.Println(err)
			break
		}
		log.Println(slot)

		pos = npos
	}

	return err
}

//---------------------------------------------------------------------------
// recv multipart to ring buffer
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ReadStreamToRing(r *bufio.Reader, ring *sr.StreamRing) error {
	var err error

	err = ring.SetStatusUsing()
	if err != nil {
		return sb.ErrStatus
	}
	defer ring.Reset()

	//  recv multipart stream
	for {
		slot, pos := ring.GetSlotIn()
		err = pt.ReadFrameToSlot(r, slot)
		if err != nil {
			log.Println(err)
			return err
		}
		//fmt.Println(slot)
		if slot.IsMajorType("multipart") {
			log.Println(slot)
			continue
		}
		ring.SetPosInByPos(pos + 1)
	}

	fmt.Println(ring)
	return err
}

//---------------------------------------------------------------------------
// recv multipart to data
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ReadStreamToData(r *bufio.Reader) error {
	var err error

	mr := multipart.NewReader(r, "myboundary")
	err = ph.RecvMultipartToData(mr)

	return err
}

//---------------------------------------------------------------------------
// send TCP response for POST request
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ResponsePost(w *bufio.Writer) error {
	var err error

	res := "HTTP/1.1 200 Ok\r\n"
	res += fmt.Sprintf("Server: %s\r\n", STR_TCP_SERVER)
	res += "\r\n"

	defer log.Printf("SEND [%d]\n%s", len(res), res)

	_, err = w.Write([]byte(res))
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// send TCP response for GET request
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ResponseGet(w *bufio.Writer) error {
	var err error

	res := "HTTP/1.1 200 Ok\r\n"
	res += fmt.Sprintf("Server: %s\r\n", STR_TCP_SERVER)
	res += fmt.Sprintf("Content-Type: multipart/x-mixed-replace; boundary=%s\r\n", pt.Boundary)
	res += "\r\n"

	defer log.Printf("SEND [%d]\n%s", len(res), res)

	_, err = w.Write([]byte(res))
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// read a message in http header style
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ReadFrameToSlot(r *bufio.Reader, ss *sr.StreamSlot) error {
	var err error

	headers, err := ParseFrameHeader(r)
	if err != nil {
		log.Println(err)
		return err
	}

	clen := 0
	value, ok := headers[sb.STR_HDR_LENGTH]
	if ok {
		clen, _ = strconv.Atoi(value)
	}

	if clen > 0 {
		ss.Type = headers[sb.STR_HDR_TYPE]
		err = pt.ReadMessageBodyToSlot(r, clen, ss)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return err
}

//---------------------------------------------------------------------------
// get boundary string for multipart
//---------------------------------------------------------------------------
func (pt *ProtoTcp) GetTypeBoundary(str string) error {
	var err error

	_, params, err := mime.ParseMediaType(str)
	//fmt.Printf("%v %v %s\n", mt, params, str)
	if err != nil {
		log.Println(err)
		return err
	}

	boundary := params["boundary"]
	if !strings.HasPrefix(boundary, "--") {
		log.Printf("expected with --, got %q", boundary)
	}

	pt.Boundary = boundary
	//fmt.Println(pt)

	return err
}

//---------------------------------------------------------------------------
// read a message in http header style
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ReadMessage(r *bufio.Reader) error {
	var err error

	headers, err := pt.ReadMessageHeader(r)
	if err != nil {
		log.Println(err)
		return err
	}

	value, ok := headers[sb.STR_HDR_TYPE]
	if ok {
		pt.GetTypeBoundary(value)
	}

	clen := 0
	value, ok = headers[sb.STR_HDR_LENGTH]
	if ok {
		clen, _ = strconv.Atoi(value)
	}

	if clen > 0 {
		_, err = pt.ReadMessageBodyToData(r, clen)
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
func (pt *ProtoTcp) ReadMessageHeader(r *bufio.Reader) (map[string]string, error) {
	var err error

	result := make(map[string]string)

	// parse a request line
	line, _, err := r.ReadLine()
	if err != nil {
		log.Println(err)
		return result, err
	}

	res := strings.Fields(string(line))
	pt.Method = res[0]

	// parse header lines
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			log.Println(err)
			return result, err
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
func (pt *ProtoTcp) ReadMessageBodyToData(r *bufio.Reader, clen int) ([]byte, error) {
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
// read(recv) body of message to slot of ring buffer
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ReadMessageBodyToSlot(r *bufio.Reader, clen int, ss *sr.StreamSlot) error {
	var err error

	ss.Length = 0

	tn := 0
	for tn < clen {
		n, err := r.Read(ss.Content[tn:])
		if err != nil {
			log.Println(err)
			break
		}
		tn += n
	}

	ss.Length = clen
	ss.Timestamp = sb.GetTimestamp()

	//fmt.Printf("[DATA] (%d/%d)\n\n", tn, clen)
	return err
}

//---------------------------------------------------------------------------
// send files in multipart
//---------------------------------------------------------------------------
func (pt *ProtoTcp) WriteStreamFiles(w *bufio.Writer, pattern string, loop bool) error {
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
			err = pt.WriteFrameFile(w, files[i])
			if err != nil {
				if err == sb.ErrSize {
					continue
				}
				log.Println(err)
				return err
			}

			//time.Sleep(time.Millisecond)
			time.Sleep(time.Second)
		}

		if !loop {
			break
		}
	}

	return err
}

//---------------------------------------------------------------------------
// write a part file
//---------------------------------------------------------------------------
func (pt *ProtoTcp) WriteFrameFile(w *bufio.Writer, file string) error {
	var err error

	fdata, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println(err)
		return err
	}

	if len(fdata) == 0 {
		fmt.Printf(">> ignore '%s' of zero size\n", file)
		return sb.ErrSize
	}

	ctype := mime.TypeByExtension(file)
	if ctype == "" {
		ctype = http.DetectContentType(fdata)
	}

	err = pt.WriteFrameData(w, fdata, ctype)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// write a part data
//---------------------------------------------------------------------------
func (pt *ProtoTcp) WriteFrameData(w *bufio.Writer, data []byte, ctype string) error {
	var err error

	clen := len(data)

	req := fmt.Sprintf("\r\n--%s\r\n", pt.Boundary)
	req += fmt.Sprintf("Content-Type: %s\r\n", ctype)
	req += fmt.Sprintf("Content-Length: %d\r\n", clen)
	req += fmt.Sprintf("x-Timestamp: %v\r\n", sb.GetTimestamp())
	req += "\r\n"

	_, err = w.Write([]byte(req))
	if err != nil {
		log.Println(err)
		return err
	}

	if clen > 0 {
		_, err = w.Write(data)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return err
}

//---------------------------------------------------------------------------
// send a part slot of ring buffer
//---------------------------------------------------------------------------
func (pt *ProtoTcp) WriteFrameSlot(w *bufio.Writer, ss *sr.StreamSlot) error {
	var err error

	// make frame header
	req := fmt.Sprintf("\r\n--%s\r\n", pt.Boundary)
	req += fmt.Sprintf("Content-Type: %s\r\n", ss.Type)
	req += fmt.Sprintf("Content-Length: %d\r\n", ss.Length)
	req += fmt.Sprintf("x-Timestamp: %v\r\n", ss.Timestamp)
	req += "\r\n"

	// send frame header
	_, err = w.Write([]byte(req))
	if err != nil {
		log.Println(err)
		return err
	}

	// send frame body
	if ss.Length > 0 {
		_, err = w.Write(ss.Content[:ss.Length])
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return err
}

//---------------------------------------------------------------------------
// recv http message
//---------------------------------------------------------------------------
func RecvMessage(conn net.Conn) error {
	var err error

	return err
}

//===========================================================================
// functions using direct socket io
//===========================================================================
// stream = [frame][frame]...
// frame  = [header][body]
// header = [(text)\r\n\r\n]
//---------------------------------------------------------------------------
// recv stream to ring buffer
//---------------------------------------------------------------------------
func RecvStreamToRing(conn net.Conn, ring *sr.StreamRing) error {
	var err error

	err = ring.SetStatusUsing()
	if err != nil {
		return sb.ErrStatus
	}
	defer ring.Reset()

	// recv a stream
	for {
		slot, pos := ring.GetSlotIn()
		err = RecvFrameToSlot(conn, slot)
		if err != nil {
			log.Println(err)
			return err
		}
		//fmt.Println(slot)
		if slot.IsMajorType("multipart") {
			log.Println(slot)
			continue
		}
		ring.SetPosInByPos(pos + 1)
	}

	fmt.Println(ring)
	return err
}

//---------------------------------------------------------------------------
// recv stream to data for debugging
//---------------------------------------------------------------------------
func RecvStreamToData(conn net.Conn) error {
	var err error

	for i := 0; ; i++ {
		err = RecvFrameToData(conn)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return err
}

//---------------------------------------------------------------------------
// recv a frame to slot
//---------------------------------------------------------------------------
func RecvFrameToSlot(conn net.Conn, ss *sr.StreamSlot) error {
	var err error

	headers, err := RecvFrameHeader(conn)
	if err != nil {
		log.Println(err)
		return err
	}

	clen := 0
	value, ok := headers[sb.STR_HDR_LENGTH]
	if ok {
		clen, _ = strconv.Atoi(value)
	}

	if clen > 0 {
		ss.Type = headers[sb.STR_HDR_TYPE]
		err = RecvFrameBodyToSlot(conn, clen, ss)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return err
}

//---------------------------------------------------------------------------
// recv frame (header + body) for debugging
//---------------------------------------------------------------------------
func RecvFrameToData(conn net.Conn) error {
	var err error

	headers, err := RecvFrameHeader(conn)
	if err != nil {
		log.Println(err)
		return err
	}

	tstamp, ok := headers[sb.STR_HDR_TSTAMP]
	if ok {
		log.Println("<", tstamp)
	}

	clen := 0
	value, ok := headers[sb.STR_HDR_LENGTH]
	if ok {
		clen, _ = strconv.Atoi(value)
	}

	if clen > 0 {
		err = RecvFrameBodyToData(conn, clen)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return err
}

//---------------------------------------------------------------------------
// recv frame header
//---------------------------------------------------------------------------
func RecvFrameHeader(conn net.Conn) (map[string]string, error) {
	var err error

	hstr, err := RecvFrameHeaderString(conn)
	//log.Println(hstr)

	reader := bufio.NewReader(strings.NewReader(hstr))
	result, err := ParseFrameHeader(reader)

	return result, err
}

//---------------------------------------------------------------------------
// recv headers (frame header or message) ended with "\r\n\r\n"
//---------------------------------------------------------------------------
func RecvFrameHeaderString(conn net.Conn) (string, error) {
	var err error

	line := make([]byte, sb.LEN_MAX_LINE)

	var tn int
	for {
		n, err := conn.Read(line[tn:sb.LEN_MAX_LINE])
		if err != nil {
			log.Println(err)
			break
		}
		if string(line[:2]) != "\r\n" && string(line[:2]) != "--" {
			log.Println("invalid frame start")
			continue
		}
		tn += n
		if tn > 3 && string(line[tn-4:tn]) == "\r\n\r\n" {
			//log.Println(string(line[:tn]))
			break
		}
	}

	return string(line[:tn]), err
}

//---------------------------------------------------------------------------
// parse frame headers
//---------------------------------------------------------------------------
func ParseFrameHeader(reader *bufio.Reader) (map[string]string, error) {
	var err error

	result := make(map[string]string)

	fstart := false
	for i := 0; ; i++ {
		line, _, err := reader.ReadLine()
		if err != nil {
			log.Println(err)
			return result, err
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
			//result[strings.ToUpper(keyvalue[0])] = strings.TrimSpace(keyvalue[1])
			result[keyvalue[0]] = strings.TrimSpace(keyvalue[1])
		}
	}

	//fmt.Println(result)
	return result, err
}

//---------------------------------------------------------------------------
// recv frame body to data
//---------------------------------------------------------------------------
func RecvFrameBodyToData(conn net.Conn, clen int) error {
	var err error

	data := make([]byte, clen)

	tn := 0
	for tn < clen {
		n, err := conn.Read(data[tn:])
		if err != nil {
			log.Println(err)
			break
		}
		tn += n
	}

	fmt.Printf("[DATA] (%d/%d)\n\n", tn, clen)
	return err
}

//---------------------------------------------------------------------------
// recv frame body to data
//---------------------------------------------------------------------------
func RecvFrameBodyToSlot(conn net.Conn, clen int, ss *sr.StreamSlot) error {
	var err error

	tn := 0
	for tn < clen {
		n, err := conn.Read(ss.Content[tn:])
		if err != nil {
			log.Println(err)
			break
		}
		tn += n
	}

	return err
}

// ---------------------------------E-----N-----D--------------------------------
