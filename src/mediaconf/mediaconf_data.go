//=========================================================================
// Author : Stoney Kang, sikang99@gmail.com, 2015
// Protocol for HTTP streaming
//=========================================================================

package mediaconf

import (
	"fmt"
	"time"

	pb "stoney/httpserver/src/protobase"

	sb "stoney/httpserver/src/streambase"
	si "stoney/httpserver/src/streaminfo"
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

//-----------------------------------------------------------------------------
// server config
// - http://stackoverflow.com/questions/31014838/parsing-json-into-a-struct
//-----------------------------------------------------------------------------
type ServerConfig struct {
	Title    string `json:"title"`
	Image    string
	Url      string
	Addr     string
	Host     string
	Port     string
	PortS    string
	Port2    string
	Mode     string
	Array    []*sr.StreamRing
	Station  []*si.Channel
	Actors   map[string]*pb.ProtoBase
	NotiChan chan []byte
	// http://giantmachines.tumblr.com/post/52184842286/golang-http-client-with-timeouts
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
}

//-----------------------------------------------------------------------------
// server config
//-----------------------------------------------------------------------------
func (sc *ServerConfig) String() string {
	str := fmt.Sprintf("\tTitle: %s", sc.Title)
	str += fmt.Sprintf("\tMode: %s", sc.Mode)
	str += fmt.Sprintf("\tAddr: %s", sc.Addr)
	str += fmt.Sprintf("\tUrl: %s", sc.Url)
	return str
}

//-----------------------------------------------------------------------------
// server config
//-----------------------------------------------------------------------------
func NewServerConfig() *ServerConfig {
	sc := &ServerConfig{
		NotiChan: make(chan []byte, 2),
		Actors:   make(map[string]*pb.ProtoBase),
	}

	sc.Title = "Happy Media System: MJPEG"
	sc.Image = "static/image/gophergun.png"
	sc.Addr = "http://localhost"
	sc.Host = sb.STR_DEF_HOST
	sc.Port = sb.STR_DEF_PORT
	sc.PortS = sb.STR_DEF_PTLS
	sc.Port2 = sb.STR_DEF_PORT2

	sc.Array = sr.NewStreamArrayWithSize(3, 3, sb.MBYTE)

	return sc
}

// ---------------------------------E-----N-----D--------------------------------
