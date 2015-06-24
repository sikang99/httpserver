//=========================================================================
// Author : Stoney Kang, sikang99@gmail.com, 2015
// Protocol for HTTP streaming
// - http://www.sanarias.com/blog/1214PlayingwithimagesinHTTPresponseingolang
// - http://stackoverflow.com/questions/30552447/how-to-set-which-ip-to-use-for-a-http-request
// - http://giantmachines.tumblr.com/post/52184842286/golang-http-client-with-timeouts
//=========================================================================

package protohttp

import (
	"bytes"
	"encoding/json"
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
	"text/template"
	"time"

	"github.com/fatih/color"

	sb "stoney/httpserver/src/streambase"
	si "stoney/httpserver/src/streamimage"
	sr "stoney/httpserver/src/streamring"
)

//---------------------------------------------------------------------------
const (
	STR_HTTP_CASTER  = "Happy Media HTTP Caster"
	STR_HTTP_SERVER  = "Happy Media HTTP Server"
	STR_HTTP_PLAYER  = "Happy Media HTTP Player"
	STR_HTTP_MONITOR = "Happy Media HTTP Monitor"
	STR_HTTP_READER  = "Happy Media HTTP Reader"
	STR_HTTP_WRITER  = "Happy Media HTTP Writer"
)

type ProtoHttp struct {
	Url      string
	Host     string
	Port     string
	PortTls  string
	Port2    string
	Method   string
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
func NewProtoHttp(args ...string) *ProtoHttp {
	ph := &ProtoHttp{
		Host:     sb.STR_DEF_HOST,
		Port:     sb.STR_DEF_PORT,
		PortTls:  sb.STR_DEF_PTLS,
		Port2:    sb.STR_DEF_PORT2,
		Boundary: sb.STR_DEF_BDRY,
	}

	ph.Desc = "NewProtoHttp"

	for i, arg := range args {
		switch {
		case i == 0:
			ph.Boundary = arg
		case i == 1:
			ph.Desc = arg
		}
	}

	return ph
}

func NewProtoHttpWithPorts(args ...string) *ProtoHttp {
	ph := NewProtoHttp()

	ph.Desc = "NewProtoHttpWithPorts"

	for i, arg := range args {
		switch {
		case i == 0:
			ph.Port = arg
		case i == 1:
			ph.PortTls = arg
		case i == 2:
			ph.Port2 = arg
		}
	}
	return ph
}

//---------------------------------------------------------------------------
// http GET to server
//---------------------------------------------------------------------------
func (ph *ProtoHttp) RequestGet(client *http.Client) error {
	var err error

	res, err := client.Get(ph.Url)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res, body)

	return err
}

