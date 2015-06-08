//=========================================================================
// Author : Stoney Kang, sikang99@gmail.com, 2015
// for HTTP streaming
//=========================================================================

package protohttp

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/bradfitz/http2"

	si "stoney/httpserver/src/streamimage"
	sr "stoney/httpserver/src/streamring"
)

//---------------------------------------------------------------------------
type ProtoHttp struct {
	Url      string
	Host     string
	Port     string
	PortTls  string
	Port2    string
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
func NewProtoHttp(hname, hport string) *ProtoHttp {

	return &ProtoHttp{
		Host: hname,
		Port: hport,
	}
}

//---------------------------------------------------------------------------
// config transport with timeout
//---------------------------------------------------------------------------
var timeout = time.Duration(3 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

func (ph *ProtoHttp) GetClientConfig() *http.Client {
	// simple timeout and tls setting
	tp := &http.Transport{
		Dial:            dialTimeout,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &http.Client{Transport: tp, Timeout: timeout}
}

//---------------------------------------------------------------------------
// http GET to server
//---------------------------------------------------------------------------
func (ph *ProtoHttp) SummitRequestGet(client *http.Client, url string) error {
	res, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	ph.PrintResponse(res, body)

	return err
}

//---------------------------------------------------------------------------
// http POST to server
//---------------------------------------------------------------------------
func (ph *ProtoHttp) SummitRequestPost(client *http.Client, url string) error {
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
	req, err := http.NewRequest("POST", url, outer)
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
// print response headers and body (text case) for debugging
//---------------------------------------------------------------------------
func (ph *ProtoHttp) PrintResponse(res *http.Response, body []byte) {
	h := res.Header
	for k, v := range h {
		fmt.Println("key:", k, "value:", v)
	}

	ct := res.Header.Get("Content-Type")

	println("")
	if strings.Contains(ct, "text") == true {
		fmt.Printf("%s\n", string(body))
	} else {
		fmt.Printf("[binary data]\n")
	}
	println("")
}

//---------------------------------------------------------------------------
//	receive a part to data
//---------------------------------------------------------------------------
func (ph *ProtoHttp) RecvPartToData(r *multipart.Reader) ([]byte, error) {
	var err error

	p, err := r.NextPart()
	if err != nil { // io.EOF
		log.Println(err)
		return nil, err
	}

	sl := p.Header.Get("Content-Length")
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

	fmt.Printf("%s %d/%d [%0x - %0x]\n", p.Header.Get("Content-Type"), tn, nl, data[:2], data[nl-2:])

	return data[:nl], err
}

//---------------------------------------------------------------------------
//	receive a part to slot of buffer
//----------------------------------------s-----------------------------------
func (ph *ProtoHttp) RecvPartToSlot(r *multipart.Reader, ss *sr.StreamSlot) error {
	var err error

	p, err := r.NextPart()
	if err != nil { // io.EOF
		log.Println(err)
		return err
	}

	sl := p.Header.Get("Content-Length")
	nl, err := strconv.Atoi(sl)
	if err != nil {
		log.Printf("%s %s %d\n", p.Header, sl, nl)
		return err
	}

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
	ss.Type = p.Header.Get("Content-Type")
	//fmt.Println(ss)

	return err
}

//---------------------------------------------------------------------------
//	receive multipart data into buffer
//---------------------------------------------------------------------------
func (ph *ProtoHttp) RecvMultipartToRing(r *multipart.Reader, sbuf *sr.StreamRing) error {
	var err error

	//sbuf := prepareStreamRing(5, MBYTE, "AXIS Camera")
	//sbuf := conf.Ring
	err = sbuf.SetStatusUsing()
	if err != nil {
		return sr.ErrStatus
	}

	// insert slots to the buffer
	for i := 0; true; i++ {
		//pre, pos := sbuf.ReadSlotIn()
		//fmt.Println("P", pos, pre)

		slot, pos := sbuf.GetSlotIn()
		//fmt.Println(i, pos, slot)

		err = ph.RecvPartToSlot(r, slot)
		if err != nil {
			log.Println(err)
			break
		}
		fmt.Println(i, pos, slot)

		sbuf.SetPosInByPos(pos + 1)
	}

	return err
}

//---------------------------------------------------------------------------
//	receive multipart data and decode jpeg
//---------------------------------------------------------------------------
func (ph *ProtoHttp) RecvMultipartToData(r *multipart.Reader) error {
	var err error

	for {
		p, err := r.NextPart()
		if err != nil { // io.EOF
			log.Println(err)
			return err
		}

		sl := p.Header.Get("Content-Length")
		nl, err := strconv.Atoi(sl)
		if err != nil {
			log.Printf("%s %s %d\n", p.Header, sl, nl)
			continue
		}

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

		fmt.Printf("%s %d/%d [%0x - %0x]\n", p.Header.Get("Content-Type"), tn, nl, data[:2], data[nl-2:])
	}

	return err
}

//---------------------------------------------------------------------------
// prepare a stream buffer
//---------------------------------------------------------------------------
func (ph *ProtoHttp) PrepareRing(nb int, size int, desc string) *sr.StreamRing {

	sbuf := sr.NewStreamRing(nb, size)
	sbuf.Desc = desc
	fmt.Println(sbuf)

	return sbuf
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
// static file server handler
//---------------------------------------------------------------------------
func (ph *ProtoHttp) FileServer(path string) http.Handler {
	log.Println("File server for " + path)
	return http.FileServer(http.Dir(path))
}

//---------------------------------------------------------------------------
// send a file
//---------------------------------------------------------------------------
func (ph *ProtoHttp) sendFile(w http.ResponseWriter, file string) error {
	w.Header().Set("Content-Type", "image/icon")
	w.Header().Set("Server", "Happy Media Server")
	body, err := ioutil.ReadFile("static/favicon.ico")
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Fprint(w, string(body))

	return nil
}

//---------------------------------------------------------------------------
// send a page
//---------------------------------------------------------------------------
func (ph *ProtoHttp) sendPage(w http.ResponseWriter, page string) error {
	t, err := template.New("mjpeg").Parse(page)
	if err != nil {
		log.Println(err)
		return err
	}

	return t.Execute(w, ph)
}

//---------------------------------------------------------------------------
// send image stream with the given format(extension) in the directory
//---------------------------------------------------------------------------
func (ph *ProtoHttp) sendMultipartImage(w io.Writer, dtype string, loop bool) error {
	var err error

	for {
		err = ph.SendPartImage(w, dtype)
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
func (ph *ProtoHttp) SendRing(w io.Writer, sbuf *sr.StreamRing) error {
	var err error

	//sbuf := conf.Ring
	if !sbuf.IsUsing() {
		return sr.ErrStatus
	}

	var pos int
	for {
		slot, npos, err := sbuf.GetSlotNextByPos(pos)
		if slot == nil && err == sr.ErrEmpty {
			time.Sleep(time.Millisecond)
			continue
		}

		err = ph.SendPartSlot(w, slot)
		if err != nil {
			log.Println(err)
			break
		}

		pos = npos
	}

	return err
}

//---------------------------------------------------------------------------
// send files with the given format(extension) in the directory
//---------------------------------------------------------------------------
func (ph *ProtoHttp) SendDir(w io.Writer, pat string, loop bool) error {
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
			err = ph.SendPartFile(w, files[i])
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
func (ph *ProtoHttp) SendPartFile(w io.Writer, file string) error {
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
	mw.SetBoundary("myboundary")

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
func (ph *ProtoHttp) SendPartImage(w io.Writer, dtype string) error {
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
	mw.SetBoundary(ph.Boundary)
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
func (ph *ProtoHttp) SendPartSlot(w io.Writer, ss *sr.StreamSlot) error {
	var err error

	mw := multipart.NewWriter(w)
	mw.SetBoundary("myboundary")
	part, err := mw.CreatePart(textproto.MIMEHeader{
		"Content-Type":   {ss.Type},
		"Content-Length": {strconv.Itoa(ss.Length)},
	})
	if err != nil {
		log.Println(err)
		return err
	}

	buf := new(bytes.Buffer)
	buf.Write(ss.Content[:ss.Length])
	buf.WriteTo(part)

	return err
}

//---------------------------------------------------------------------------
// send response for /stream with multipart format
//---------------------------------------------------------------------------
func (ph *ProtoHttp) SendRequest(w http.ResponseWriter) error {
	return nil
}

// for Caster
func (ph *ProtoHttp) SendResponseOk(w http.ResponseWriter) error {
	w.Header().Set("Server", "Happy Media Server")
	w.WriteHeader(http.StatusOK)

	return nil
}

// for Player
func (ph *ProtoHttp) SendResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=--"+ph.Boundary)
	w.Header().Set("Server", "Happy Media Server")
	w.WriteHeader(http.StatusOK)

	return nil
}

func (ph *ProtoHttp) GetBoundary(ctype string) error {
	var err error

	mt, params, err := mime.ParseMediaType(ctype)
	fmt.Printf("%v %v\n", params, ctype)
	if err != nil {
		log.Println("ParseMediaType: %s %v", mt, err)
		return err
	}

	boundary := params["boundary"]
	if !strings.HasPrefix(boundary, "--") {
		log.Printf("expected boundary to start with --, got %q", boundary)
	}

	ph.Boundary = boundary

	return err
}

//---------------------------------------------------------------------------
// respond message simply
//---------------------------------------------------------------------------
func (ph *ProtoHttp) Responder(w http.ResponseWriter, r *http.Request, status int, message string) {
	w.WriteHeader(status)
	log.Println(message)
	fmt.Fprintf(w, message)
}

// ---------------------------------E-----N-----D--------------------------------
