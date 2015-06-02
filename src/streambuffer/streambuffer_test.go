//==================================================================================
// Test for stream buffer for multipart media
//==================================================================================
package streambuffer

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
func TestStreamBuffer(t *testing.T) {
	var err error

	nb := 3
	sb := NewStreamBuffer(nb, KBYTE)
	fmt.Println(sb)

	sb.Resize(2)
	fmt.Println("Len: ", sb.Len(), "Cap: ", sb.Cap())

	sb.Resize(5)
	fmt.Println(sb)

	sb.Reset()
	fmt.Println(sb)

	data := make([]byte, 3)
	in := NewStreamSlotByData(KBYTE, "image/jpeg", 3, data)

	out, err := sb.PutSlotInNext(in)
	if err != nil {
		log.Fatalln("PutSlot")
	}

	in.Type = "text/plain"
	in.Content = []byte("Sample Message")
	in.Length = len(in.Content)

	out, _ = sb.PutSlotInNext(in)

	in.Type = "text/html"
	in.Content = []byte("<html>Home Page</html>")
	in.Length = len(in.Content)

	out, _ = sb.PutSlotInByPos(in, 3)
	fmt.Println(sb)

	out, _ = sb.GetSlotOutNext()
	fmt.Println(out)

	out, _ = sb.GetSlotOutByPos(2)
	out.Type = "video/mp4"
	out.Length = 10
	fmt.Printf("[%d] ", 2)
	fmt.Println(out)

	log.Println(sb.Resize(1))
	log.Println(sb.Resize(100000))
	sb.Resize(4)
	fmt.Println(sb)

	out, _ = sb.GetSlotOutByPos(3)
	fmt.Println(out)
	fmt.Println(sb)
}

//----------------------------------------------------------------------------------
// test for continuous read from stream buffer
//----------------------------------------------------------------------------------
func TestStreamRead(t *testing.T) {
	// prepare a buffer
	nb := 5
	sb := NewStreamBuffer(nb, MBYTE)

	data := make([]byte, 128)
	in := NewStreamSlotByData(KBYTE, "image/jpeg", len(data), data)
	sb.PutSlotInNext(in)

	in.Type = "text/html"
	in.Content = []byte("<html>Home Page</html>")
	in.Length = len(in.Content)

	sb.PutSlotInByPos(in, 3)
	fmt.Println(sb)

	println("Reading ...")
	for i := 0; i < 20; i++ {
		out, err := sb.GetSlotOutByPos(i)
		if err != nil {
			break
		}
		//sb.Resize(i + 1)
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
	nb := 5
	sb := NewStreamBuffer(nb, MBYTE)

	for i := 1; i < 50; i++ {
		data := []byte(fmt.Sprintf("count %d", i))
		in := NewStreamSlotByData(KBYTE, "text/plain", len(data), data)
		sb.PutSlotInNext(in)
	}
	fmt.Println(sb)
}

//----------------------------------------------------------------------------------
// test for multiple readers and single writer on the stream buffer
//----------------------------------------------------------------------------------
func TestStreamReadWrite(t *testing.T) {
	// prepare a buffer
	nb := 5
	sb := NewStreamBuffer(nb, MBYTE)
	sb.Desc = "Test buffer"

	var fend bool = false

	// define buffer reader
	reader := func(i int) {
		var pos int
		for fend == false {
			out, npos, err := sb.GetSlotOutNextByPos(pos)
			if out == nil && err == ErrEmpty {
				time.Sleep(time.Millisecond)
				continue
			}
			//fmt.Println("o>", i, out, pos, npos)
			pos = npos
		}
	}

	// define buffer writer
	writer := func(n int) {
		for i := 0; i < n; i++ {
			tn := 100 * i * i * i
			data := []byte(fmt.Sprintf("saved %d-th data", tn))
			/*
				in := NewStreamSlotByData(KBYTE, "text/plain", len(data), data)
				ss, _ := sb.PutSlotInNext(in)
			*/
			ss, pos := sb.GetSlotIn()

			ss.Type = "text/plain"
			ss.Length = len(data)
			copy(ss.Content, data)

			fmt.Println("i>", i, ss)
			fmt.Println(sb)
			sb.SetPosInByPos(pos + 1)

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

// ---------------------------------E-----N-----D-----------------------------------
