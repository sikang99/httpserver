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
}
