//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for File operation
//  - http://stackoverflow.com/questions/25090690/how-to-write-a-proxy-in-go-golang-using-tcp-connections
//  - https://github.com/nf/gohttptun - A tool to tunnel TCP over HTTP, written in Go
//==================================================================================

package protofile

import (
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
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
	STR_DEF_PATN = "*.jpg"

	STR_PGM_READER = "Happy Media File Reader"
	STR_PGM_WRITER = "Happy Media File Writer"
)

//---------------------------------------------------------------------------
type ProtoFile struct {
	Pattern  string
	Desc     string
	Boundary string
}

//---------------------------------------------------------------------------
// string ProtoTcp information
//---------------------------------------------------------------------------
func (pf *ProtoFile) String() string {
	str := fmt.Sprintf("\tPattern: %s", pf.Pattern)
	str += fmt.Sprintf("\tBoundary: %s", pf.Boundary)
	str += fmt.Sprintf("\tDesc: %s", pf.Desc)
	return str
}

//---------------------------------------------------------------------------
// info handling
//---------------------------------------------------------------------------
func (pf *ProtoFile) Reset() {
	pf.Pattern = STR_DEF_PATN
	pf.Boundary = STR_DEF_BDRY
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
		Pattern:  pat,
		Desc:     desc,
		Boundary: STR_DEF_BDRY,
	}
}

//---------------------------------------------------------------------------
// act file reader
//---------------------------------------------------------------------------
func (pf *ProtoFile) ActReader(sbuf *sr.StreamRing, pat string) {
	log.Println(STR_PGM_READER)

	pf.ReadDirToRing(sbuf, pat, true)
	fmt.Println(sbuf)

	return
}

//---------------------------------------------------------------------------
// act file writer
//---------------------------------------------------------------------------
func (pf *ProtoFile) Actwriter(sbuf *sr.StreamRing) {
	log.Println(STR_PGM_WRITER)

	pf.WriteRingToFile(sbuf, "./static/output.mjpg")
}

//---------------------------------------------------------------------------
// write the ring buffer to file
//---------------------------------------------------------------------------
func (pf *ProtoFile) WriteRingToFile(sbuf *sr.StreamRing, file string) error {
	var err error

	out, err := os.Open(file)
	if err != nil {
		log.Println(err)
		return err
	}
	defer out.Close()

	return err
}

//---------------------------------------------------------------------------
// read files with the given pattern in the directory and put them to the ring buffer
//---------------------------------------------------------------------------
func (pf *ProtoFile) ReadDirToRing(sbuf *sr.StreamRing, pat string, loop bool) error {
	var err error

	// direct pattern matching
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
		return sr.ErrStatus
	}

	for {
		for i := range files {
			slot, pos := sbuf.GetSlotIn()
			err = pf.ReadFileToSlot(files[i], slot)
			if err != nil {
				if err == sr.ErrEmpty {
					continue
				}
				log.Println(err)
				return err
			}
			fmt.Println(slot)
			sbuf.SetPosInByPos(pos + 1)
			time.Sleep(time.Second)
		}

		if !loop {
			break
		}
	}

	return err
}

//---------------------------------------------------------------------------
// read a file to slot of ring buffer
//---------------------------------------------------------------------------
func (pf *ProtoFile) ReadFileToSlot(file string, ss *sr.StreamSlot) error {
	var err error

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println(err)
		return err
	}
	dsize := len(data)

	if dsize == 0 {
		log.Printf("%s(%d) is empty\n", file, dsize)
		return sr.ErrEmpty
	}

	ctype := mime.TypeByExtension(filepath.Ext(file))
	if ctype == "" {
		ctype = http.DetectContentType(data)
	}

	copy(ss.Content, data)
	ss.Length = dsize
	ss.Type = ctype

	return err
}

// ---------------------------------E-----N-----D--------------------------------
