//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for TCP Socket Test
//==================================================================================

package protofile

import (
	"fmt"
	"log"
	"testing"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

//---------------------------------------------------------------------------
// test for info handling
//---------------------------------------------------------------------------
func TestHandleInfo(t *testing.T) {
	file := NewProtoFile("www.google.com", "9000", "Testing")
	fmt.Println(file)

	file.Reset()
	fmt.Println(file)

	file.Clear()
	fmt.Println(file)
}

//---------------------------------------------------------------------------
// test for single send and receive
//---------------------------------------------------------------------------
func TestReader(t *testing.T) {

}

// ---------------------------------E-----N-----D--------------------------------
