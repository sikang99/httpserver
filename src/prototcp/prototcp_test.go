//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for TCP Socket Test
//==================================================================================

package prototcp

import (
	"fmt"
	"log"
	"testing"
	"time"

	sr "stoney/httpserver/src/streamring"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

//---------------------------------------------------------------------------
// test for info handling
//---------------------------------------------------------------------------
func TestHandleInfo(t *testing.T) {
	tcp := NewProtoTcp("www.google.com", "9000", "Testing")
	fmt.Println(tcp)

	tcp.Reset()
	fmt.Println(tcp)

	tcp.Clear()
	fmt.Println(tcp)

	tcp.SetAddr("www.facebook.com", "8080", "Redirect")
	fmt.Println(tcp)
}

//---------------------------------------------------------------------------
// test for single send and receive
//---------------------------------------------------------------------------
func TestSingleSendReceive(t *testing.T) {
	sbuf := sr.NewStreamRing(2, MBYTE)

	rx := NewProtoTcp("localhost", "8087", "Rx")
	go rx.ActReceiver(sbuf)

	time.Sleep(time.Millisecond)

	tx := NewProtoTcp("localhost", "8087", "Tx")
	tx.ActSender()
}

//---------------------------------------------------------------------------
// test for multiple send and receive
//---------------------------------------------------------------------------
func TestMultiSendReceive(t *testing.T) {
	//var wg sync.WaitGroup
}

// ---------------------------------E-----N-----D--------------------------------
