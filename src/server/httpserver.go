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
	"net/url"
	"os"
	"runtime"

	pf "stoney/httpserver/src/protofile"
	ph "stoney/httpserver/src/protohttp"
	pt "stoney/httpserver/src/prototcp"
	pw "stoney/httpserver/src/protows"

	sb "stoney/httpserver/src/streambase"
)

//---------------------------------------------------------------------------
var index_tmpl = `<!DOCTYPE html>
<html>
<head>
</head>
<body>
<center>
<h2>Hello! from Stoney Kang, a Novice Gopher</h2>.
<img src="{{ .Image }}">Gopher with a gun</img>
</center>
</body>
</html>
`

var hello_tmpl = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8" />
<script type="text/javascript" src="static/eventemitter2.min.js"></script>
<script type="text/javascript" src="static/mjpegcanvas.min.bak.js"></script>
  
<script type="text/javascript" type="text/javascript">
  function init() {
    var viewer = new MJPEGCANVAS.Viewer({
      divID : 'mjpeg',
      host : '{{ .Addr }}',
      port : {{ .Port }},
      width : 1024,
      height : 768,
      topic : 'agilecam'
    });
  }
</script>
</head>

<body onload="init()">
<center>
  <h1>{{ .Title }}</h1>
  <div id="mjpeg"></div>
</center>
</body>
</html>
`

//---------------------------------------------------------------------------
const (
	STR_MEDIA_VERSION = "0.9.3"
	STR_MEDIA_SYSTEM  = "Happy Media System"
)

//---------------------------------------------------------------------------
var (
	fmode  = flag.String("m", "player", "Working mode of program [caster|server|player|reader|sender|receiver|shooter|catcher]")
	fhost  = flag.String("host", "localhost", "server host address")
	fport  = flag.String("port", sb.STR_DEF_PORT, "TCP port to be used for http")
	fports = flag.String("ports", sb.STR_DEF_PTLS, "TCP port to be used for https")
	fport2 = flag.String("port2", sb.STR_DEF_PORT2, "TCP port to be used for http2")
	furl   = flag.String("url", "http://localhost:8000/[index|hello|/stream]", "url to be accessed")
	froot  = flag.String("root", ".", "Define the root filesystem path")
	vflag  = flag.Bool("verbose", false, "Verbose display")
)

var conf = ph.NewServerConfig()

//---------------------------------------------------------------------------
// init for main
//---------------------------------------------------------------------------
func init() {
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

	url, err := url.ParseRequestURI(*furl)
	if err != nil {
		log.Println(err)
		return
	}

	// set command parameters to config
	conf.Url = *furl
	conf.Host = *fhost
	conf.Port = *fport
	conf.PortS = *fports
	conf.Port2 = *fport2
	conf.Mode = *fmode

	fmt.Printf("%s, v.%s\n", STR_MEDIA_SYSTEM, STR_MEDIA_VERSION)
	fmt.Printf("Default config: %s %s\n", url.Scheme, url.Host)
	fmt.Printf("Working mode: %s\n", conf.Mode)

	hp := ph.NewProtoHttpWithPorts(conf.Port, conf.PortS, conf.Port2)

	// let's do by the working mode
	switch conf.Mode {

	// package protohttp
	case "http_reader":
		hp.StreamReader(conf.Url, conf.Ring)
	case "http_player":
		hp.StreamPlayer(conf.Url, conf.Ring)
	case "http_caster":
		hp.StreamCaster(conf.Url)
	case "http_server":
		hp.StreamServer(conf.Ring)
	case "http_monitor":
		hp.StreamMonitor(conf.Url)

	// package prototcp
	case "tcp_caster":
		pt.NewProtoTcpWithPorts("8087").StreamCaster()
	case "tcp_server":
		ts := pt.NewProtoTcpWithPorts("8087")
		ts.StreamServer(conf.Ring)
	case "tcp_player":
		tp := pt.NewProtoTcpWithPorts("8087")
		tp.StreamPlayer(conf.Ring)

	// package protows
	case "ws_caster":
		wc := pw.NewProtoWsWithPorts("8087", "8443")
		wc.StreamCaster()
	case "ws_server":
		ws := pw.NewProtoWsWithPorts("8087", "8443")
		ws.StreamServer()
	case "ws_player":
		wp := pw.NewProtoWsWithPorts("8087", "8443")
		wp.StreamPlayer()

	// package protofile
	case "file_reader":
		fr := pf.NewProtoFile("./static/image/*.jpg", "F-Rr")
		fr.StreamReader(conf.Ring)
	case "file_writer":
		fw := pf.NewProtoFile("output.mjpg", "F-Wr")
		fw.StreamWriter(conf.Ring)

	default:
		fmt.Println("Unknown working mode")
		os.Exit(0)
	}
}

// ---------------------------------E-----N-----D--------------------------------
