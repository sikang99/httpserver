//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for TCP Socket Test
//==================================================================================

package protofile

import (
	"fmt"
	"log"
	"testing"

	sr "stoney/httpserver/src/streamring"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

//---------------------------------------------------------------------------
// test for info handling
//---------------------------------------------------------------------------
func TestHandleInfo(t *testing.T) {
	file := NewProtoFile("../../static/image/*", "Testing")
	fmt.Println(file)

	file.Reset()
	fmt.Println(file)

	file.Clear()
	fmt.Println(file)
}

//---------------------------------------------------------------------------
// test for reader
//---------------------------------------------------------------------------
func TestReader(t *testing.T) {
	sbuf := sr.NewStreamRing(5, MBYTE)
	fmt.Println(sbuf)

	fr := NewProtoFile("../../static/image/*.jpg", "Testing")
	fr.ActReader(sbuf)
}

//---------------------------------------------------------------------------
// test for writer
//---------------------------------------------------------------------------
func TestWriter(t *testing.T) {
	sbuf := sr.NewStreamRing(5, MBYTE)
	fmt.Println(sbuf)

	fw := NewProtoFile("output.mjpg", "Testing")
	fw.ActWriter(sbuf)
}

// ---------------------------------E-----N-----D--------------------------------
