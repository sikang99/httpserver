//==================================================================================
// Test for ring buffer for streaming
//==================================================================================
package ring

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

	sb := NewStreamBuffer(3, KBYTE)
	fmt.Println(sb)

	sb.Resize(2)
	fmt.Println("Len: ", sb.Len(), "Cap: ", sb.Cap())

	sb.Resize(5)
	fmt.Println(sb)

	sb.Reset()
	fmt.Println(sb)

	data := make([]byte, 3)
	in := NewStreamSlot("image/jpeg", 3, data)

	out, err := sb.PutSlot(in)
	if err != nil {
		log.Fatalln("PutSlot")
	}

	in.Type = "text/plain"
	in.Content = []byte("Sample Message")
	in.Length = len(in.Content)

	out, _ = sb.PutSlot(in)

	in.Type = "text/html"
	in.Content = []byte("<html>Home Page</html>")
	in.Length = len(in.Content)

	out, _ = sb.PutSlotByPos(in, 3)
	fmt.Println(sb)

	out, _ = sb.GetSlot()
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
// continuous read from stream buffer
//----------------------------------------------------------------------------------
func TestStreamRead(t *testing.T) {
	n := 5
	sb := NewStreamBuffer(n, MBYTE)

	data := make([]byte, 128)
	in := NewStreamSlot("image/jpeg", len(data), data)
	sb.PutSlot(in)

	in.Type = "text/html"
	in.Content = []byte("<html>Home Page</html>")
	in.Length = len(in.Content)

	sb.PutSlotByPos(in, 3)
	fmt.Println(sb)

	for i := 0; i < 20; i++ {
		out, err := sb.GetSlot()
		if err != nil {
			break
		}
		//sb.Resize(i + 1)
		fmt.Printf("\t(%d) %s\n", i, out)
		if (i+1)%n == 0 {
			println()
			//time.Sleep(time.Second)
		}
	}
}

//----------------------------------------------------------------------------------
// continuous write from stream buffer
//----------------------------------------------------------------------------------
func TestStreamWrite(t *testing.T) {
	n := 5
	sb := NewStreamBuffer(n, MBYTE)

	for i := 1; i < 50; i++ {
		data := []byte(fmt.Sprintf("count %d", i))
		in := NewStreamSlot("text/plain", len(data), data)
		sb.PutSlot(in)
	}
	fmt.Println(sb)
}

//----------------------------------------------------------------------------------
// continuous read/write to/from stream buffer
//----------------------------------------------------------------------------------
func TestStreamReadWrite(t *testing.T) {
	n := 2
	sb := NewStreamBuffer(n, MBYTE)

	go func() {
		for i := 0; i < 20; i++ {
			data := []byte(fmt.Sprintf("count %d", i))
			in := NewStreamSlot("text/plain", len(data), data)
			sb.PutSlot(in)
			fmt.Println("i>", in)
			time.Sleep(time.Second)
		}
	}()

	for {
		out, err := sb.GetSlot()
		if out == nil && err == nil {
			time.Sleep(time.Millisecond)
			continue
		}
		fmt.Println("o>", out)
	}
}

// ---------------------------------E-----N-----D-----------------------------------
