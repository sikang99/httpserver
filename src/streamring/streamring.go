//==================================================================================
// Author : Stoney Kang, sikang99@gmail.com, 2015
// Circular Stream Buffer : Ring
// - https://github.com/zfjagann/golang-ring
// - https://github.com/glycerine/rbuf
// - http://blog.pivotal.io/labs/labs/a-concurrent-ring-buffer-for-go
//==================================================================================

package streamring

import (
	"fmt"
	"mime"
	"strconv"
	"strings"
	"sync"

	sb "stoney/httpserver/src/streambase"
)

//----------------------------------------------------------------------------------
const (
	LEN_DEF_SLOT = sb.MBYTE
	LEN_MAX_SLOT = sb.GBYTE

	NUM_DEF_SLOTS = 30 // 30fps
	NUM_MAX_SLOTS = 1024
)

//==================================================================================
// stream slot struc
//----------------------------------------------------------------------------------
type StreamSlot struct {
	sync.Mutex
	Type      string
	Length    int
	LengthMax int
	Content   []byte
	Timestamp int64
}

//----------------------------------------------------------------------------------
// make a new slot of given max size
//----------------------------------------------------------------------------------
func NewStreamSlot() *StreamSlot {
	return &StreamSlot{
		Length:    0,
		LengthMax: LEN_DEF_SLOT,
		Content:   make([]byte, LEN_DEF_SLOT),
	}
}

func NewStreamSlotBySize(cmax int) *StreamSlot {
	return &StreamSlot{
		Length:    0,
		LengthMax: cmax,
		Content:   make([]byte, cmax),
	}
}

func NewStreamSlotBySlot(cmax int, in *StreamSlot) *StreamSlot {
	return &StreamSlot{
		Type:      in.Type,
		Length:    in.Length,
		LengthMax: cmax,
		Content:   in.Content,
	}
}

func NewStreamSlotByData(cmax int, ctype string, clen int, cdata []byte) *StreamSlot {
	return &StreamSlot{
		Type:      ctype,
		Length:    clen,
		LengthMax: cmax,
		Content:   cdata,
	}
}

//----------------------------------------------------------------------------------
// string information for the single slot
//----------------------------------------------------------------------------------
func (ss *StreamSlot) String() string {
	str := fmt.Sprintf("\tTimestamp: %v", ss.Timestamp)
	str += fmt.Sprintf("\tType: %v", ss.Type)
	str += fmt.Sprintf("\tLength: %v/%v(%v)", ss.Length, ss.LengthMax, len(ss.Content))
	str += fmt.Sprintf("\tContent: ")
	if ss.Length > 1 {
		if strings.Contains(ss.Type, "text/") {
			str += fmt.Sprintf("%s [%0x:%0x]", string(ss.Content[:ss.Length]), ss.Content[:2], ss.Content[ss.Length-2:ss.Length])
		} else {
			str += fmt.Sprintf("[%0x-%0x]", ss.Content[:2], ss.Content[ss.Length-2:ss.Length])
		}
	}
	return str
}

//----------------------------------------------------------------------------------
// check media type, its major and sub type of content
//----------------------------------------------------------------------------------
func (ss *StreamSlot) IsType(ctype string) bool {
	mt, _, _ := mime.ParseMediaType(ss.Type)
	return strings.EqualFold(mt, ctype)
}

func (ss *StreamSlot) IsMajorType(ctype string) bool {
	var res bool

	mt, _, _ := mime.ParseMediaType(ss.Type)
	st := strings.Split(mt, "/")
	if len(st) > 0 {
		mt, ctype = strings.ToUpper(st[0]), strings.ToUpper(ctype)
		res = strings.Contains(mt, ctype)
	}

	return res
}

func (ss *StreamSlot) IsSubType(ctype string) bool {
	var res bool

	mt, _, _ := mime.ParseMediaType(ss.Type)
	st := strings.Split(mt, "/")
	if len(st) > 1 {
		mt, ctype = strings.ToUpper(st[1]), strings.ToUpper(ctype)
		res = strings.Contains(mt, ctype)
	}

	return res
}

//==================================================================================
// stream ring struc
//----------------------------------------------------------------------------------
type StreamRing struct {
	sync.Mutex
	Id       string // Name or Id?
	Status   int
	Num      int    // number of slots used
	NumMax   int    // number of slots allocated
	Size     int    // size of the slot content
	In       int    // input position of buffer to be written
	Out      int    // output position of buffer to be read
	Boundary string // description of buffer
	Desc     string // description of buffer
	Slots    []StreamSlot
}

