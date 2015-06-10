//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for File operation
//  - http://stackoverflow.com/questions/25090690/how-to-write-a-proxy-in-go-golang-using-tcp-connections
//  - https://github.com/nf/gohttptun - A tool to tunnel TCP over HTTP, written in Go
//==================================================================================

package protofile

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	sb "stoney/httpserver/src/streambase"
	sr "stoney/httpserver/src/streamring"
)

//---------------------------------------------------------------------------
const (
	STR_FILE_READER = "Happy Media File Reader"
	STR_FILE_WRITER = "Happy Media File Writer"
)

//---------------------------------------------------------------------------
type ProtoFile struct {
	Pattern  string
	Desc     string
	Boundary string
	//CreatedAt time.Time `bson:"created_at,omitempty json:"created_at,omitempty"`
	CreatedAt int64
}

//---------------------------------------------------------------------------
// string ProtoTcp information
//---------------------------------------------------------------------------
func (pf *ProtoFile) String() string {
	str := fmt.Sprintf("\tPattern: %v", pf.Pattern)
	str += fmt.Sprintf("\tBoundary: %v", pf.Boundary)
	str += fmt.Sprintf("\tDesc: %v", pf.Desc)
	str += fmt.Sprintf("\tCreatedAt: %v", pf.CreatedAt)
	return str
}

//---------------------------------------------------------------------------
// info handling
//---------------------------------------------------------------------------
func (pf *ProtoFile) Reset() {
	pf.Pattern = sb.STR_DEF_PATN
	pf.Boundary = sb.STR_DEF_BDRY
	pf.Desc = "reset"
}

func (pf *ProtoFile) Clear() {
	pf.Pattern = ""
	pf.Boundary = ""
	pf.Desc = "clear"
}

//---------------------------------------------------------------------------
// new ProtoTcp struct
//---------------------------------------------------------------------------
func NewProtoFile(pat, desc string) *ProtoFile {
	return &ProtoFile{
		Pattern:   pat,
		Desc:      desc,
		Boundary:  sb.STR_DEF_BDRY,
		CreatedAt: time.Now().Unix(),
	}
}

//---------------------------------------------------------------------------
// act file reader
//---------------------------------------------------------------------------
func (pf *ProtoFile) ActReader(sbuf *sr.StreamRing) {
	log.Println(STR_FILE_READER)

	ReadDirToRing(sbuf, pf.Pattern, false)
	fmt.Println(sbuf)

	return
}

//---------------------------------------------------------------------------
// act file writer
//---------------------------------------------------------------------------
func (pf *ProtoFile) ActWriter(sbuf *sr.StreamRing) {
	log.Println(STR_FILE_WRITER)

	WriteRingToMultipartFile(sbuf, pf.Pattern)

	return
}

//---------------------------------------------------------------------------
// read files with the given pattern in the directory and put them to the ring buffer
//---------------------------------------------------------------------------
func ReadDirToRing(sbuf *sr.StreamRing, pat string, loop bool) error {
	var err error

	// ReadDir : read directory
	files, err := filepath.Glob(pat)
	if err != nil {
		log.Println(err)
		return err
	}
	if files == nil {
		log.Printf("no matched file for %s\n", pat)
		return err
	}

	err = sbuf.SetStatusUsing()
	if err != nil {
		return sb.ErrStatus
	}
	defer sbuf.SetStatusIdle()

	// ToRing : to ring buffer
	for {
		for i := range files {
			slot, pos := sbuf.GetSlotIn()

			err = ReadFileToSlot(files[i], slot)
			if err == sb.ErrNull {
				continue
			}

			sbuf.SetPosInByPos(pos + 1)
			fmt.Println(slot)

			time.Sleep(time.Second)
		}

		if !loop {
			break
		}
	}

	return err
}

