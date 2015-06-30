//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for UDP Socket
//  - http://www.slideshare.net/feyeleanor/go-for-the-paranoid-network-programmer
//==================================================================================

package protoudp

import (
	"fmt"
	"log"
	"net"

	//"github.com/kisom/go-schannel"	// Bidirectional secure channels over TCP/IP

	pb "stoney/httpserver/src/protobase"
	sb "stoney/httpserver/src/streambase"
	sr "stoney/httpserver/src/streamring"
)

//---------------------------------------------------------------------------
const (
	STR_UDP_CASTER = "Happy Media UDP Caster"
	STR_UDP_SERVER = "Happy Media UDP Server"
	STR_UDP_PLAYER = "Happy Media UDP Player"
)

//---------------------------------------------------------------------------
type ProtoUdp struct {
	Host     string
	Port     string
	Desc     string
	Method   string // POST or GET
	Boundary string
	Conn     net.Conn
	Base     *pb.ProtoBase
}

//---------------------------------------------------------------------------
// string ProtoUdp information
//---------------------------------------------------------------------------
func (pt *ProtoUdp) String() string {
	str := fmt.Sprintf("\tHost: %s", pt.Host)
	str += fmt.Sprintf("\tPort: %s", pt.Port)
	str += fmt.Sprintf("\tBoundary: %s", pt.Boundary)
	str += fmt.Sprintf("\tMethod: %s", pt.Method)
	str += fmt.Sprintf("\tDesc: %s", pt.Desc)
	str += fmt.Sprintf("\tConn: %v", pt.Conn)
	return str
}

//---------------------------------------------------------------------------
// info handling
//---------------------------------------------------------------------------
func (pt *ProtoUdp) SetAddr(hname, hport, desc string) {
	pt.Host = hname
	pt.Port = hport
	pt.Desc = desc
}

func (pt *ProtoUdp) Reset() {
	pt.Host = sb.STR_DEF_HOST
	pt.Port = sb.STR_DEF_PORT
	pt.Boundary = sb.STR_DEF_BDRY
	pt.Desc = "reset"
	if pt.Conn != nil {
		pt.Conn.Close()
		pt.Conn = nil
	}
}

func (pt *ProtoUdp) Clear() {
	pt.Host = ""
	pt.Port = ""
	pt.Desc = ""
	if pt.Conn != nil {
		pt.Conn.Close()
		pt.Conn = nil
	}
}

//---------------------------------------------------------------------------
// new ProtoUdp struct in variadic style
//---------------------------------------------------------------------------
func NewProtoUdp(args ...string) *ProtoUdp {
	base := pb.NewProtoBase()

	pt := &ProtoUdp{
		Host:     "localhost",
		Port:     "8080",
		Boundary: sb.STR_DEF_BDRY,
		Base:     base,
	}

	for i, arg := range args {
		if i == 0 {
			pt.Host = arg
		} else if i == 1 {
			pt.Port = arg
		} else if i == 2 {
			pt.Desc = arg
		}
	}

	return pt
}

//---------------------------------------------------------------------------
// action points : Caster (1)-> [NET] ->(2) Server (3)-> [NET] ->(4) Player
//---------------------------------------------------------------------------
// Caster for test and debugging
//---------------------------------------------------------------------------
func (pt *ProtoUdp) ActCaster() error {
	var err error
	log.Printf("%s to %s:%s\n", STR_UDP_CASTER, pt.Host, pt.Port)

	return err
}

//---------------------------------------------------------------------------
// Server for debugging
//---------------------------------------------------------------------------
func (pt *ProtoUdp) ActServer(ring *sr.StreamRing) error {
	var err error
	log.Printf("%s on :%s\n", STR_UDP_SERVER, pt.Port)

	return err
}

//---------------------------------------------------------------------------
// Player to receive data in multipart
//---------------------------------------------------------------------------
func (pt *ProtoUdp) ActPlayer(ring *sr.StreamRing) error {
	var err error
	log.Printf("%s\n", STR_UDP_PLAYER)

	return err
}

// ---------------------------------E-----N-----D--------------------------------
