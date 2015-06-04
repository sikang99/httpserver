//=================================================================================
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
	"os"
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
)

//---------------------------------------------------------------------------
type ProtoTcp struct {
	hname string
	hport string
}

//---------------------------------------------------------------------------
// string ProtoTcp information
//---------------------------------------------------------------------------
func (pt *ProtoTcp) String() string {
	str := fmt.Sprintf("Hostname: %s", pt.hname)
	str += fmt.Sprintf("Port: %s", pt.hport)
	return str
}

//---------------------------------------------------------------------------
// new ProtoTcp struct
//---------------------------------------------------------------------------
func NewProtoTcp() *ProtoTcp {
	return &ProtoTcp{}
}

//---------------------------------------------------------------------------
// act TCP sender for test and debugging
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ActSender(hname, hport string) {
	log.Printf("Happy Media TCP Sender\n")

	addr, _ := net.ResolveTCPAddr("tcp", hname+":"+hport)
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

	return
}

//---------------------------------------------------------------------------
// TCP receiver for debugging
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ActReceiver(hport string) {
	log.Printf("Happy Media TCP Receiver\n")

	l, err := net.Listen("tcp", ":"+hport)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer l.Close()

	log.Printf("TCP Server on :%s\n", hport)

	sbuf := sr.NewStreamRing(5, MBYTE)

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		go pt.HandleRequest(conn, sbuf)
	}
}

//---------------------------------------------------------------------------
// summit TCP request
//---------------------------------------------------------------------------
func (pt *ProtoTcp) SummitRequest(conn net.Conn) (map[string]string, error) {
	var err error

	boundary := "myboundary"

	// send POST request
	req := "POST /stream HTTP/1.1\r\n"
	req += fmt.Sprintf("Content-Type: multipart/x-mixed-replace; boundary=%s\r\n", boundary)
	req += "User-Agent: Happy Media TCP Client\r\n"
	req += "\r\n"

	fmt.Printf("SEND [%d]\n%s", len(req), req)

	_, err = conn.Write([]byte(req))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	_, err = pt.ReadMessage(conn)

	// send multipart stream, ex) jpg files
	err = pt.SendMultipartFiles(conn, "static/image/*.jpg")

	return nil, err
}

//---------------------------------------------------------------------------
// 41 handle request, please close socket after use
//---------------------------------------------------------------------------
func (pt *ProtoTcp) HandleRequest(conn net.Conn, sbuf *sr.StreamRing) error {
	var err error

	log.Printf("in %s\n", conn.RemoteAddr())
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
		_, err = pt.ReadPart(conn, "--agilemedia", slot)
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
func (pt *ProtoTcp) ReadPart(conn net.Conn, boundary string, ss *sr.StreamSlot) (map[string]string, error) {
	var err error

	reader := bufio.NewReader(conn)

	headers, err := pt.ReadPartHeader(reader, boundary)
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
		//_, err = pt.ReadBodyToData(reader, clen)
		ss.Type = headers["CONTENT-TYPE"]
		err = pt.ReadBodyToSlot(reader, clen, ss)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	return headers, err
}

//---------------------------------------------------------------------------
// read a message in http header style
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ReadMessage(conn net.Conn) (map[string]string, error) {
	var err error

	reader := bufio.NewReader(conn)

	headers, err := pt.ReadHeader(reader)
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
		_, err = pt.ReadBodyToData(reader, clen)
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
func (pt *ProtoTcp) ReadPartHeader(reader *bufio.Reader, boundary string) (map[string]string, error) {
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
			if strings.Contains(string(line), boundary) || strings.Contains(string(line), "POST") {
				fstart = true
			}
		}

		// if maybe invalid header data
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
func (pt *ProtoTcp) ReadHeader(reader *bufio.Reader) (map[string]string, error) {
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
// read body of message
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ReadBodyToData(reader *bufio.Reader, clen int) ([]byte, error) {
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
// read body of message
//---------------------------------------------------------------------------
func (pt *ProtoTcp) ReadBodyToSlot(reader *bufio.Reader, clen int, ss *sr.StreamSlot) error {
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
		return err
	}

	for i := range files {
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

		err = pt.SendPart(conn, fdata, ctype)
		if err != nil {
			log.Println(err)
			return err
		}

		time.Sleep(time.Second)
	}

	return err
}

//---------------------------------------------------------------------------
// send a part
//---------------------------------------------------------------------------
func (pt *ProtoTcp) SendPart(conn net.Conn, data []byte, ctype string) error {
	var err error

	clen := len(data)

	req := fmt.Sprintf("--%s\r\n", "myboundary")
	req += fmt.Sprintf("Content-Type: %s\r\n", ctype)
	req += fmt.Sprintf("Content-Length: %d\r\n", clen)
	req += "\r\n"

	defer fmt.Printf("SEND [%d,%d]\n%s", len(req), clen, req)

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