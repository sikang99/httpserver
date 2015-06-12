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
// test for send and receive
//---------------------------------------------------------------------------
func TestSendReceive(t *testing.T) {
	sbuf := sr.NewStreamRing(5, sb.MBYTE)

	sx := NewProtoTcp("localhost", "8087", "Cx")
	go sx.ActServer(sbuf)

	time.Sleep(time.Millisecond)

	cx := NewProtoTcp("localhost", "8087", "Sx")
	cx.ActCaster()
}

//---------------------------------------------------------------------------
// test for cast, serve, play
//---------------------------------------------------------------------------
func TestCastServePlay(t *testing.T) {
	sbuf := sr.NewStreamRing(5, sb.MBYTE)

	sx := NewProtoTcp("localhost", "8087", "Cx")
	go sx.ActServer(sbuf)

	time.Sleep(time.Millisecond)

	cx := NewProtoTcp("localhost", "8087", "Sx")
	go cx.ActCaster()

	time.Sleep(time.Second)

	rbuf := sr.NewStreamRing(5, sb.MBYTE)

	px := NewProtoTcp("localhost", "8087", "Px")
	px.ActPlayer(rbuf)
}

// ---------------------------------E-----N-----D--------------------------------