//---------------------------------------------------------------------------
// read multipart file to the ring buffer
//---------------------------------------------------------------------------
func ReadPartToSlot(mr *multipart.Reader, ss *sr.StreamSlot) error {
	var err error

	p, err := mr.NextPart()
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
	ss.Type = p.Header.Get("Content-Type")
	ss.Timestamp = sb.MakeTimestampMillisecond()
	//fmt.Println(ss)

	return err
}

//---------------------------------------------------------------------------
// read multipart file to the ring buffer
//---------------------------------------------------------------------------
func ReadMultipartFileToRing(sbuf *sr.StreamRing, file string) error {
	var err error

	f, err := os.Open(file)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()

	err = sbuf.SetStatusUsing()
	if err != nil {
		return sb.ErrStatus
	}
	defer sbuf.SetStatusIdle()

	mr := multipart.NewReader(f, sbuf.Boundary)

	for {
		slot, pos := sbuf.GetSlotIn()

		err = ReadPartToSlot(mr, slot)
		if err != nil {
			log.Println(err)
			break
		}
		fmt.Println(slot)

		sbuf.SetPosInByPos(pos + 1)
	}

	return err
}

//---------------------------------------------------------------------------
// write the ring buffer to file
//---------------------------------------------------------------------------
func WriteRingToMultipartFile(sbuf *sr.StreamRing, file string) error {
	var err error

	f, err := os.Create(file)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()

	// write ring buffer to file
	var pos int
	for sbuf.IsUsing() {
		slot, npos, err := sbuf.GetSlotNextByPos(pos)
		if err != nil {
			if err == sb.ErrEmpty {
				time.Sleep(sb.TIME_DEF_WAIT)
				continue
			}
			log.Println(err)
			break
		}

		// write slot
		//err = WriteSlotToFile(out, slot, sbuf.Boundary)

		w := bufio.NewWriter(f)
		err = WriteSlotToHandle(w, slot, sbuf.Boundary)

		pos = npos
	}

	return err
}

//---------------------------------------------------------------------------
// read a file to slot of ring buffer
//---------------------------------------------------------------------------
func ReadFileToSlot(file string, ss *sr.StreamSlot) error {
	var err error

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println(err)
		return err
	}
	dsize := len(data)

	if dsize == 0 {
		log.Printf("%s(%d) is null.\n", file, dsize)
		return sb.ErrNull
	}

	ctype := mime.TypeByExtension(filepath.Ext(file))
	if ctype == "" {
		ctype = http.DetectContentType(data)
	}

	copy(ss.Content, data)
	ss.Length = dsize
	ss.Type = ctype
	ss.Timestamp = sb.MakeTimestampMillisecond()

	return err
}

//---------------------------------------------------------------------------
// write the slot to file in a part
//---------------------------------------------------------------------------
func WriteSlotToFile(f *os.File, ss *sr.StreamSlot, boundary string) error {
	var err error

	str := fmt.Sprintf("--%s\r\n", boundary)
	str += fmt.Sprintf("Content-Type: %s\r\n", ss.Type)
	str += fmt.Sprintf("Content-Length: %d\r\n", ss.Length)
	str += "\r\n"

	f.WriteString(str)
	f.Write(ss.Content[:ss.Length])
	f.WriteString("\r\n")
	f.Sync()

	return err
}

//---------------------------------------------------------------------------
// write the slot to handle, bufio.Writer(io.Writer)
//---------------------------------------------------------------------------
func WriteSlotToHandle(w *bufio.Writer, ss *sr.StreamSlot, boundary string) error {
	var err error

	str := fmt.Sprintf("--%s\r\n", boundary)
	str += fmt.Sprintf("Content-Type: %s\r\n", ss.Type)
	str += fmt.Sprintf("Content-Length: %d\r\n", ss.Length)
	str += "\r\n"

	_, err = w.WriteString(str)
	_, err = w.Write(ss.Content[:ss.Length])
	_, err = w.WriteString("\r\n")
	w.Flush()

	return err
}

// ---------------------------------E-----N-----D--------------------------------
