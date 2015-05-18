/*
Example
	https://github.com/golang-samples/websocket

Library
	https://code.google.com/p/go.net/websocket
	https://github.com/gorilla/websocket
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"code.google.com/p/go.net/websocket"
)

const Version = "0.2.2"

var index_jquery_tmpl = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8" />
<title>Sample of websocket with golang</title>
<script src="static/jquery-1.8.3.min.js"></script>
<script>
$(function() {
	var ws = new WebSocket("ws://{{.Host}}:{{.Port}}/{{.Mode}}");
	ws.onmessage = function(e) {
		console.log("Rx message:" + event.data);
	};

	var $ul = $('#msg-list');
	$('#sendBtn').click(function(){
		var data = $('#name').val();
			ws.send(data);
			console.log("Tx message:" + data);
			$('<li>').text(data).appendTo($ul);
		});
	});
</script>
</head>
<body>
	<input id="name" type="text" />
	<input type="button" id="sendBtn" value="send"></input>
	<ul id="msg-list"></ul>
</body>
</html>
`

type Config struct {
	Title string
	Host  string
	Port  string
	Mode  string
}

// command line options
var (
	fhost = flag.String("host", "localhost", "websocket host address")
	fport = flag.String("port", "8080", "websocket port")
	fmode = flag.String("mode", "echo", "websocket run mode")
	fserv = flag.Bool("d", false, "run as a server")
	fver  = flag.String("version", Version, "program version")
)

func init() {
	log.SetFlags(log.Lshortfile)

	flag.Parse()
}

func printCopyright() {
	fmt.Printf("WebSocket Echo Service, %s, (c)2015 by Stoney Kang, sikang99@gmail.com\n", Version)
}

func main() {
	printCopyright()

	// determine run type of program
	if *fserv == true {
		echoServer(*fhost, *fport, *fmode)
	} else {
		echoClient(*fhost, *fport, *fmode)
	}
}

func echoHello(ws *websocket.Conn) {
	smsg := []byte("Hello, World!")

	n, err := ws.Write(smsg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Tx: %s (%d)\n", smsg, n)

	var rmsg = make([]byte, 512)

	m, err := ws.Read(rmsg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Rx: %s (%d)\n", rmsg, m)
}

func echoInput(ws *websocket.Conn) {
	for {
		fmt.Print("> Enter text: ")
		reader := bufio.NewReader(os.Stdin)
		text, _, _ := reader.ReadLine()
		//fmt.Println(text)
		if len(text) == 0 {
			continue
		}

		if strings.Contains(string(text), ".quit") {
			fmt.Println("bye bye!\n")
			break
		}

		_, err := ws.Write([]byte(text))
		if err != nil {
			log.Fatal(err)
		}

		var rmsg = make([]byte, 512)

		m, err := ws.Read(rmsg)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s (%d)\n", string(rmsg[:m]), m)
	}
}

func echoClient(host, port, mode string) {
	fmt.Printf("> connect to %s:%s/%s\n", host, port, mode)

	// because websocket is a switching protocol from http at first
	origin := "http://" + host
	url := "ws://" + host + ":" + port + "/" + mode

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	echoHello(ws)
	echoInput(ws)
}

func procHandler(ws *websocket.Conn) {
	rmsg := make([]byte, 512)

	n, err := ws.Read(rmsg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Rx: %s (%d)\n", rmsg[:n], n)

	m, err := ws.Write(rmsg[:n])
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Tx: %s (%d)\n", rmsg[:m], m)
}

func echoHandler(ws *websocket.Conn) {
	log.Printf("%s\n", ws.Request().URL.Path)
	io.Copy(ws, ws)
}

var conf = Config{
	Title: "Go file upload and delete",
	Host:  *fhost,
	Port:  *fport,
	Mode:  *fmode,
}

func sendPage(w http.ResponseWriter, page string) error {
	//log.Printf("send %s\n", page)

	t, err := template.New("upload").Parse(page)
	if err != nil {
		log.Println(err)
		return err
	}

	return t.Execute(w, conf)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v\n", r.URL)

	if strings.Contains(r.URL.Path, "favicon.ico") {
		w.Header().Set("Content-Type", "image/icon")
		body, err := ioutil.ReadFile("static/favicon.ico")
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Fprint(w, string(body))
		return
	}

	sendPage(w, index_jquery_tmpl)
	return
}

func echoServer(host, port, mode string) {
	fmt.Printf("> access at http://%s:%s and ws://%s:%s/%s\n", host, port, host, port, mode)

	http.HandleFunc("/", indexHandler)
	http.Handle("/echo", websocket.Handler(echoHandler))
	http.Handle("/proc", websocket.Handler(procHandler))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Println(err)
	}
}
