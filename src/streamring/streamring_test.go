//==================================================================================
// Author : Stoney Kang, sikang99@gmail.com, 2015
// Test for stream ring(slots) buffer for multipart media
//==================================================================================

package streamring

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	sb "stoney/httpserver/src/streambase"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

//----------------------------------------------------------------------------------
// - http://stackoverflow.com/questions/30525184/array-vs-slice-accessing-speed
// - https://github.com/ChristianSiegert/go-testing-example
//----------------------------------------------------------------------------------
var gs = make([]byte, 1000) // Global slice
var ga [1000]byte           // Global array

func BenchmarkSliceGlobal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j, v := range gs {
			gs[j]++
			gs[j] = gs[j] + v + 10
			gs[j] += v
		}
	}
}

func BenchmarkArrayGlobal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j, v := range ga {
			ga[j]++
			ga[j] = ga[j] + v + 10
			ga[j] += v
		}
	}
}

func BenchmarkSliceLocal(b *testing.B) {
	var s = make([]byte, 1000)
	for i := 0; i < b.N; i++ {
		for j, v := range s {
			s[j]++
			s[j] = s[j] + v + 10
			s[j] += v
		}
	}
}

func BenchmarkArrayLocal(b *testing.B) {
	var a [1000]byte
	for i := 0; i < b.N; i++ {
		for j, v := range a {
			a[j]++
			a[j] = a[j] + v + 10
			a[j] += v
		}
	}
}

//----------------------------------------------------------------------------------
// test for single slot of stream buffer
//----------------------------------------------------------------------------------
func TestStreamSlot(t *testing.T) {
	slot := NewStreamSlotBySize(1024)
	fmt.Println(slot)

	dsize := 12
	data := make([]byte, dsize)
	data[0] = 0xff
	data[dsize-1] = 0xee

	slot = NewStreamSlotByData(dsize, "image/jpeg", len(data), data)
	fmt.Println(slot)

	assert.Equal(t, false, slot.IsType("text/plain"))
	assert.Equal(t, true, slot.IsType("image/JPEG"))
	assert.Equal(t, false, slot.IsMajorType("video"))
	assert.Equal(t, true, slot.IsSubType("jpeg"))

	slot = NewStreamSlotByData(dsize, "image", len(data), data)
	fmt.Println(slot)

	assert.Equal(t, true, slot.IsMajorType("IMAGE"))
	assert.Equal(t, false, slot.IsSubType("jpeg"))

	slot = NewStreamSlotBySlot(dsize, slot)
	fmt.Println(slot)
}

//----------------------------------------------------------------------------------
// test for handling functions of the stream buffer
//----------------------------------------------------------------------------------
func TestStreamRing(t *testing.T) {
	var err error

	nb := 3
	sr := NewStreamRingWithSize(nb, sb.KBYTE)
	fmt.Println(sr)

	sr.Resize(2)
	fmt.Println("Len: ", sr.Len(), "Cap: ", sr.Cap())

	sr.Resize(5)
	fmt.Println(sr)

	sr.Reset()
	fmt.Println(sr)

	data := make([]byte, 3)
	in := NewStreamSlotByData(sb.KBYTE, "image/jpeg", 3, data)

	out, err := sr.PutSlotInNext(in)
	if err != nil {
		log.Fatalln("PutSlot")
	}

	in.Type = "text/plain"
	in.Content = []byte("Sample Message")
	in.Length = len(in.Content)

	out, _ = sr.PutSlotInNext(in)

	in.Type = "text/html"
	in.Content = []byte("<html>Home Page</html>")
	in.Length = len(in.Content)

	out, _ = sr.PutSlotInByPos(in, 3)
	fmt.Println(sr)

	out, _ = sr.GetSlotOutNext()
	fmt.Println(out)

	out, _ = sr.GetSlotByPos(2)
	out.Type = "video/mp4"
	out.Length = 10
	fmt.Printf("[%d] ", 2)
	fmt.Println(out)

	log.Println(sr.Resize(1))
	log.Println(sr.Resize(100000))
	sr.Resize(4)
	fmt.Println(sr)

	out, _ = sr.GetSlotByPos(3)
	fmt.Println(out)
	fmt.Println(sr)
}

