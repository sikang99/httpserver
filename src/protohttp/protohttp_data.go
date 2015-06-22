//=========================================================================
// Author : Stoney Kang, sikang99@gmail.com, 2015
// Protocol for HTTP streaming
//=========================================================================

package protohttp

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	sb "stoney/httpserver/src/streambase"
	sr "stoney/httpserver/src/streamring"
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
// new client config transport with timeout
//---------------------------------------------------------------------------
var timeout = time.Duration(3 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

func NewClientConfig() *http.Client {
	// simple timeout and tls setting
	tp := &http.Transport{
		Dial:            dialTimeout,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &http.Client{Transport: tp, Timeout: timeout}
}

//-----------------------------------------------------------------------------
// server config
//-----------------------------------------------------------------------------
type ServerConfig struct {
	Title        string
	Image        string
	Url          string
	Addr         string
	Host         string
	Port         string
	PortS        string
	Port2        string
	Mode         string
	Ring         *sr.StreamRing
	ImageChannel chan []byte
	// http://giantmachines.tumblr.com/post/52184842286/golang-http-client-with-timeouts
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
}

func (sc *ServerConfig) String() string {
	str := fmt.Sprintf("\tTitle: %s", sc.Title)
	str += fmt.Sprintf("\tMode: %s", sc.Mode)
	return str
}

func NewServerConfig() *ServerConfig {
	sc := &ServerConfig{
		ImageChannel: make(chan []byte, 2)}

	sc.Title = "Happy Media System: MJPEG"
	sc.Image = "static/image/gophergun.png"
	sc.Addr = "http://localhost"
	sc.Host = sb.STR_DEF_HOST
	sc.Port = sb.STR_DEF_PORT
	sc.PortS = sb.STR_DEF_PTLS

	sc.Ring = PrepareRing(3, sb.MBYTE, "Server ring buffer")

	return sc
}

var conf = NewServerConfig()

// ---------------------------------E-----N-----D--------------------------------
