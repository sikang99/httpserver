//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for TCP Socket
//  - http://stackoverflow.com/questions/25090690/how-to-write-a-proxy-in-go-golang-using-tcp-connections
//  - https://github.com/nf/gohttptun - A tool to tunnel TCP over HTTP, written in Go
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

	err = conn.SetNoDelay(true)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Connecting to %s\n", addr)

	_, err = pt.SendRequestPost(conn)
	if err != nil {
		log.Println(err)
		return
	}

	// send multipart stream of files
	err = pt.SendMultipartFiles(conn, "../../static/image/*", true)

	return
}

//---------------------------------------------------------------------------
// TCP receiver for debugging
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ActServer(sbuf *sr.StreamRing) {
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

		go pt.HandleRequest(conn, sbuf)
	}
}

//---------------------------------------------------------------------------
// TCP Player to receive data in multipart
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ActPlayer(sbuf *sr.StreamRing) {
	log.Printf("%s\n", STR_TCP_PLAYER)

	addr, _ := net.ResolveTCPAddr("tcp", pt.Host+":"+pt.Port)
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

	_, err = pt.SendRequestGet(conn)
	if err != nil {
		log.Println(err)
		return
	}
	//fmt.Println(headers)

	// recv multipart stream from server
	err = pt.RecvMultipartToData(conn)
	if err != nil {
		log.Println(err)
		return
	}

	return
}

//---------------------------------------------------------------------------
// summit a TCP GET request
//---------------------------------------------------------------------------
func (pt *ProtoTcp) SendRequestGet(conn net.Conn) (map[string]string, error) {
	var err error

	// send GET request
	req := "GET /stream HTTP/1.1\r\n"
	req += fmt.Sprintf("User-Agent: %s\r\n", STR_TCP_PLAYER)
	req += "\r\n"

	log.Printf("SEND [%d]\n%s", len(req), req)

	_, err = conn.Write([]byte(req))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// recv response
	_, err = pt.RecvMessage(conn)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return nil, err
}

//---------------------------------------------------------------------------
// summit a TCP POST request
//---------------------------------------------------------------------------
func (pt *ProtoTcp) SendRequestPost(conn net.Conn) (map[string]string, error) {
	var err error

	// send POST request
	req := "POST /stream HTTP/1.1\r\n"
	req += fmt.Sprintf("Content-Type: multipart/x-mixed-replace; boundary=%s\r\n", pt.Boundary)
	req += fmt.Sprintf("User-Agent: %s\r\n", STR_TCP_CASTER)
	req += "\r\n"

	log.Printf("SEND [%d]\n%s", len(req), req)

	_, err = conn.Write([]byte(req))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// recv response
	_, err = pt.RecvMessage(conn)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return nil, err
}

