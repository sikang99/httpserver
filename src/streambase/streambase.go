//==================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Base information for streaming
// Author : Stoney Kang, sikang99@gmail.com
//==================================================================================

package streambase

import (
	"errors"
	"time"
)

//----------------------------------------------------------------------------------
const (
	KBYTE = 1024
	MBYTE = 1024 * KBYTE // Kilo
	GBYTE = 1024 * MBYTE // Giga
	TBYTE = 1024 * GBYTE // Tera
	HBYTE = 1024 * TBYTE // Hexa
)

const (
	LEN_MAX_LINE = 128

	STATUS_IDLE = iota
	STATUS_USING

	STR_DEF_HOST = "localhost"
	STR_DEF_PORT = "8080"
	STR_DEF_BDRY = "myboundary"
	STR_DEF_PATN = "*.jpg"

	TIME_DEF_WAIT = time.Millisecond
)

var (
	ErrEmpty  = errors.New("empty")
	ErrFull   = errors.New("full")
	ErrStatus = errors.New("invalid status")
)

//---------------------------------------------------------------------------
// make timestamp in sec, msec, nsec
//  - https://medium.com/coding-and-deploying-in-the-cloud/time-stamps-in-golang-abcaf581b72f
//---------------------------------------------------------------------------
func MakeTimestampNanosecond() int64 {
	return time.Now().UnixNano()
}

func MakeTimestampMillisecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func MakeTimestampSecond() int64 {
	return time.Now().Unix()
}

// ---------------------------------E-----N-----D-----------------------------------
