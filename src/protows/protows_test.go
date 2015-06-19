//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for WebSocket Test
//=================================================================================

package protows

import (
	"fmt"
	"log"
	"strconv"
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
	w1 := NewProtoWs()
	fmt.Println(w1)

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
	go rx.EchoServer()

	time.Sleep(time.Second)

	tx := NewProtoWs("localhost", "8087", "8443", "Tx")
	for i := 0; i < 5; i++ {
		tx.EchoClient(strconv.Itoa(i+1) + "> Hello World!")
	}
}

//---------------------------------------------------------------------------------
// test for caster and server
//---------------------------------------------------------------------------------
func TestCastServe(t *testing.T) {
	nx := NewProtoWs("localhost", "8087", "8443", "Nx")
	go nx.ActServer()

	time.Sleep(time.Second)

	tx := NewProtoWs("localhost", "8087", "8443", "Tx")
	tx.ActCaster("sec")
}

//---------------------------------------------------------------------------------
// test for server and player
//---------------------------------------------------------------------------------
func TestServePlay(t *testing.T) {
	nx := NewProtoWs("localhost", "8087", "8443", "Nx")
	go nx.ActServer()

	time.Sleep(time.Second)

	rx := NewProtoWs("localhost", "8087", "8443", "Rx")
	rx.ActPlayer()
}

//---------------------------------------------------------------------------------
// test for caster, server, and player
//---------------------------------------------------------------------------------
func TestCastServePlay(t *testing.T) {
	nx := NewProtoWs("localhost", "8087", "8443", "Nx")
	go nx.ActServer()

	time.Sleep(time.Second)

	tx := NewProtoWs("localhost", "8087", "8443", "Tx")
	go tx.ActCaster()

	time.Sleep(time.Millisecond)

	rx := NewProtoWs("localhost", "8087", "8443", "Rx")
	rx.ActPlayer()
}

// ----------------------------------E-----N-----D---------------------------------
