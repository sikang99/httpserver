//=========================================================================
// Author : Stoney Kang, sikang99@gmail.com, 2015
// Protocol for HTTP streaming
//=========================================================================

package protohttp

import (
	"time"

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
type ServerConfig struct {
	Title        string
	Image        string
	Addr         string
	Mode         string
	Ring         *sr.StreamRing
	ImageChannel chan []byte
	// http://giantmachines.tumblr.com/post/52184842286/golang-http-client-with-timeouts
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
	Http             *ProtoHttp
}

//---------------------------------------------------------------------------
// new config struct
//---------------------------------------------------------------------------
func NewServerConfig() *ServerConfig {
	sc := &ServerConfig{
		ImageChannel: make(chan []byte, 2)}

	sc.Title = "Happy Media System: MJPEG"
	sc.Image = "static/image/gophergun.png"
	sc.Addr = "http://localhost"

	return sc
}

var conf = NewServerConfig()

// ---------------------------------E-----N-----D--------------------------------
