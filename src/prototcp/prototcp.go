//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for TCP Socket
//  - http://stackoverflow.com/questions/25090690/how-to-write-a-proxy-in-go-golang-using-tcp-connections
//  - https://github.com/nf/gohttptun - A tool to tunnel TCP over HTTP, written in Go
//==================================================================================

package prototcp

import (
	"bufio"
	"errors"
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

	sr "stoney/httpserver/src/streamring"
)

//---------------------------------------------------------------------------
const (
	KBYTE = 1024
	MBYTE = 1024 * KBYTE

	LEN_MAX_LINE = 128

	STR_DEF_HOST = "localhost"
	STR_DEF_PORT = "8080"
	STR_DEF_BDRY = "myboundary"

	STR_PGM_SENDER = "Happy Media TCP Sender"
	STR_PGM_RECVER = "Happy Media TCP Receiver"
)

var (
	ErrSize = errors.New("Invalid size")
	ErrNone = errors.New("Nothing to do")
)

//---------------------------------------------------------------------------
type ProtoTcp struct {
	Host     string
	Port     string
	Desc     string
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
	pt.Host = STR_DEF_HOST
	pt.Port = STR_DEF_PORT
	pt.Boundary = STR_DEF_BDRY
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
		Boundary: STR_DEF_BDRY,
	}
}

//---------------------------------------------------------------------------
// act TCP sender for test and debugging
//---------------------------------------------------------------------------
func ActSender(pt *ProtoTcp) {
	log.Printf("Happy Media TCP Sender\n")

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

	headers, err := pt.SummitRequest(conn)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(headers)

	// send multipart stream of files
	err = pt.SendMultipartFiles(conn, "../../static/image/*")

	return
}

//---------------------------------------------------------------------------
// TCP receiver for debugging
//---------------------------------------------------------------------------
func ActReceiver(pt *ProtoTcp, sbuf *sr.StreamRing) {
	log.Printf("Happy Media TCP Receiver\n")

	l, err := net.Listen("tcp", ":"+pt.Port)
	if err != nil {
		log.Println(err)
		return
	}
	defer l.Close()

	log.Printf("TCP Server on :%s\n", pt.Port)

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
// summit a TCP request
//---------------------------------------------------------------------------
func (pt *ProtoTcp) SummitRequest(conn net.Conn) (map[string]string, error) {
	var err error

	// send POST request
	req := "POST /stream HTTP/1.1\r\n"
	req += fmt.Sprintf("Content-Type: multipart/x-mixed-replace; boundary=%s\r\n", pt.Boundary)
	req += "User-Agent: Happy Media TCP Client\r\n"
	req += "\r\n"

	fmt.Printf("SEND [%d]\n%s", len(req), req)

	_, err = conn.Write([]byte(req))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	_, err = pt.ReadMessage(conn)
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

	log.Printf("in from client at %s\n", conn.RemoteAddr())
	defer log.Printf("out %s\n", conn.RemoteAddr())
	defer conn.Close()

	// recv POST request
	_, err = pt.ReadMessage(conn)
	if err != nil {
		log.Println(err)
		return err
	}

	err = pt.SendResponse(conn)
	if err != nil {
		log.Println(err)
		return err
	}

	err = sbuf.SetStatusUsing()
	if err != nil {
		return sr.ErrStatus
	}
	defer sbuf.Reset()

	//  recv multipart stream
	//for i := 0; i < 20; i++ {
	for {
		slot, pos := sbuf.GetSlotIn()
		_, err = pt.ReadPart(conn, slot)
		if err != nil {
			log.Println(err)
			return err
		}
		if slot.IsMajorType("multipart") {
			continue
		}
		sbuf.SetPosInByPos(pos + 1)
	}

	fmt.Println(sbuf)
	return err
}

//---------------------------------------------------------------------------
// send TCP response
//---------------------------------------------------------------------------
func (pt *ProtoTcp) SendResponse(conn net.Conn) error {
	var err error

	res := "HTTP/1.1 200 Ok\r\n"
	res += "Server: Happy Media TCP Server\r\n"
	res += "\r\n"

	fmt.Printf("SEND [%d]\n%s", len(res), res)

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
func (pt *ProtoTcp) ReadPart(conn net.Conn, ss *sr.StreamSlot) (map[string]string, error) {
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
		log.Printf("expected boundary to start with --, got %q", boundary)
	}

	pt.Boundary = boundary
	fmt.Println(pt)

	return err
}

//---------------------------------------------------------------------------
// read a message in http header style
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ReadMessage(conn net.Conn) (map[string]string, error) {
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

	return headers, err
}

//---------------------------------------------------------------------------
// read header of message and return a map
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ReadPartHeader(reader *bufio.Reader) (map[string]string, error) {
	var err error

	result := make(map[string]string)

	fstart := false
	for {
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
			if strings.Contains(string(line), pt.Boundary) || strings.Contains(string(line), "POST") {
				fstart = true
			}
		}

		// if maybe invalid header data, TODO: stop or ignore?
		if len(line) > LEN_MAX_LINE {
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

	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			log.Println(err)
			return result, err
		}
		if string(line) == "" {
			break
		}

		fmt.Println(string(line))

		keyvalue := strings.SplitN(string(line), ":", 2)
		if len(keyvalue) > 1 {
			result[strings.ToUpper(keyvalue[0])] = strings.TrimSpace(keyvalue[1])
		}
	}

	fmt.Println(result)
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
func (pt *ProtoTcp) SendMultipartFiles(conn net.Conn, pattern string) error {
	var err error

	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Println(err)
		return err
	}
	if files == nil {
		log.Printf("no file for '%s'\n", pattern)
		return ErrNone
	}

	for i := range files {
		err = pt.SendPartFile(conn, files[i])
		if err != nil {
			if err == ErrSize {
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
		return ErrSize
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

	fmt.Printf("SEND [%d,%d]\n%s", len(req), clen, req)
	return err
}

//---------------------------------------------------------------------------
// read a part header and parse it
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ReadMultipartHeader(mr *multipart.Reader) (*multipart.Part, int, error) {
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
