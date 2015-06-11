//==================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Base information for streaming
// Author : Stoney Kang, sikang99@gmail.com
//==================================================================================

package streambase

import (
	"errors"
	"fmt"
	"log"
	"net"
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
	STR_DEF_HOST = "localhost"
	STR_DEF_PORT = "8080"
	STR_DEF_PTLS = "8443"
	STR_DEF_BDRY = "myboundary"
	STR_DEF_PATN = "*.jpg"
)

const (
	STATUS_IDLE = iota
	STATUS_USING

	LEN_MAX_LINE = 128

	TIME_DEF_WAIT      = 100 * time.Nanosecond
	TIME_DEF_PRECISION = time.Millisecond
)

var (
	ErrEmpty   = errors.New("empty")
	ErrFull    = errors.New("full")
	ErrNull    = errors.New("null")
	ErrSize    = errors.New("size")
	ErrStatus  = errors.New("invalid status")
	ErrSupport = errors.New("not supported")
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
func GetTimestamp() int64 {
	switch TIME_DEF_PRECISION {
	case time.Second:
		return GetTimestampSecond()
	case time.Millisecond:
		return GetTimestampMillisecond()
	case time.Microsecond:
		return GetTimestampMicrosecond()
	case time.Nanosecond:
		return GetTimestampNanosecond()
	default:
		return 0
	}
}

//---------------------------------------------------------------------------
// get wait time from timestamp difference
//---------------------------------------------------------------------------
func GetDuration(value int64) time.Duration {
	switch TIME_DEF_PRECISION {
	case time.Second:
		value = value * 1000 * 1000 * 1000
	case time.Millisecond:
		value = value * 1000 * 1000
	case time.Microsecond:
		value = value * 1000
	case time.Nanosecond:
	}

	return time.Duration(value)
}

//---------------------------------------------------------------------------
// show network interfaces
//---------------------------------------------------------------------------
func ShowNetInterfaces() {
	list, err := net.Interfaces()
	if err != nil {
		log.Println(err)
	}

	for i, iface := range list {
		fmt.Printf("%d %s %v\n", i, iface.Name, iface)
		addrs, err := iface.Addrs()
		if err != nil {
			log.Println(err)
		}
		for j, addr := range addrs {
			fmt.Printf("\t%d %v\n", j, addr)
		}
	}
}

//---------------------------------------------------------------------------
// function template
//---------------------------------------------------------------------------
func malfunction() error {
	var err error

	return err
}

// ---------------------------------E-----N-----D-----------------------------------
