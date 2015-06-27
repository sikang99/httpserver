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

//---------------------------------------------------------------------------
// for debugging
//---------------------------------------------------------------------------
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

//---------------------------------------------------------------------------
// test for info handling
//---------------------------------------------------------------------------
func TestHandleInfo(t *testing.T) {
	t1 := NewProtoTcp()
	fmt.Println(t1)

	t2 := NewProtoTcp("www.google.com")
	fmt.Println(t2)

	t3 := NewProtoTcp("www.google.com", "9000")
	fmt.Println(t3)

	t4 := NewProtoTcp("www.google.com", "9000", "Testing")
	fmt.Println(t4)

	t4.Reset()
	fmt.Println(t4)

	t4.Clear()
	fmt.Println(t4)

	t4.SetAddr("www.facebook.com", "8080", "Redirect")
	fmt.Println(t4)
}

//---------------------------------------------------------------------------
// test for send and receive
//---------------------------------------------------------------------------
func TestCastServe(t *testing.T) {
	sbuf := sr.NewStreamRingWithSize(5, sb.MBYTE)

	sx := NewProtoTcp("localhost", "8087", "Cx")
	go sx.StreamServer(sbuf)

	time.Sleep(time.Millisecond)

	cx := NewProtoTcp("localhost", "8087", "Sx")
	cx.StreamCaster()
}

//---------------------------------------------------------------------------
// test for send and receive
//---------------------------------------------------------------------------
func TestServePlay(t *testing.T) {
	sbuf := sr.NewStreamRingWithSize(3, 2*sb.MBYTE)
	sbuf.SetStatusUsing()

	// generate slot data
	go func() {
		num := time.Tick(1 * time.Second)

		for i := 0; ; {
			select {
			case <-num:
				slot, _ := sbuf.GetSlotIn()
				slot.Type = "test/data"
				slot.Length = i * 100 * sb.KBYTE
				slot.Timestamp = sb.GetTimestampNow()
				sbuf.SetPosInByPos(i)
				i++
			default:
				time.Sleep(time.Millisecond)
			}
		}
	}()

	sx := NewProtoTcp("localhost", "8087", "Cx")
	go sx.StreamServer(sbuf)

	time.Sleep(time.Millisecond)

	rbuf := sr.NewStreamRingWithSize(5, sb.MBYTE)

	px := NewProtoTcp("localhost", "8087", "Px")
	px.StreamPlayer(rbuf)
}

//---------------------------------------------------------------------------
// test for cast, serve, play
//---------------------------------------------------------------------------
func TestCastServePlay(t *testing.T) {
	sbuf := sr.NewStreamRingWithSize(4, sb.MBYTE)

	sx := NewProtoTcp("localhost", "8087", "Cx")
	go sx.StreamServer(sbuf)

	time.Sleep(time.Millisecond)

	cx := NewProtoTcp("localhost", "8087", "Sx")
	go cx.StreamCaster()

	time.Sleep(time.Second)

	rbuf := sr.NewStreamRingWithSize(3, sb.MBYTE)

	px := NewProtoTcp("localhost", "8087", "Px")
	px.StreamPlayer(rbuf)
}

// ---------------------------------E-----N-----D--------------------------------
