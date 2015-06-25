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
	STR_DIR_READER  = "Happy Media Dir Reader"
	STR_FILE_READER = "Happy Media File Reader"
	STR_FILE_WRITER = "Happy Media File Writer"
)

//---------------------------------------------------------------------------
type ProtoFile struct {
	Pattern    string
	Boundary   string
	Desc       string
	CreatedAt  int64
	ModifiedAt int64
}

//---------------------------------------------------------------------------
// string ProtoTcp information
//---------------------------------------------------------------------------
func (pf *ProtoFile) String() string {
	str := fmt.Sprintf("\tPattern: %v", pf.Pattern)
	str += fmt.Sprintf("\tBoundary: %v", pf.Boundary)
	str += fmt.Sprintf("\tDesc: %v", pf.Desc)
	str += fmt.Sprintf("\tCreatedAt: %v,%v", pf.CreatedAt, pf.ModifiedAt)
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
	pf.Desc = "cleared"
}

//---------------------------------------------------------------------------
// new ProtoTcp struct
//---------------------------------------------------------------------------
func NewProtoFile(args ...string) *ProtoFile {
	pf := &ProtoFile{
		Boundary:  sb.STR_DEF_BDRY,
		CreatedAt: time.Now().Unix(),
	}

	for i, arg := range args {
		if i == 0 {
			pf.Pattern = arg
		} else if i == 1 {
			pf.Desc = arg
		}
	}

	return pf
}

//---------------------------------------------------------------------------
// act file reader
//---------------------------------------------------------------------------
func (pf *ProtoFile) DirReader(ring *sr.StreamRing, loop bool) error {
	log.Printf("%s for %s\n", STR_DIR_READER, pf.Pattern)
	defer log.Printf("out %s\n", STR_DIR_READER)

	var err error

	err = ReadDirToRing(ring, pf.Pattern, loop)
	//fmt.Println(ring)

	return err
}

//---------------------------------------------------------------------------
// act file reader
//---------------------------------------------------------------------------
func (pf *ProtoFile) StreamReader(ring *sr.StreamRing) error {
	log.Printf("%s for %s\n", STR_FILE_READER, pf.Pattern)
	defer log.Printf("out %s\n", STR_FILE_READER)

	var err error

	err = ReadMultipartFileToRing(ring, pf.Pattern)
	//fmt.Println(ring)

	return err
}

//---------------------------------------------------------------------------
// act file writer
//---------------------------------------------------------------------------
func (pf *ProtoFile) StreamWriter(ring *sr.StreamRing) error {
	log.Printf("%s for %s\n", STR_FILE_WRITER, pf.Pattern)
	defer log.Printf("out %s\n", STR_FILE_WRITER)

	var err error

	err = WriteRingToMultipartFile(ring, pf.Pattern)
	//fmt.Println(ring)

	return err
}

//---------------------------------------------------------------------------
// read files with the given pattern in the directory and put them to the ring buffer
//---------------------------------------------------------------------------
func ReadDirToRing(ring *sr.StreamRing, pat string, loop bool) error {
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

	err = ring.SetStatusUsing()
	if err != nil {
		return sb.ErrStatus
	}
	defer ring.SetStatusIdle()

	// ToRing : to ring buffer
	for ring.IsUsing() {
		for i := range files {
			slot, pos := ring.GetSlotIn()

			err = ReadFileToSlot(files[i], slot)
			if err == sb.ErrNull {
				continue
			}

			ring.SetPosInByPos(pos + 1)
			fmt.Println("FR", slot)

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

	sl := p.Header.Get(sb.STR_HDR_CONTENT_LENGTH)
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
	ss.Type = p.Header.Get(sb.STR_HDR_CONTENT_TYPE)
	ss.Timestamp = sb.GetTimestampNow()
	//fmt.Println(ss)

	ts := p.Header.Get(sb.STR_HDR_TIMESTAMP)
	ss.Timestamp, err = strconv.ParseInt(ts, 10, 64)

	return err
}

//---------------------------------------------------------------------------
// read multipart file to the ring buffer
// TODO: change relative time gap into absolute one to prevent drift
//---------------------------------------------------------------------------
func ReadMultipartFileToRing(ring *sr.StreamRing, file string) error {
	var err error

	f, err := os.Open(file)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()

	err = ring.SetStatusUsing()
	if err != nil {
		log.Println(err)
		return err
	}
	defer ring.SetStatusIdle()

	// TODO: set boundary from the multipart file

	mr := multipart.NewReader(f, ring.Boundary)

	var preTimestamp int64 = 0
	for ring.IsUsing() {
		slot, pos := ring.GetSlotIn()

		err = ReadPartToSlot(mr, slot)
		if err != nil {
			log.Println(err)
			break
		}

		// read multipart to ring buffer by timestamp
		if preTimestamp > 0 {
			diff := slot.Timestamp - preTimestamp
			time.Sleep(sb.GetDuration(diff - 1))
		}
		preTimestamp = slot.Timestamp

		fmt.Println("MR", slot)
		ring.SetPosInByPos(pos + 1)
	}

	return err
}

//---------------------------------------------------------------------------
// write the ring buffer to file
//---------------------------------------------------------------------------
func WriteRingToMultipartFile(ring *sr.StreamRing, file string) error {
	var err error

	f, err := os.Create(file)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()

	// write ring buffer to file
	var pos int
	for ring.IsUsing() {
		slot, npos, err := ring.GetSlotNextByPos(pos)
		if err != nil {
			if err == sb.ErrEmpty {
				time.Sleep(sb.TIME_DEF_WAIT)
				continue
			}
			log.Println(err)
			break
		}

		// write slot
		//err = WriteSlotToFile(out, slot, ring.Boundary)

		w := bufio.NewWriter(f)
		err = WriteSlotToHandle(w, slot, ring.Boundary)

		fmt.Println("MW", slot)
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

	if dsize > ss.LengthMax {
		log.Printf("%d is too big than %d\n", dsize, ss.LengthMax)
		return sb.ErrSize
	}

	ctype := mime.TypeByExtension(filepath.Ext(file))
	if ctype == "" {
		ctype = http.DetectContentType(data)
	}

	copy(ss.Content, data)
	ss.Length = dsize
	ss.Type = ctype
	ss.Timestamp = sb.GetTimestampNow()

	return err
}

//---------------------------------------------------------------------------
// write the slot to file in a part
//---------------------------------------------------------------------------
func WriteSlotToFile(f *os.File, ss *sr.StreamSlot, boundary string) error {
	var err error

	str := fmt.Sprintf("--%s\r\n", boundary)
	str += fmt.Sprintf("%s: %s\r\n", sb.STR_HDR_CONTENT_TYPE, ss.Type)
	str += fmt.Sprintf("%s: %d\r\n", sb.STR_HDR_CONTENT_LENGTH, ss.Length)
	str += fmt.Sprintf("%s: %v; scale=%s\r\n", sb.STR_HDR_TIMESTAMP, ss.Timestamp, sb.STR_TIME_PRECISION)
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
	str += fmt.Sprintf("%s: %s\r\n", sb.STR_HDR_CONTENT_TYPE, ss.Type)
	str += fmt.Sprintf("%s: %d\r\n", sb.STR_HDR_CONTENT_LENGTH, ss.Length)
	str += fmt.Sprintf("%s: %v\r\n", sb.STR_HDR_TIMESTAMP, ss.Timestamp)
	str += "\r\n"

	_, err = w.WriteString(str)
	_, err = w.Write(ss.Content[:ss.Length])
	_, err = w.WriteString("\r\n")
	w.Flush()

	return err
}

// ---------------------------------E-----N-----D--------------------------------
