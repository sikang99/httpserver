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

	sb "stoney/httpserver/src/streambase"
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
func TestCastServePlay(t *testing.T) {
	sbuf := sr.NewStreamRing(2, sb.MBYTE)

	sx := NewProtoTcp("localhost", "8087", "Rx")
	go sx.ActServer(sbuf)

	time.Sleep(time.Millisecond)

	cx := NewProtoTcp("localhost", "8087", "Tx")
	go cx.ActCaster()

	time.Sleep(time.Millisecond)

	rbuf := sr.NewStreamRing(2, sb.MBYTE)

	px := NewProtoTcp("localhost", "8087", "Tx")
	px.ActPlayer(rbuf)
}

//---------------------------------------------------------------------------
// test for multiple send and receive
//---------------------------------------------------------------------------
func TestSendReceive(t *testing.T) {
	sbuf := sr.NewStreamRing(2, sb.MBYTE)

	sx := NewProtoTcp("localhost", "8087", "Rx")
	go sx.ActServer(sbuf)

	time.Sleep(time.Millisecond)

	cx := NewProtoTcp("localhost", "8087", "Tx")
	cx.ActCaster()
}

// ---------------------------------E-----N-----D--------------------------------
