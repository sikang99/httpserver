//==================================================================================
// Circular Stream Buffer
// - https://github.com/zfjagann/golang-ring
// - https://github.com/glycerine/rbuf
// - http://blog.pivotal.io/labs/labs/a-concurrent-ring-buffer-for-go
//==================================================================================

package streambuffer

import (
	"fmt"
	"strings"
	"sync"
)

//----------------------------------------------------------------------------------
const (
	KBYTE = 1024
	MBYTE = 1024 * KBYTE // Kilo
	GBYTE = 1024 * MBYTE // Giga
	TBYTE = 1024 * GBYTE // Tera
	HBYTE = 1024 * TBYTE // Hexa

	LEN_MAX_SLOT  = GBYTE
	NUM_MAX_SLOTS = 512
)

type StreamSlot struct {
	sync.Mutex
	Type      string
	Length    int
	LengthMax int
	Content   []byte
}

//----------------------------------------------------------------------------------
// string information for the slot
//----------------------------------------------------------------------------------
func (ss *StreamSlot) String() string {
	str := fmt.Sprintf("\tType: %s", ss.Type)
	str += fmt.Sprintf("\tLength: %d/%d", ss.Length, ss.LengthMax)
	str += fmt.Sprintf("\tContent: ")
	if ss.Length > 0 {
		if strings.Contains(ss.Type, "text/") {
			str += fmt.Sprintf("%s", string(ss.Content[:ss.Length]))
		} else {
			str += fmt.Sprintf("%02x-%02x", ss.Content[0], ss.Content[ss.Length-1])
		}
	}
	return str
}

//----------------------------------------------------------------------------------
// make a new slot of given size
//----------------------------------------------------------------------------------
func NewStreamSlotBySize(clen int) *StreamSlot {
	return &StreamSlot{
		Length:    clen,
		LengthMax: clen,
		Content:   make([]byte, clen),
	}
}

//----------------------------------------------------------------------------------
// make a new slot assigned with given data
//----------------------------------------------------------------------------------
func NewStreamSlotByData(ctype string, clen int, cdata []byte) *StreamSlot {
	return &StreamSlot{
		Type:      ctype,
		Length:    clen,
		LengthMax: clen,
		Content:   cdata,
	}
}

//----------------------------------------------------------------------------------
type StreamBuffer struct {
	sync.Mutex
	Status int
	Num    int    // number of slots used
	NumMax int    // number of slots allocated
	Size   int    // size of the slot content
	In     int    // input position of buffer to be written
	Out    int    // output position of buffer to be read
	Desc   string // description of buffer
	Slots  []StreamSlot
}

//----------------------------------------------------------------------------------
// string information for the stream buffer
//----------------------------------------------------------------------------------
func (sb *StreamBuffer) String() string {
	str := fmt.Sprintf("[StreamBuffer]")
	str += fmt.Sprintf("\tPos: %d,%d", sb.In, sb.Out)
	str += fmt.Sprintf("\tSize: %d/%d, %d KB", sb.Num, sb.NumMax, sb.Size/KBYTE)
	str += fmt.Sprintf("\tDesc: %s\n", sb.Desc)

	for i := 0; i < sb.Num; i++ {
		str += fmt.Sprintf("\t[%d] %s\n", i, &sb.Slots[i])
	}

	return str
}

//----------------------------------------------------------------------------------
// make a new stream ring buffer
//----------------------------------------------------------------------------------
func NewStreamBuffer(num int, size int) *StreamBuffer {
	slot := StreamSlot{
		//Type:    "application/octet-stream",
		Length:    size,
		LengthMax: size,
		Content:   make([]byte, size),
	}

	var slots []StreamSlot
	for i := 0; i < num; i++ {
		slots = append(slots, slot)
	}

	return &StreamBuffer{
		Slots: slots,
		Num:   num, NumMax: num,
		Size: size,
		In:   0, Out: 0,
		Desc: "Null data",
	}
}

//----------------------------------------------------------------------------------
// get the number of slots in buffer for used(len) and allocted(cap)
//----------------------------------------------------------------------------------
func (sb *StreamBuffer) Len() int {
	return sb.Num
}

