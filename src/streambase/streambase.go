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
// make timestamp in sec, msec, nsec
// - https://blog.cloudflare.com/its-go-time-on-linux/
// - https://medium.com/coding-and-deploying-in-the-cloud/time-stamps-in-golang-abcaf581b72f
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

func MakeTimestamp() int64 {
	switch TIME_DEF_PRECISION {
	case time.Second:
		return MakeTimestampSecond()
	case time.Millisecond:
		return MakeTimestampMillisecond()
	case time.Nanosecond:
		return MakeTimestampNanosecond()
	default:
		return 0
	}
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
//
//---------------------------------------------------------------------------
func malfunction() error {
	var err error

	return err
}

// ---------------------------------E-----N-----D-----------------------------------
