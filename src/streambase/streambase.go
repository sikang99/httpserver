//==================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Base information for streaming
// Author : Stoney Kang, sikang99@gmail.com
//==================================================================================

package streambase

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"time"

	"github.com/fatih/color"

	"golang.org/x/crypto/ssh/terminal"
)

//----------------------------------------------------------------------------------
const (
	_     = iota
	KBYTE = 1024             // Kilo byte
	MBYTE = 1 << (10 * iota) // Mega
	GBYTE                    // Giga
	TBYTE                    // Tera
	PBYTE                    // Peta
	EBYTE                    // Exa
	ZBYTE                    // Zetta
	YBYTE                    // Yotta
)

const (
	STR_DEF_MODE  = "normal" // or "secure"
	STR_DEF_HOST  = "localhost"
	STR_DEF_PORT  = "8080"
	STR_DEF_PTLS  = "8081" // for HTTPS
	STR_DEF_PORT2 = "8082" // for HTTP2
	STR_DEF_PORTM = "8088" // for Monitor
	STR_DEF_BDRY  = "myboundary"
	STR_DEF_PATN  = "*.jpg"
)

const (
	LEN_MAX_LINE = 128
	LEN_MAX_MSG  = 1024

	TIME_DEF_WAIT      = 100 * time.Microsecond
	STR_TIME_PRECISION = "Millisecond"
)

const (
	STATUS_IDLE = iota
	STATUS_USING
	STATUS_CLOSE
	STATUS_RUN
	STATUS_PAUSE
	STATUS_STOP
	STATUS_RESUME
)

var StatusText = map[int]string{
	// for ring or other objects
	STATUS_IDLE:  "Idle",
	STATUS_USING: "Using",
	STATUS_CLOSE: "Close",
	// for actor or worker
	STATUS_RUN:    "Run",
	STATUS_PAUSE:  "Pause",
	STATUS_STOP:   "Stop",
	STATUS_RESUME: "Resume",
}

// multipart headers
const (
	STR_HDR_SERVER         = "Server"
	STR_HDR_USER_AGENT     = "User-Agent"
	STR_HDR_CONTENT_TYPE   = "Content-Type"
	STR_HDR_CONTENT_LENGTH = "Content-Length"
	STR_HDR_TIMESTAMP      = "X-Timestamp"
	STR_HDR_AUDIO_FORMAT   = "X-Audio-Format"
	STR_HDR_VIDEO_FORMAT   = "X-Video-Format"
	STR_HDR_GPS_FORMAT     = "X-GPS-Format"
)

var (
	ErrFound   = errors.New("error not found")
	ErrParse   = errors.New("error to parse")
	ErrEmpty   = errors.New("error empty")
	ErrFull    = errors.New("error full")
	ErrNull    = errors.New("error null")
	ErrSize    = errors.New("error size")
	ErrStatus  = errors.New("error invalid status")
	ErrValue   = errors.New("error invalid value")
	ErrSupport = errors.New("error not supported")
)

//---------------------------------------------------------------------------
type Timestamp struct {
	Scale time.Time
	Value int64
}

//---------------------------------------------------------------------------
// make timestamp in sec, msec, nsec
// - https://blog.cloudflare.com/its-go-time-on-linux/
// - https://medium.com/coding-and-deploying-in-the-cloud/time-stamps-in-golang-abcaf581b72f
//---------------------------------------------------------------------------
func GetTimestampNanosecond() int64 {
	return time.Now().UnixNano()
}

func GetTimestampMicrosecond() int64 {
	return time.Now().UnixNano() / int64(time.Microsecond)
}

func GetTimestampMillisecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func GetTimestampSecond() int64 {
	return time.Now().Unix()
}

//---------------------------------------------------------------------------
// get current timestamp
//---------------------------------------------------------------------------
func GetTimestampFromString(str string) int64 {
	tstamp, _ := strconv.ParseInt(str, 0, 64)
	return tstamp
}

//---------------------------------------------------------------------------
// get current timestamp
//---------------------------------------------------------------------------
func GetTimestampNow() int64 {
	switch STR_TIME_PRECISION {
	case "Second":
		return GetTimestampSecond()
	case "Millisecond":
		return GetTimestampMillisecond()
	case "Microsecond":
		return GetTimestampMicrosecond()
	case "Nanosecond":
		return GetTimestampNanosecond()
	}

	return 0
}

//---------------------------------------------------------------------------
// get wait time from timestamp difference
//---------------------------------------------------------------------------
func GetDuration(value int64) time.Duration {
	switch STR_TIME_PRECISION {
	case "Second":
		value = value * 1000 * 1000 * 1000
	case "Millisecond":
		value = value * 1000 * 1000
	case "Microsecond":
		value = value * 1000
	case "Nanosecond":
	}

	return time.Duration(value)
}

//---------------------------------------------------------------------------
// show network interfaces
//---------------------------------------------------------------------------
func ShowNetInterfaces() string {
	list, err := net.Interfaces()
	if err != nil {
		log.Println(err)
		return ""
	}

	var str string
	for i, iface := range list {
		str += fmt.Sprintf("%d %s %v\n", i, iface.Name, iface)
		addrs, err := iface.Addrs()
		if err != nil {
			log.Println(err)
		}
		for j, addr := range addrs {
			str += fmt.Sprintf("\t%d %v\n", j, addr)
		}
	}

	return str
}

//---------------------------------------------------------------------------
// dump data in hex format
// usage : HexDump(data[10:20])
//---------------------------------------------------------------------------
func HexDump(data []byte) {
	fmt.Println(hex.Dump(data))
}

//---------------------------------------------------------------------------
// check if stdin is terminal or not
// - http://rosettacode.org/wiki/Check_input_device_is_a_terminal
//---------------------------------------------------------------------------
func IsTerminal() bool {
	return terminal.IsTerminal(int(os.Stdin.Fd()))
}

//---------------------------------------------------------------------------
// print interface in colored log
//---------------------------------------------------------------------------
func RedString(obj interface{}) string {
	return color.RedString(fmt.Sprint(obj))
}

//---------------------------------------------------------------------------
// get function name of given
//---------------------------------------------------------------------------
func GetFuncName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}

//---------------------------------------------------------------------------
// get function name of current position
//---------------------------------------------------------------------------
func FName() string {
	pc, _, _, _ := runtime.Caller(0)
	return runtime.FuncForPC(pc).Name()
}

//---------------------------------------------------------------------------
// get function name of caller
//---------------------------------------------------------------------------
func FuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

//---------------------------------------------------------------------------
// trace
//---------------------------------------------------------------------------
func Trace() {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	fmt.Printf("%s:%d %s\n", file, line, f.Name())
}

//---------------------------------------------------------------------------
// sanity check
//---------------------------------------------------------------------------
func AskOrder(ch chan string, order string) string {
	ch <- order

	select {
	case str := <-ch:
		return str
	default:
		return ""
	}
}

func AnswerStatus(ch chan string) string {
	select {
	case str := <-ch:
		log.Println(str)
		return str
	default:
		return ""
	}
}

// ---------------------------------E-----N-----D-----------------------------------
