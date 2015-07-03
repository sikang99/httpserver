//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for WebSocket Test
// - http://talks.golang.kr/2015/go-test.slide
//=================================================================================

package protowsm

import (
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"
)

//---------------------------------------------------------------------------------
func init() {
	//log.SetFlags(log.Lshortfile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

//---------------------------------------------------------------------------------
// test for input
//---------------------------------------------------------------------------------
func testMain(m *testing.M) {
	m.Run()
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
// test for echo server
//---------------------------------------------------------------------------------
func TestEchoReceive(t *testing.T) {
	rx := NewProtoWs("localhost", "8087", "8443", "Rx")
	rx.EchoServer()
}

//---------------------------------------------------------------------------------
// test for echo send and receive
//---------------------------------------------------------------------------------
func TestEchoSendReceive(t *testing.T) {
	rx := NewProtoWs("localhost", "8087", "8443", "Rx")
	go rx.EchoServer()

	time.Sleep(time.Second)

	tx := NewProtoWs("localhost", "8087", "8443", "Tx")
	for i := 0; i < 5; i++ {
		tx.EchoClient(strconv.Itoa(i+1) + "> Hello World!")
	}
	tx.EchoClient("quit")
}

//---------------------------------------------------------------------------------
// test for stream server
//---------------------------------------------------------------------------------
func TestStreamServe(t *testing.T) {
	rx := NewProtoWs("localhost", "8087", "8443", "Nx")
	rx.StreamServer()
}

//---------------------------------------------------------------------------------
// test for caster and server
//---------------------------------------------------------------------------------
func TestStreamCastServe(t *testing.T) {
	nx := NewProtoWs("localhost", "8087", "8443", "Nx")
	go nx.StreamServer()

	time.Sleep(time.Second)

	tx := NewProtoWs("localhost", "8087", "8443", "Tx")
	tx.StreamCaster("sec")
}

//---------------------------------------------------------------------------------
// test for server and player
//---------------------------------------------------------------------------------
func TestStreamServePlay(t *testing.T) {
	nx := NewProtoWs("localhost", "8087", "8443", "Nx")
	go nx.StreamServer()

	time.Sleep(time.Second)

	rx := NewProtoWs("localhost", "8087", "8443", "Rx")
	rx.StreamPlayer()
}

//---------------------------------------------------------------------------------
// test for caster, server, and player
//---------------------------------------------------------------------------------
func TestStreamCastServePlay(t *testing.T) {
	nx := NewProtoWs("localhost", "8087", "8443", "Nx")
	go nx.StreamServer()

	time.Sleep(time.Second)

	tx := NewProtoWs("localhost", "8087", "8443", "Tx")
	go tx.StreamCaster()

	time.Sleep(time.Millisecond)

	rx := NewProtoWs("localhost", "8087", "8443", "Rx")
	rx.StreamPlayer()
}

// ----------------------------------E-----N-----D---------------------------------
