package ring

import (
	"fmt"
	"testing"
)

func TestStreamSlot(t *testing.T) {
	data := make([]byte, 3)
	slot := NewStreamSlot("image/jpeg", 3, data)
	fmt.Println(slot)

}

func TestStreamBuffer(t *testing.T) {
	sb := NewStreamBuffer(3, MBYTE)
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

	out = sb.PutSlotByPos(4, in)
	fmt.Println(sb)

	out = sb.GetSlot()
	fmt.Println(out)

	out = sb.GetSlot()
	fmt.Println(out)

	out = sb.GetSlotByPos(15)
	fmt.Println(out)
	fmt.Println(sb)
}