//----------------------------------------------------------------------------------
// test for continuous read from stream buffer
//----------------------------------------------------------------------------------
func TestStreamRead(t *testing.T) {
	// prepare a buffer
	nb := 5
	sr := NewStreamRingWithSize(nb, sb.MBYTE)

	data := make([]byte, 128)
	in := NewStreamSlotByData(sb.KBYTE, "image/jpeg", len(data), data)
	sr.PutSlotInNext(in)

	in.Type = "text/html"
	in.Content = []byte("<html>Home Page</html>")
	in.Length = len(in.Content)

	sr.PutSlotInByPos(in, 3)
	fmt.Println(sr)

	println("Reading ...")
	for i := 0; i < 20; i++ {
		out, err := sr.GetSlotByPos(i)
		if err != nil {
			break
		}
		//sr.Resize(i + 1)
		fmt.Printf("\t(%d) %s\n", i, out)
		if (i+1)%nb == 0 {
			println()
			//time.Sleep(time.Second)
		}
	}
}

//----------------------------------------------------------------------------------
// test for continuous write from stream buffer
//----------------------------------------------------------------------------------
func TestStreamWrite(t *testing.T) {
	var err error

	ns := 5
	sr := NewStreamRingWithSize(ns, sb.MBYTE)
	fmt.Println(sr)

	err = sr.SetStatusUsing()
	if err != nil {
		log.Fatalln(err)
	}
	sr.Desc = "Testing ..."

	for i := 1; i < 50; i++ {
		data := []byte(fmt.Sprintf("count %d", i))
		in := NewStreamSlotByData(sb.KBYTE, "text/plain", len(data), data)
		sr.PutSlotInNext(in)
	}
	fmt.Println(sr)

	if !sr.IsUsing() {
		log.Fatalln(err)
	}

	sr.Reset()
	fmt.Println(sr)

	if sr.GetStatus() != sb.STATUS_IDLE {
		log.Fatalln(err)
	}
}

//----------------------------------------------------------------------------------
// test for multiple readers and single writer on the stream buffer
//----------------------------------------------------------------------------------
func TestStreamReadWrite(t *testing.T) {
	// prepare a buffer
	nb := 5
	sr := NewStreamRingWithSize(nb, sb.MBYTE)
	sr.Desc = "Test buffer"

	var fend bool = false

	// define buffer reader()
	reader := func(i int) {
		var pos int
		for fend == false {
			out, npos, err := sr.GetSlotNextByPos(pos)
			if out == nil && err == sb.ErrEmpty {
				time.Sleep(time.Millisecond)
				continue
			}
			fmt.Println("R>", i, out, pos, npos)
			pos = npos
		}
	}

	// define buffer writer()
	writer := func(n int) {
		for i := 0; i < n; i++ {
			tn := 100 * i * i * i
			data := []byte(fmt.Sprintf("saved %d-th data", tn))
			in := NewStreamSlotByData(sb.KBYTE, "text/plain", len(data), data)
			ss, _ := sr.PutSlotInNext(in)
			/*
				ss, pos := sr.GetSlotIn()

				ss.Type = "text/plain"
				ss.Length = len(data)
				copy(ss.Content, data)

				fmt.Println(sr)
				sr.SetPosInByPos(pos + 1)
			*/
			fmt.Println("W>", i, ss)
			time.Sleep(time.Millisecond)
		}
		fend = true
	}

	// multi reader, single writer
	nr := 3
	for i := 0; i < nr; i++ {
		go reader(i)
	}

	writer(10)
}

//----------------------------------------------------------------------------------
// test for stream array (set of rings)
//----------------------------------------------------------------------------------
func TestStreamArray(t *testing.T) {
	array := NewStreamArrayWithSize(4, 3, sb.GBYTE)
	for i := range array {
		fmt.Println(array[i])
	}
}

// ---------------------------------E-----N-----D-----------------------------------
