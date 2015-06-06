//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for WebSocket Test
//=================================================================================

package protows

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func init() {
	//log.SetFlags(log.Lshortfile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

//---------------------------------------------------------------------------------
// test for
//---------------------------------------------------------------------------------
func TestProtoInfo(t *testing.T) {
	ws := NewProtoWs("www.google.com", "9000", "8443", "Testing")
	fmt.Println(ws)

	ws.Reset()
	fmt.Println(ws)

	ws.Clear()
	fmt.Println(ws)

	ws.SetAddr("www.facebook.com", "8080", "8443", "Redirect")
	fmt.Println(ws)
}

//---------------------------------------------------------------------------------
// test for echo
//---------------------------------------------------------------------------------
func TestEcho(t *testing.T) {
	rx := NewProtoWs("localhost", "8087", "8443", "Rx")
	go ActCatcher(rx)

	time.Sleep(time.Millisecond)

	tx := NewProtoWs("localhost", "8087", "8443", "Tx")
	for i := 0; i < 5; i++ {
		EchoClient(tx, "Hello World!")
	}
}

//---------------------------------------------------------------------------------
// test for single shoot
//---------------------------------------------------------------------------------
func TestSingleShootCatch(t *testing.T) {
	/*
		rx := NewProtoWs("localhost", "8087", "8443", "Rx")
		go ActCatcher(rx)

		time.Sleep(time.Millisecond)
	*/
	tx := NewProtoWs("localhost", "8087", "8443", "Tx")
	ActShooter(tx)
}

//---------------------------------------------------------------------------------
// test for multi shooters
//---------------------------------------------------------------------------------
func TestMultiShootCatch(t *testing.T) {
}

// ----------------------------------E-----N-----D---------------------------------
