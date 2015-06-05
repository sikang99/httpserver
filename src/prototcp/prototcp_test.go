//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for TCP Socket
//  - http://stackoverflow.com/questions/25090690/how-to-write-a-proxy-in-go-golang-using-tcp-connections
//  - https://github.com/nf/gohttptun - A tool to tunnel TCP over HTTP, written in Go
//==================================================================================

package prototcp

import (
	"fmt"
	"testing"
	"time"
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
	//var wg sync.WaitGroup

	rx := NewProtoTcp("localhost", "8087", "Rx")
	go ActReceiver(rx)

	time.Sleep(time.Millisecond)

	tx := NewProtoTcp("localhost", "8087", "Tx")
	ActSender(tx)
}

//---------------------------------------------------------------------------
// test for multiple send and receive
//---------------------------------------------------------------------------
func TestMultiSendReceive(t *testing.T) {
}

// ---------------------------------E-----N-----D--------------------------------
