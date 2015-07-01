//=================================================================================
// Happy Media System
// one program including agents such as caster, server, player, monitor
// Author : Stoney Kang, sikang99@gmail.com
//=================================================================================
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	mc "stoney/httpserver/src/mediaconf"

	pf "stoney/httpserver/src/protofile"
	pt "stoney/httpserver/src/prototcp"
	pw "stoney/httpserver/src/protows"

	sb "stoney/httpserver/src/streambase"
)

//---------------------------------------------------------------------------
const (
	STR_MEDIA_VERSION = "0.9.9, 2015/06/30"
	STR_MEDIA_SYSTEM  = "Happy Media System"
)

//---------------------------------------------------------------------------
var (
	fmode  = flag.String("m", "player", "Working mode of program")
	fhost  = flag.String("host", sb.STR_DEF_HOST, "server host address")
	fport  = flag.String("port", sb.STR_DEF_PORT, "TCP port to be used for http")
	fports = flag.String("ports", sb.STR_DEF_PTLS, "TCP port to be used for https")
	fport2 = flag.String("port2", sb.STR_DEF_PORT2, "TCP port to be used for http2")
	furl   = flag.String("url", "http://"+sb.STR_DEF_HOST+":"+sb.STR_DEF_PORT, "base url to be accessed")
	froot  = flag.String("root", ".", "Define the root filesystem path")
	vflag  = flag.Bool("verbose", false, "Verbose display")
)

//---------------------------------------------------------------------------
// init for main
//---------------------------------------------------------------------------
func init() {
	// maximize concurrency
	runtime.GOMAXPROCS(runtime.NumCPU())

	//log.SetOutput(os.Stdout)
	//log.SetPrefix("TRACE: ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// flag setting and parsing
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: %v [flags], v.%s\n\n", os.Args[0], STR_MEDIA_VERSION)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		os.Exit(1)
	}

	flag.Parse()
}

//---------------------------------------------------------------------------
// a single program including client and server in go style
//---------------------------------------------------------------------------
func main() {
	// check arguments
	if flag.NFlag() == 0 && flag.NArg() == 0 {
		flag.Usage()
	}

	// set command parameters to config
	sc := mc.NewServerConfig()

	sc.Mode = *fmode
	sc.Url = *furl
	sc.Host = *fhost
	sc.Port = *fport
	sc.PortS = *fports
	sc.Port2 = *fport2

	fmt.Printf("%s, v.%s\n", STR_MEDIA_SYSTEM, STR_MEDIA_VERSION)
	fmt.Printf("Default ports: %s,%s,%s\n", sc.Port, sc.PortS, sc.Port2)
	fmt.Printf("Working mode: %s\n", sc.Mode)

	//hp := ph.NewProtoHttpWithPorts(sc.Port, sc.PortS, sc.Port2)
	tp := pt.NewProtoTcpWithPorts("8087")
	wp := pw.NewProtoWsWithPorts("8087", "8443")

	ring := sc.Array[0]

	// let's do work by the working mode
	switch sc.Mode {

	// package protohttp
	case "http_reader":
		sc.StreamReader(ring, sc.Url)
	case "http_player":
		sc.StreamPlayer(ring, sc.Url)
	case "http_caster":
		sc.StreamCaster(sc.Url)
	case "http_server":
		sc.StreamServer(ring)
	case "http_monitor":
		sc.StreamMonitor(sc.Url)

	// package prototcp
	case "tcp_caster":
		tp.StreamCaster()
	case "tcp_server":
		tp.StreamServer(ring)
	case "tcp_player":
		tp.StreamPlayer(ring)

	// package protows
	case "ws_caster":
		wp.StreamCaster()
	case "ws_server":
		wp.StreamServer()
	case "ws_player":
		wp.StreamPlayer()

	// package protofile
	case "file_reader":
		fr := pf.NewProtoFile("./static/image/*.jpg", "F-Rr")
		fr.StreamReader(ring)
	case "file_writer":
		fw := pf.NewProtoFile("output.mjpg", "F-Wr")
		fw.StreamWriter(ring)

	default:
		fmt.Println("Unknown working mode")
		os.Exit(0)
	}
}

// ---------------------------------E-----N-----D--------------------------------
