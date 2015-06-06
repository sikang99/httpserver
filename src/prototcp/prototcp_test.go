//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for TCP Socket Test
//==================================================================================

package prototcp

import (
	"fmt"
	"testing"
	"time"

	sr "stoney/httpserver/src/streamring"
)

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
	rx := NewProtoTcp("localhost", "8087", "Rx")
	sbuf := sr.NewStreamRing(2, MBYTE)
	go ActReceiver(rx, sbuf)

	time.Sleep(time.Millisecond)

	tx := NewProtoTcp("localhost", "8087", "Tx")
	ActSender(tx)
}

//---------------------------------------------------------------------------
// test for multiple send and receive
//---------------------------------------------------------------------------
func TestMultiSendReceive(t *testing.T) {
	//var wg sync.WaitGroup
}

// ---------------------------------E-----N-----D--------------------------------
