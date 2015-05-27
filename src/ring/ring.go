//==================================================================================
// Circular Stream Buffer
// - https://github.com/zfjagann/golang-ring
// - https://github.com/glycerine/rbuf
// - http://blog.pivotal.io/labs/labs/a-concurrent-ring-buffer-for-go
//==================================================================================

package ring

import (
	"fmt"
	"sync"
)

//----------------------------------------------------------------------------------
const (
	KBYTE = 1024
	MBYTE = 1024 * KBYTE // Kilo
	GBYTE = 1024 * MBYTE // Giga
	TBYTE = 1024 * GBYTE // Tera
	HBYTE = 1024 * TBYTE // Hexa
)

type StreamSlot struct {
	sync.Mutex
	Type    string
	Length  int
	Max     int
	Content []byte
}

func NewStreamSlot(ctype string, clen int, cdata []byte) *StreamSlot {
	return &StreamSlot{
		Type:    ctype,
		Length:  clen,
		Content: cdata,
	}
}

func (ss *StreamSlot) String() string {
	str := fmt.Sprintf("Type: %s\t", ss.Type)
	str += fmt.Sprintf("Length: %d\t", ss.Length)
	str += fmt.Sprintf("Content: %02x\t", ss.Content[0])
	return str
}

//----------------------------------------------------------------------------------
type StreamBuffer struct {
	sync.Mutex
	Slots []StreamSlot
	Num   int // number of slots used
	Max   int // number of slots allocated
	Size  int // number of slots allocated
	In    int // input position of buffer
	Out   int // output position of buffer
	Desc  string
}

func NewStreamBuffer(num int, size int) *StreamBuffer {
	slot := StreamSlot{
		Type:    "application/octet-stream",
		Length:  size,
		Content: make([]byte, size),
	}

	var slots []StreamSlot
	for i := 0; i < num; i++ {
		slots = append(slots, slot)
	}

	return &StreamBuffer{
		Slots: slots,
		Num:   num, Max: num,
		Size: size,
		In:   0, Out: 0,
		Desc: "Null stream",
	}
}

func (sb *StreamBuffer) Len() int {
	return sb.Num
}

func (sb *StreamBuffer) Cap() int {
	return sb.Max
}

func (sb *StreamBuffer) String() string {
	str := fmt.Sprintf("StreamBuffer: ")
	str += fmt.Sprintf("Pos: %d,%d\t", sb.In, sb.Out)
	str += fmt.Sprintf("Size: %d/%d,%d\t", sb.Num, sb.Max, sb.Size)
	str += fmt.Sprintf("Desc: %s\t", sb.Desc)

	for i := 0; i < sb.Num; i++ {
		str += fmt.Sprintf("\n")
		str += fmt.Sprintf("\t%s", &sb.Slots[i])
	}

	return str
}

func (sb *StreamBuffer) GetSlot() *StreamSlot {
	return &sb.Slots[sb.Out]
}

func (sb *StreamBuffer) PutSlot() *StreamSlot {
	return &sb.Slots[sb.In]
}

func (sb *StreamBuffer) Clear() {
	sb.Lock()
	defer sb.Unlock()

	for i := 0; i < sb.Num; i++ {
		sb.Slots[i].Type = ""
		sb.Slots[i].Length = 0
	}
}

func (sb *StreamBuffer) Reset() {
	sb.Lock()
	defer sb.Unlock()

	for i := 0; i < sb.Max; i++ {
		sb.Slots[i].Type = ""
		sb.Slots[i].Length = 0
	}
	sb.Num = sb.Max
}

func (sb *StreamBuffer) Resize(num int) {
	sb.Lock()
	defer sb.Unlock()

	if num > sb.Max {
		slot := StreamSlot{
			Content: make([]byte, sb.Size),
		}
		for i := 0; i < num-sb.Max; i++ {
			sb.Slots = append(sb.Slots, slot)
		}
		sb.Max = num
	}
	sb.Num = num
}

// ---------------------------------E-----N-----D-----------------------------------
