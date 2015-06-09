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

	sr "stoney/httpserver/src/streamring"
)

//---------------------------------------------------------------------------------
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

//---------------------------------------------------------------------------------
// test for timestamp
//---------------------------------------------------------------------------------
func TestTimestamp(t *testing.T) {
	//tstamp, _ := time.Parse(time.RFC3339, strconv.Itoa(file.CreatedAt))
	//println(tstamp)

	CreatedAt := time.Now().Unix()
	tstring := time.Unix(CreatedAt, 0).Format(time.RFC3339)
	fmt.Printf("\tCurrent Timestamp: %v, %v\n", CreatedAt, tstring)

	sec := MakeTimestampSecond()
	msec := MakeTimestampMillisecond()
	nsec := MakeTimestampNanosecond()
	fmt.Printf("%v\n%v\n%v\n", sec, msec, nsec)
}

//---------------------------------------------------------------------------------
// test for info handling
//---------------------------------------------------------------------------------
func TestHandleInfo(t *testing.T) {
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
func TestFileReadWrite(t *testing.T) {
	sbuf := sr.NewStreamRing(5, MBYTE)
	fmt.Println(sbuf)

	fr := NewProtoFile("../../static/image/*.jpg", "Testing")
	go fr.ActReader(sbuf)

	time.Sleep(time.Second)

	fw := NewProtoFile("output.mjpg", "Testing")
	fw.ActWriter(sbuf)
}

//---------------------------------------------------------------------------------
// test for reading multipart file
//---------------------------------------------------------------------------------
func TestReadMultipartFile(t *testing.T) {
	sbuf := sr.NewStreamRing(5, MBYTE)
	fmt.Println(sbuf)

	ReadMultipartFileToRing(sbuf, "output.mjpg")
	fmt.Println(sbuf)
}

//----------------------------------E-----N-----D----------------------------------
