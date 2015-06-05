//=================================================================================
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
// test
//---------------------------------------------------------------------------
func TestSender(t *testing.T) {
	tcp := NewProtoTcp("www.google.com", "9000", "Testing")
	fmt.Println(tcp)

	tcp.Reset()
	fmt.Println(tcp)

	tcp.Clear()
	fmt.Println(tcp)
}

func TestSendReceive(t *testing.T) {
	//var wg sync.WaitGroup

	rx := NewProtoTcp("localhost", "8087", "Rx")
	go ActReceiver(rx)

	time.Sleep(time.Second)

	tx := NewProtoTcp("localhost", "8087", "Tx")
	ActSender(tx)
}

// ---------------------------------E-----N-----D--------------------------------
