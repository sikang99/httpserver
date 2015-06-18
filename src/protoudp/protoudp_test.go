//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for UDP Socket Test
//==================================================================================

package protoudp

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
	t1 := NewProtoUdp()
	fmt.Println(t1)

	t2 := NewProtoUdp("www.google.com")
	fmt.Println(t2)

	t3 := NewProtoUdp("www.google.com", "9000")
	fmt.Println(t3)

	t4 := NewProtoUdp("www.google.com", "9000", "Testing")
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
	sbuf := sr.NewStreamRing(5, sb.MBYTE)

	sx := NewProtoUdp("localhost", "8087", "Cx")
	go sx.ActServer(sbuf)

	time.Sleep(time.Millisecond)

	cx := NewProtoUdp("localhost", "8087", "Sx")
	cx.ActCaster()
}

//---------------------------------------------------------------------------
// test for send and receive
//---------------------------------------------------------------------------
func TestServePlay(t *testing.T) {
	sbuf := sr.NewStreamRing(3, 2*sb.MBYTE)
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
				slot.Timestamp = sb.GetTimestamp()
				sbuf.SetPosInByPos(i)
				i++
			default:
				time.Sleep(time.Millisecond)
			}
		}
	}()

	sx := NewProtoUdp("localhost", "8087", "Cx")
	go sx.ActServer(sbuf)

	time.Sleep(time.Millisecond)

	rbuf := sr.NewStreamRing(5, sb.MBYTE)

	px := NewProtoUdp("localhost", "8087", "Px")
	px.ActPlayer(rbuf)
}

//---------------------------------------------------------------------------
// test for cast, serve, play
//---------------------------------------------------------------------------
func TestCastServePlay(t *testing.T) {
	sbuf := sr.NewStreamRing(4, sb.MBYTE)

	sx := NewProtoUdp("localhost", "8087", "Cx")
	go sx.ActServer(sbuf)

	time.Sleep(time.Millisecond)

	cx := NewProtoUdp("localhost", "8087", "Sx")
	go cx.ActCaster()

	time.Sleep(time.Second)

	rbuf := sr.NewStreamRing(3, sb.MBYTE)

	px := NewProtoUdp("localhost", "8087", "Px")
	px.ActPlayer(rbuf)
}

// ---------------------------------E-----N-----D--------------------------------
