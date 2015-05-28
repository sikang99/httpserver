package ring

import (
	"fmt"
	"testing"
	"time"
)

func TestStreamSlot(t *testing.T) {
	data := make([]byte, 3)
	slot := NewStreamSlot("image/jpeg", 3, data)
	fmt.Println(slot)

}

func TestStreamBuffer(t *testing.T) {
	sb := NewStreamBuffer(3, KBYTE)
	fmt.Println(sb)

	sb.Resize(5)
	fmt.Println(sb)

	sb.Reset()
	fmt.Println(sb)

	data := make([]byte, 3)
	in := NewStreamSlot("image/jpeg", 3, data)

	out := sb.PutSlot(in)

	in.Type = "text/plain"
	in.Content = []byte("Sample Message")
	in.Length = len(in.Content)

	out = sb.PutSlot(in)

	in.Type = "text/html"
	in.Content = []byte("<html>Home Page</html>")
	in.Length = len(in.Content)

	out = sb.PutSlotByPos(3, in)
	fmt.Println(sb)

	out = sb.GetSlot()
	fmt.Println(out)

	out = sb.GetSlot()
	fmt.Println(out)

	fmt.Println(sb.Resize(1))
	fmt.Println(sb.Resize(100000))
	sb.Resize(4)
	fmt.Println(sb)

	out = sb.GetSlotByPos(3)
	fmt.Println(out)
	fmt.Println(sb)
}

func TestBufferStreaming(t *testing.T) {
	n := 5
	sb := NewStreamBuffer(n, MBYTE)

	data := make([]byte, 3)
	in := NewStreamSlot("image/jpeg", 3, data)

	in.Type = "text/html"
	in.Content = []byte("<html>Home Page</html>")
	in.Length = len(in.Content)

	sb.PutSlotByPos(3, in)
	fmt.Println(sb)

	for i := 0; i < 50; i++ {
		out := sb.GetSlot()
		fmt.Println(i, ": ", out)
		if i%n == 0 {
			time.Sleep(time.Second)
		}
	}
}