func (sb *StreamBuffer) Cap() int {
	return sb.NumMax
}

//----------------------------------------------------------------------------------
// set the position of slot to be read and written
//----------------------------------------------------------------------------------
func (sb *StreamBuffer) SetPosIn(pos int) int {
	sb.In = (pos % sb.Num)
	return sb.In
}

func (sb *StreamBuffer) SetPosOut(pos int) int {
	sb.Out = (pos % sb.Num)
	return sb.Out
}

//----------------------------------------------------------------------------------
// get the pointer of slot to be read and move to the next
//----------------------------------------------------------------------------------
func (sb *StreamBuffer) GetSlotNext() (*StreamSlot, error) {
	var err error

	// no data to read
	if sb.In == sb.Out {
		return nil, nil
	}

	slot := &sb.Slots[sb.Out]
	sb.Out = (sb.Out + 1) % sb.Num

	return slot, err
}

//----------------------------------------------------------------------------------
// get the pointer of slot designated
//----------------------------------------------------------------------------------
func (sb *StreamBuffer) GetSlotByPos(pos int) (*StreamSlot, error) {
	var err error

	pos = pos % sb.Num
	slot := &sb.Slots[pos]

	return slot, err
}

//----------------------------------------------------------------------------------
// get the pointer of slot designated and go to the next
//----------------------------------------------------------------------------------
func (sb *StreamBuffer) GetSlotByPosNext(pos int) (*StreamSlot, int, error) {
	var err error

	pos = pos % sb.Num

	// no data to read
	if sb.In == pos {
		return nil, pos, nil
	}

	slot := &sb.Slots[pos]
	pos = (pos + 1) % sb.Num

	return slot, pos, err
}

//----------------------------------------------------------------------------------
// write the information to the slot and go to the next
//----------------------------------------------------------------------------------
func (sb *StreamBuffer) PutSlotNext(slot *StreamSlot) (*StreamSlot, error) {
	sb.Lock()
	defer sb.Unlock()

	var err error

	if slot.Length > sb.Size {
		return nil, fmt.Errorf("too big data size")
	}

	st := &sb.Slots[sb.In]

	//*st = *slot	// CAUTION: Don't copy slot.LengthMax
	st.Type = slot.Type
	st.Length = slot.Length
	copy(st.Content, slot.Content)

	sb.In = (sb.In + 1) % sb.Num

	return st, err
}

//----------------------------------------------------------------------------------
// write the information to the slot designated
//----------------------------------------------------------------------------------
func (sb *StreamBuffer) PutSlotByPos(slot *StreamSlot, pos int) (*StreamSlot, error) {
	sb.Lock()
	defer sb.Unlock()

	var err error

	pos = (pos % sb.Num)
	st := &sb.Slots[pos]
	*st = *slot

	return slot, err
}

//----------------------------------------------------------------------------------
// clear the stream buffer
//----------------------------------------------------------------------------------
func (sb *StreamBuffer) Reset() {
	sb.Lock()
	defer sb.Unlock()

	for i := 0; i < sb.NumMax; i++ {
		sb.Slots[i].Type = ""
		sb.Slots[i].Length = 0
	}

	sb.Num = sb.NumMax
}

//----------------------------------------------------------------------------------
// change the size of buffer, i.e, the number of slots
//----------------------------------------------------------------------------------
func (sb *StreamBuffer) Resize(num int) error {
	sb.Lock()
	defer sb.Unlock()

	var err error

	if num < 2 || num > NUM_MAX_SLOTS {
		return fmt.Errorf("%d is invalid, use a number between 2 - %d", num, NUM_MAX_SLOTS)
	}

	if num > sb.NumMax {
		slot := StreamSlot{
			Length:    sb.Size,
			LengthMax: sb.Size,
			Content:   make([]byte, sb.Size),
		}
		for i := 0; i < num-sb.NumMax; i++ {
			sb.Slots = append(sb.Slots, slot)
		}
		sb.NumMax = num
	}
	sb.Num = num

	return err
}

// ---------------------------------E-----N-----D-----------------------------------