//----------------------------------------------------------------------------------
// make a new circular stream buffer
//----------------------------------------------------------------------------------
func NewStreamRingWithParams(num int, size int, desc string) *StreamRing {
	slots := make([]StreamSlot, num)
	for i := 0; i < num; i++ {
		slots[i].Content = make([]byte, size)
		slots[i].LengthMax = size
	}

	return &StreamRing{
		Slots:  slots,
		Status: sb.STATUS_IDLE,
		Num:    num, NumMax: num,
		Size: size,
		In:   0, Out: 0,
		Boundary: sb.STR_DEF_BDRY,
		Desc:     desc,
	}
}

func NewStreamRingWithSize(num int, size int) *StreamRing {
	return NewStreamRingWithParams(num, size, "New sized ring buffer")
}

func NewStreamRing() *StreamRing {
	return NewStreamRingWithParams(3, sb.MBYTE, "New default ring buffer")
}

//----------------------------------------------------------------------------------
// make a new ring arrary(slice)
//----------------------------------------------------------------------------------
func NewStreamArrayWithSize(rnum, snum, size int) []*StreamRing {
	var array []*StreamRing
	for i := 0; i < rnum; i++ {
		array = append(array, NewStreamRingWithParams(snum, size, strconv.Itoa(i)+"-th ring"))
	}
	return array
}

//----------------------------------------------------------------------------------
// string information for the stream buffer
//----------------------------------------------------------------------------------
func (sr *StreamRing) BaseString() string {
	str := fmt.Sprintf("[StreamRing] %s", sr.Id)
	str += fmt.Sprintf("\tStatus: %s(%d)", sb.StatusText[sr.Status], sr.Status)
	str += fmt.Sprintf("\tPos: %d,%d", sr.In, sr.Out)
	str += fmt.Sprintf("\tSize: %d/%d, %d KB", sr.Num, sr.NumMax, sr.Size/sb.KBYTE)
	str += fmt.Sprintf("\tBoundary: %s", sr.Boundary)
	str += fmt.Sprintf("\tDesc: %s", sr.Desc)
	return str
}

func (sr *StreamRing) String() string {
	str := fmt.Sprintf("%s\n", sr.BaseString())
	for i := 0; i < sr.Num; i++ {
		str += fmt.Sprintf("\t[%d] %s\n", i, sr.Slots[i].String())
	}

	return str
}

//----------------------------------------------------------------------------------
// get the number of slots in buffer for used(len) and allocted(cap)
//----------------------------------------------------------------------------------
func (sr *StreamRing) Len() int {
	return sr.Num
}

func (sr *StreamRing) Cap() int {
	return sr.NumMax
}

//----------------------------------------------------------------------------------
// set status of buffer
//----------------------------------------------------------------------------------
func (sr *StreamRing) SetStatus(status int) int {
	sr.Status = status
	return sr.Status
}

func (sr *StreamRing) SetStatusUsing() error {
	var err error
	if sr.Status != sb.STATUS_IDLE {
		return sb.ErrStatus
	}
	sr.Status = sb.STATUS_USING
	return err
}

func (sr *StreamRing) SetStatusIdle() error {
	var err error
	if sr.Status != sb.STATUS_USING {
		return sb.ErrStatus
	}
	sr.Status = sb.STATUS_IDLE
	return err
}

func (sr *StreamRing) GetStatus() int {
	return sr.Status
}

func (sr *StreamRing) IsUsing() bool {
	return sr.Status == sb.STATUS_USING
}

func (sr *StreamRing) IsIdle() bool {
	return sr.Status == sb.STATUS_IDLE
}

//----------------------------------------------------------------------------------
// set the position of slot to be read and written
//----------------------------------------------------------------------------------
func (sr *StreamRing) SetPosInByPos(pos int) int {
	sr.In = (pos % sr.Num)
	return sr.In
}

func (sr *StreamRing) SetPosOutByPos(pos int) int {
	sr.Out = (pos % sr.Num)
	return sr.Out
}

//----------------------------------------------------------------------------------
// get the current position of slot to read and write
//----------------------------------------------------------------------------------
func (sr *StreamRing) GetPosIn() int {
	return sr.In
}

func (sr *StreamRing) GetPosOut() int {
	return sr.Out
}