//---------------------------------------------------------------------------
// http POST to server
//---------------------------------------------------------------------------
func (ph *ProtoHttp) RequestPost(client *http.Client) error {
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
	req, err := http.NewRequest("POST", ph.Url, outer)
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
//	receive a part to data
//---------------------------------------------------------------------------
func ReadPartToData(mr *multipart.Reader) ([]byte, error) {
	var err error

	p, err := mr.NextPart()
	if err != nil { // io.EOF
		log.Println(err)
		return nil, err
	}

	sl := p.Header.Get(sb.STR_HDR_CONTENT_LENGTH)
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

	//fmt.Printf("%s %d/%d [%0x - %0x]\n", p.Header.Get("Content-Type"), tn, nl, data[:2], data[nl-2:])
	return data[:nl], err
}

//---------------------------------------------------------------------------
//	receive a part to slot of buffer
//---------------------------------------------------------------------------
func ReadPartToSlot(mr *multipart.Reader, slot *sr.StreamSlot) error {
	var err error

	p, err := mr.NextPart()
	if err != nil { // io.EOF
		log.Println(err)
		return err
	}

	sl := p.Header.Get(sb.STR_HDR_CONTENT_LENGTH)
	nl, err := strconv.Atoi(sl)
	if err != nil {
		log.Printf("%s %s %d\n", p.Header, sl, nl)
		return err
	}

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
	slot.Type = p.Header.Get(sb.STR_HDR_CONTENT_TYPE)
	slot.Timestamp = sb.GetTimestampNow()
	//fmt.Println(slot)

	return err
}

//---------------------------------------------------------------------------
//	receive multipart data into buffer
//---------------------------------------------------------------------------
func ReadMultipartToRing(mr *multipart.Reader, ring *sr.StreamRing) error {
	var err error

	err = ring.SetStatusUsing()
	if err != nil {
		log.Println(sb.RedString("ErrStatus/ReadMultipartToRing"))
		return sb.ErrStatus
	}
	defer ring.SetStatusIdle()
	//fmt.Println(ring)

	// insert slots to the buffer
	for i := 0; ring.IsUsing(); i++ {
		//pre, pos := ring.ReadSlotIn()
		//fmt.Println("P", pos, pre)

		slot, pos := ring.GetSlotIn()
		//fmt.Println(i, pos, slot)

		err = ReadPartToSlot(mr, slot)
		if err != nil {
			log.Println(err)
			break
		}
		//fmt.Println(i, pos, slot)

		ring.SetPosInByPos(pos + 1)
	}

	return err
}

//---------------------------------------------------------------------------
//	receive multipart data and decode jpeg
//---------------------------------------------------------------------------
func ReadMultipartToData(mr *multipart.Reader) error {
	var err error

	for {
		p, err := mr.NextPart()
		if err != nil { // io.EOF
			log.Println(err)
			return err
		}

		sl := p.Header.Get(sb.STR_HDR_CONTENT_LENGTH)
		nl, err := strconv.Atoi(sl)
		if sl == "" || nl == 0 {
			log.Printf("%s %s %d\n", p.Header, sl, nl)
			continue
		}
		if err != nil {
			log.Println(err)
			return err
		}

		if nl > 0 {
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

			//log.Printf("%s %d/%d [%0x - %0x]\n", p.Header.Get(sb.STR_HDR_CONTENT_LENGTH), tn, nl, data[:2], data[nl-2:])
		}
	}

	return err
}

//---------------------------------------------------------------------------
// send a favicon
//---------------------------------------------------------------------------
func WriteFavicon(w http.ResponseWriter, file string) error {
	w.Header().Set("Content-Type", "image/icon")
	w.Header().Set("Server", STR_HTTP_SERVER)
	body, err := ioutil.ReadFile("static/favicon.ico")
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Fprint(w, string(body))

	return nil
}

//---------------------------------------------------------------------------
// send a template page
//---------------------------------------------------------------------------
func WriteTemplatePage(w http.ResponseWriter, page string, data interface{}) error {
	t, err := template.New("mjpeg").Parse(page)
	if err != nil {
		log.Println(err)
		return err
	}

	return t.Execute(w, data)
}

//---------------------------------------------------------------------------
// send image stream with the given format(extension) in the directory
//---------------------------------------------------------------------------
func WriteImagesInMultipart(w io.Writer, dtype string, loop bool) error {
	var err error

	for {
		err = WriteImageInPart(w, dtype, "myboundary")
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
func WriteRingInMultipart(w io.Writer, ring *sr.StreamRing) error {
	var err error

	if !ring.IsUsing() {
		fmt.Println(ring)
		log.Println(sb.RedString("ErrStatus/WriteRingInMultipart"))
		return sb.ErrStatus
	}
	//fmt.Println(ring)

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

		err = WriteSlotInPart(w, slot, ring.Boundary)
		if err != nil {
			log.Println(err)
			break
		}
		//fmt.Println(slot)

		pos = npos
	}

	return err
}

//---------------------------------------------------------------------------
// send files with the given format(extension) in the directory
//---------------------------------------------------------------------------
func WriteDirInMultipart(w io.Writer, pat string, loop bool) error {
	var err error

	// direct pattern matching
	files, err := filepath.Glob(pat)
	if err != nil {
		log.Println(err)
		return err
	}
	if files == nil {
		log.Println(color.RedString("ErrFound/WriteDirInMultipart"))
		return sb.ErrFound
	}

	for {
		for i := range files {
			err = WriteFileInPart(w, files[i], "myboundary")
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
func WriteFileInPart(w io.Writer, file string, boundary string) error {
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
	mw.SetBoundary(boundary)

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
func WriteImageInPart(w io.Writer, dtype string, boundary string) error {
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
	mw.SetBoundary(boundary)
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
func WriteSlotInPart(w io.Writer, slot *sr.StreamSlot, boundary string) error {
	var err error

	mw := multipart.NewWriter(w)
	mw.SetBoundary(boundary)
	part, err := mw.CreatePart(textproto.MIMEHeader{
		"Content-Type":   {slot.Type},
		"Content-Length": {strconv.Itoa(slot.Length)},
	})
	if err != nil {
		log.Println(err)
		return err
	}

	buf := new(bytes.Buffer)
	_, err = buf.Write(slot.Content[:slot.Length])
	_, err = buf.WriteTo(part)

	return err
}

//---------------------------------------------------------------------------
// get boundary string
//---------------------------------------------------------------------------
func GetTypeBoundary(ctype string) (string, error) {
	var err error

	mt, params, err := mime.ParseMediaType(ctype)
	//fmt.Printf("%v %v\n", params, ctype)
	if err != nil {
		log.Println("ParseMediaType: %s %v", mt, err)
		return "", err
	}

	boundary := params["boundary"]
	if !strings.HasPrefix(boundary, "--") {
		log.Printf("expected to start with --, got %q", boundary)
	}

	return boundary, err
}

//---------------------------------------------------------------------------
// send response for GET
//---------------------------------------------------------------------------
func ResponseGet(w http.ResponseWriter, boundary string) error {
	var err error

	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=--"+boundary)
	w.Header().Set("Server", STR_HTTP_SERVER)
	w.WriteHeader(http.StatusOK)

	return err
}

//---------------------------------------------------------------------------
// send response for POST
//---------------------------------------------------------------------------
func ResponsePost(w http.ResponseWriter, boundary string) error {
	var err error

	w.Header().Set("Server", STR_HTTP_SERVER)
	w.WriteHeader(http.StatusOK)

	return err
}

//---------------------------------------------------------------------------
// send response message simply
//---------------------------------------------------------------------------
func WriteResponseMessage(w http.ResponseWriter, status int, message string) error {
	var err error

	w.WriteHeader(status)
	log.Println(message)
	fmt.Fprintf(w, message)

	return err
}

//---------------------------------------------------------------------------
// send post message with json format
//---------------------------------------------------------------------------
func RequestPostJson(url string) error {
	var err error

	var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	return err
}

//---------------------------------------------------------------------------
// get json data from request
// - http://stackoverflow.com/questions/15672556/handling-json-post-request-in-go
//---------------------------------------------------------------------------
func GetJsonFromRequest(r *http.Request, struc interface{}) error {
	var err error

	//dec := json.NewDecoder(strings.NewReader(jdata))
	//dec := json.NewDecoder(r.Body)
	//err = dec.Decode(&struc)
	err = json.NewDecoder(r.Body).Decode(&struc)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

// ---------------------------------E-----N-----D--------------------------------