//---------------------------------------------------------------------------
// handle request, please close socket after use
//---------------------------------------------------------------------------
func (pt *ProtoTcp) HandleRequest(conn net.Conn, sbuf *sr.StreamRing) error {
	var err error

	log.Printf("from client at %s\n", conn.RemoteAddr())
	defer log.Printf("out %s\n", conn.RemoteAddr())
	defer conn.Close()

	// recv request and parse
	_, err = pt.RecvMessage(conn)
	if err != nil {
		log.Println(err)
		return err
	}

	if pt.Method == "POST" {
		err = pt.SendResponsePost(conn)
		err = pt.RecvMultipartToRing(conn, sbuf)
	} else if pt.Method == "GET" {
		err = pt.SendResponseGet(conn)
		err = pt.SendRingToMultipart(conn, sbuf)
	} else {
		err = sb.ErrSupport
	}
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// send ring buffer to client in multipart
//---------------------------------------------------------------------------
func (pt *ProtoTcp) SendRingToMultipart(conn net.Conn, sbuf *sr.StreamRing) error {
	var err error

	if !sbuf.IsUsing() {
		return sb.ErrStatus
	}

	var pos int
	for {
		slot, npos, err := sbuf.GetSlotNextByPos(pos)
		if err != nil {
			if err == sb.ErrEmpty {
				time.Sleep(sb.TIME_DEF_WAIT)
				continue
			}
			log.Println(err)
			break
		}

		err = pt.SendPartSlot(conn, slot)
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
// recv multipart to ring buffer
//---------------------------------------------------------------------------
func (pt *ProtoTcp) RecvMultipartToRing(conn net.Conn, sbuf *sr.StreamRing) error {
	var err error

	err = sbuf.SetStatusUsing()
	if err != nil {
		return sb.ErrStatus
	}
	defer sbuf.Reset()

	//  recv multipart stream
	for {
		slot, pos := sbuf.GetSlotIn()
		_, err = pt.ReadPartToSlot(conn, slot)
		if err != nil {
			log.Println(err)
			return err
		}
		//fmt.Println(slot)
		if slot.IsMajorType("multipart") {
			log.Println(slot)
			continue
		}
		sbuf.SetPosInByPos(pos + 1)
	}

	fmt.Println(sbuf)
	return err
}

//---------------------------------------------------------------------------
// recv multipart to data
//---------------------------------------------------------------------------
func (pt *ProtoTcp) RecvMultipartToData(conn net.Conn) error {
	var err error

	slot := sr.NewStreamSlot()

	for i := 0; ; i++ {
		_, err = pt.ReadPartToSlot(conn, slot)
		if err != nil {
			log.Println(err)
			return err
		}
		fmt.Println(i, slot)
	}

	return err
}

//---------------------------------------------------------------------------
// send TCP response for POST request
//---------------------------------------------------------------------------
func (pt *ProtoTcp) SendResponsePost(conn net.Conn) error {
	var err error

	res := "HTTP/1.1 200 Ok\r\n"
	res += fmt.Sprintf("Server: %s\r\n", STR_TCP_SERVER)
	res += "\r\n"

	defer log.Printf("SEND [%d]\n%s", len(res), res)

	_, err = conn.Write([]byte(res))
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// send TCP response for GET request
//---------------------------------------------------------------------------
func (pt *ProtoTcp) SendResponseGet(conn net.Conn) error {
	var err error

	res := "HTTP/1.1 200 Ok\r\n"
	res += fmt.Sprintf("Server: %s\r\n", STR_TCP_SERVER)
	res += fmt.Sprintf("Content-Type: multipart/x-mixed-replace; boundary=%s\r\n", pt.Boundary)
	res += "\r\n"

	defer log.Printf("SEND [%d]\n%s", len(res), res)

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
func (pt *ProtoTcp) ReadPartToSlot(conn net.Conn, ss *sr.StreamSlot) (map[string]string, error) {
	var err error

	reader := bufio.NewReader(conn)

	headers, err := pt.ReadPartHeader(reader)
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
		//_, err = pt.ReadMessageBodyToData(reader, clen)
		ss.Type = headers["CONTENT-TYPE"]
		err = pt.ReadMessageBodyToSlot(reader, clen, ss)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	return headers, err
}

//---------------------------------------------------------------------------
// read part data
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ReadPartData(conn net.Conn) error {
	var err error

	fmt.Println("TODO")

	return err
}

//---------------------------------------------------------------------------
// get boundary string for multipart
//---------------------------------------------------------------------------
func (pt *ProtoTcp) GetBoundary(str string) error {
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
// peek a message
//---------------------------------------------------------------------------
func (pt *ProtoTcp) PeekMessage(conn net.Conn) error {
	var err error

	return err
}

//---------------------------------------------------------------------------
// read a message in http header style
//---------------------------------------------------------------------------
func (pt *ProtoTcp) RecvMessage(conn net.Conn) (map[string]string, error) {
	var err error

	reader := bufio.NewReader(conn)

	headers, err := pt.ReadMessageHeader(reader)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	value, ok := headers["CONTENT-TYPE"]
	if ok {
		pt.GetBoundary(value)
	}

	clen := 0
	value, ok = headers["CONTENT-LENGTH"]
	if ok {
		clen, _ = strconv.Atoi(value)
	}

	if clen > 0 {
		_, err = pt.ReadMessageBodyToData(reader, clen)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	//fmt.Println(headers)
	return headers, err
}

//---------------------------------------------------------------------------
// read header of message and return a map
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ReadPartHeader(reader *bufio.Reader) (map[string]string, error) {
	var err error

	result := make(map[string]string)

	fstart := false
	for i := 0; ; i++ {
		line, _, err := reader.ReadLine()
		if err != nil {
			log.Println(err)
			return result, err
		}
		/*
			if i == 0 {
				log.Printf("%x\n", line[:4])
			}
		*/
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
			if strings.Contains(string(line), pt.Boundary) || strings.Contains(string(line), "POST") {
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
			result[strings.ToUpper(keyvalue[0])] = strings.TrimSpace(keyvalue[1])
		}
	}

	//fmt.Println(result)
	return result, err
}

//---------------------------------------------------------------------------
// read header of message and return a map
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ReadMessageHeader(reader *bufio.Reader) (map[string]string, error) {
	var err error

	result := make(map[string]string)

	// parse a request line
	line, _, err := reader.ReadLine()
	if err != nil {
		log.Println(err)
		return result, err
	}

	res := strings.Fields(string(line))
	pt.Method = res[0]

	// parse header lines
	for {
		line, _, err := reader.ReadLine()
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
			result[strings.ToUpper(keyvalue[0])] = strings.TrimSpace(keyvalue[1])
		}
	}

	//fmt.Println(result)
	return result, err
}

//---------------------------------------------------------------------------
// read(recv) body of message to data
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ReadMessageBodyToData(reader *bufio.Reader, clen int) ([]byte, error) {
	var err error

	data := make([]byte, clen)

	tn := 0
	for tn < clen {
		n, err := reader.Read(data[tn:])
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
func (pt *ProtoTcp) ReadMessageBodyToSlot(reader *bufio.Reader, clen int, ss *sr.StreamSlot) error {
	var err error

	ss.Length = 0

	tn := 0
	for tn < clen {
		n, err := reader.Read(ss.Content[tn:])
		if err != nil {
			log.Println(err)
			break
		}
		tn += n
	}

	ss.Length = clen

	//fmt.Printf("[DATA] (%d/%d)\n\n", tn, clen)
	return err
}

//---------------------------------------------------------------------------
// send files in multipart
//---------------------------------------------------------------------------
func (pt *ProtoTcp) SendMultipartFiles(conn net.Conn, pattern string, loop bool) error {
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
			err = pt.SendPartFile(conn, files[i])
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
// send a part file
//---------------------------------------------------------------------------
func (pt *ProtoTcp) SendPartFile(conn net.Conn, file string) error {
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

	err = pt.SendPartData(conn, fdata, ctype)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

//---------------------------------------------------------------------------
// send a part data
//---------------------------------------------------------------------------
func (pt *ProtoTcp) SendPartData(conn net.Conn, data []byte, ctype string) error {
	var err error

	clen := len(data)

	req := fmt.Sprintf("--%s\r\n", pt.Boundary)
	req += fmt.Sprintf("Content-Type: %s\r\n", ctype)
	req += fmt.Sprintf("Content-Length: %d\r\n", clen)
	req += fmt.Sprintf("x-Timestamp: %v\r\n", sb.GetTimestamp())
	req += "\r\n"

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

	//fmt.Printf("SEND [%d,%d]\n%s", len(req), clen, req)
	return err
}

//---------------------------------------------------------------------------
// send a part slot of ring buffer
//---------------------------------------------------------------------------
func (pt *ProtoTcp) SendPartSlot(conn net.Conn, ss *sr.StreamSlot) error {
	var err error

	req := fmt.Sprintf("--%s\r\n", pt.Boundary)
	req += fmt.Sprintf("Content-Type: %s\r\n", ss.Type)
	req += fmt.Sprintf("Content-Length: %d\r\n", ss.Length)
	req += fmt.Sprintf("x-Timestamp: %v\r\n", ss.Timestamp)
	req += "\r\n"

	_, err = conn.Write([]byte(req))
	if err != nil {
		log.Println(err)
		return err
	}

	if ss.Length > 0 {
		_, err = conn.Write(ss.Content[:ss.Length])
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return err
}

//---------------------------------------------------------------------------
// read a part header and parse it
//---------------------------------------------------------------------------
func ReadMultipartHeader(mr *multipart.Reader) (*multipart.Part, int, error) {
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

// ---------------------------------E-----N-----D--------------------------------
