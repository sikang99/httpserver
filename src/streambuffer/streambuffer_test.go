//==================================================================================
// Test for stream buffer for multipart media
//==================================================================================
package streambuffer

import (
	"fmt"
	"log"
	"testing"
	"time"
)

//----------------------------------------------------------------------------------
// test for slot of stream buffer
//----------------------------------------------------------------------------------
func TestStreamSlot(t *testing.T) {
	data := make([]byte, 3)
	slot := NewStreamSlot("image/jpeg", 3, data)
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
	in := NewStreamSlot("image/jpeg", 3, data)

	out, err := sb.PutSlotNext(in)
	if err != nil {
		log.Fatalln("PutSlot")
	}

	in.Type = "text/plain"
	in.Content = []byte("Sample Message")
	in.Length = len(in.Content)

	out, _ = sb.PutSlotNext(in)

	in.Type = "text/html"
	in.Content = []byte("<html>Home Page</html>")
	in.Length = len(in.Content)

	out, _ = sb.PutSlotByPos(in, 3)
	fmt.Println(sb)

	out, _ = sb.GetSlotNext()
	fmt.Println(out)

	out, _ = sb.GetSlotByPos(2)
	out.Type = "video/mp4"
	out.Length = 10
	fmt.Printf("[%d] ", 2)
	fmt.Println(out)

	log.Println(sb.Resize(1))
	log.Println(sb.Resize(100000))
	sb.Resize(4)
	fmt.Println(sb)

	out, _ = sb.GetSlotByPos(3)
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
	in := NewStreamSlot("image/jpeg", len(data), data)
	sb.PutSlotNext(in)

	in.Type = "text/html"
	in.Content = []byte("<html>Home Page</html>")
	in.Length = len(in.Content)

	sb.PutSlotByPos(in, 3)
	fmt.Println(sb)

	println("Reading ...")
	for i := 0; i < 20; i++ {
		out, err := sb.GetSlotByPos(i)
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
		in := NewStreamSlot("text/plain", len(data), data)
		sb.PutSlotNext(in)
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

	var fend bool = false

	// define buffer reader
	reader := func(i int) {
		var pos int
		for fend == false {
			npos, out, err := sb.GetSlotByPosNext(pos)
			if out == nil && err == nil {
				time.Sleep(time.Millisecond)
				continue
			}
			fmt.Println("o>", i, out, pos, npos)
			pos = npos
		}
	}

	// define buffer writer
	writer := func(n int) {
		for i := 0; i < n; i++ {
			data := []byte(fmt.Sprintf("count %d", i))
			in := NewStreamSlot("text/plain", len(data), data)
			sb.PutSlotNext(in)
			fmt.Println("i>", in)
			time.Sleep(time.Millisecond)
		}
		fend = true
	}

	// multi reader, single writer
	nr := 3
	for i := 0; i < nr; i++ {
		go reader(i)
	}

	writer(20)
}

// ---------------------------------E-----N-----D-----------------------------------