//----------------------------------------------------------------------------------
// get the current slot to read and write
//----------------------------------------------------------------------------------
func (sr *StreamRing) GetSlotIn() (*StreamSlot, int) {
	slot := &sr.Slots[sr.In]
	return slot, sr.In
}

func (sr *StreamRing) GetSlotOut() (*StreamSlot, int) {
	slot := &sr.Slots[sr.Out]
	return slot, sr.Out
}

//----------------------------------------------------------------------------------
// get the slot designated
//----------------------------------------------------------------------------------
func (sr *StreamRing) GetSlotByPos(pos int) (*StreamSlot, error) {
	var err error

	pos = pos % sr.Num
	slot := &sr.Slots[pos]

	return slot, err
}

//----------------------------------------------------------------------------------
// get the slot to be read and move to the next
//----------------------------------------------------------------------------------
func (sr *StreamRing) GetSlotOutNext() (*StreamSlot, error) {
	var err error

	// no data to read
	if sr.In == sr.Out {
		return nil, sb.ErrEmpty
	}

	slot := &sr.Slots[sr.Out]
	sr.Out = (sr.Out + 1) % sr.Num

	return slot, err
}

//----------------------------------------------------------------------------------
// get the pointer of slot designated and go to the next
//----------------------------------------------------------------------------------
func (sr *StreamRing) GetSlotNextByPos(pos int) (*StreamSlot, int, error) {
	var err error

	pos = pos % sr.Num

	// no data to read
	if sr.In == pos {
		return nil, pos, sb.ErrEmpty
	}

	slot := &sr.Slots[pos]
	pos = (pos + 1) % sr.Num

	return slot, pos, err
}

//----------------------------------------------------------------------------------
// write the information to the slot and go to the next
//----------------------------------------------------------------------------------
func (sr *StreamRing) PutSlotInNext(slot *StreamSlot) (*StreamSlot, error) {
	sr.Lock()
	defer sr.Unlock()

	var err error

	if slot.Length > sr.Size {
		return nil, fmt.Errorf("too big data size")
	}

	st := &sr.Slots[sr.In]

	st.Type = slot.Type
	st.Length = slot.Length
	copy(st.Content, slot.Content)

	sr.In = (sr.In + 1) % sr.Num

	return st, err
}

//----------------------------------------------------------------------------------
// write the information to the slot designated
//----------------------------------------------------------------------------------
func (sr *StreamRing) PutSlotInByPos(slot *StreamSlot, pos int) (*StreamSlot, error) {
	sr.Lock()
	defer sr.Unlock()

	var err error

	pos = (pos % sr.Num)
	st := &sr.Slots[pos]

	st.Type = slot.Type
	st.Length = slot.Length
	copy(st.Content, slot.Content)

	return slot, err
}

//----------------------------------------------------------------------------------
// reset(clear) the stream buffer
//----------------------------------------------------------------------------------
func (sr *StreamRing) Reset() {
	sr.Lock()
	defer sr.Unlock()

	for i := 0; i < sr.NumMax; i++ {
		sr.Slots[i].Type = ""
		sr.Slots[i].Length = 0
	}

	sr.In = 0
	sr.Out = 0
	sr.Num = sr.NumMax
	sr.Status = sb.STATUS_IDLE
	sr.Desc = "Buffer is reset"
}

//----------------------------------------------------------------------------------
// change the size of buffer, i.e, the number of slots
//----------------------------------------------------------------------------------
func (sr *StreamRing) Resize(num int) error {
	sr.Lock()
	defer sr.Unlock()

	var err error

	if num < 2 || num > NUM_MAX_SLOTS {
		return fmt.Errorf("%d is invalid, use a number between 2 - %d", num, NUM_MAX_SLOTS)
	}

	if num > sr.NumMax {
		slot := StreamSlot{
			Length:    sr.Size,
			LengthMax: sr.Size,
			Content:   make([]byte, sr.Size),
		}
		for i := 0; i < num-sr.NumMax; i++ {
			sr.Slots = append(sr.Slots, slot)
		}
		sr.NumMax = num
	}
	sr.Num = num

	return err
}

//----------------------------------------------------------------------------------
// read out the slot to new one
//----------------------------------------------------------------------------------
func (sr *StreamRing) ReadSlotIn() (*StreamSlot, int) {
	in := &sr.Slots[sr.In]
	slot := NewStreamSlotByData(sr.Size, in.Type, in.Length, in.Content)
	return slot, sr.In
}

// ---------------------------------E-----N-----D-----------------------------------
