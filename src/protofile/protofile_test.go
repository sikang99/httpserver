//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for TCP Socket Test
//=================================================================================

package protofile

import (
	"fmt"
	"log"
	"testing"
	"time"

	sb "stoney/httpserver/src/streambase"
	sr "stoney/httpserver/src/streamring"
)

//---------------------------------------------------------------------------------
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

//---------------------------------------------------------------------------------
// test for info handling
//---------------------------------------------------------------------------------
func TestHandleInfo(t *testing.T) {
	f1 := NewProtoFile()
	fmt.Println(f1)

	f2 := NewProtoFile("../../static/image/*")
	fmt.Println(f2)

	file := NewProtoFile("../../static/image/*", "Testing")
	fmt.Println(file)

	file.Reset()
	fmt.Println(file)

	file.Clear()
	fmt.Println(file)
}

//---------------------------------------------------------------------------------
// test for read files and write to a file in multipart
//---------------------------------------------------------------------------------
func TestReadDirWriteMultipart(t *testing.T) {
	ring := sr.NewStreamRing(5, sb.MBYTE)
	fmt.Println(ring)

	fr := NewProtoFile("../../static/image/*.jpg", "Testing")
	go fr.DirReader(ring, false)

	time.Sleep(time.Millisecond)

	fw := NewProtoFile("output.mjpg", "Testing")
	fw.StreamWriter(ring)
}

//---------------------------------------------------------------------------------
// test for reading multipart file
//---------------------------------------------------------------------------------
func TestReadMultipartFile(t *testing.T) {
	ring := sr.NewStreamRing(5, sb.MBYTE)
	fmt.Println(ring)

	fd := NewProtoFile("output.mjpg", "Testing")
	fd.StreamReader(ring)
}

//----------------------------------E-----N-----D----------------------------------
