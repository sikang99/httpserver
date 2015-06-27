//=========================================================================
// Author : Stoney Kang, sikang99@gmail.com, 2015
// Protocol for HTTP streaming
// - http://www.sanarias.com/blog/1214PlayingwithimagesinHTTPresponseingolang
// - http://stackoverflow.com/questions/30552447/how-to-set-which-ip-to-use-for-a-http-request
// - http://giantmachines.tumblr.com/post/52184842286/golang-http-client-with-timeouts
//=========================================================================

package mediaconf

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/net/websocket"

	"github.com/bradfitz/http2"

	pf "stoney/httpserver/src/protofile"
	ph "stoney/httpserver/src/protohttp"
	pt "stoney/httpserver/src/prototcp"

	sb "stoney/httpserver/src/streambase"
	sr "stoney/httpserver/src/streamring"
)

//---------------------------------------------------------------------------
// http monitor client
//---------------------------------------------------------------------------
func (sc *ServerConfig) StreamMonitor(url string) error {
	log.Printf("start %s for %s\n", ph.STR_HTTP_MONITOR, url)
	defer log.Printf("end %s for %s\n", ph.STR_HTTP_MONITOR, url)

	var err error
	var prestr string

	r := bufio.NewReader(os.Stdin)

	for {
		cmdstr, err := PromptReadLine("command> ", r)
		if err != nil {
			log.Println(err)
			break
		}

		if strings.EqualFold(cmdstr, "quit") {
			fmt.Println("Bye bye.")
			break
		} else if cmdstr == "." {
			// previous command
			cmdstr = prestr
		}

		// remember command
		prestr = cmdstr

		err = sc.ParseMonitorCommand(cmdstr, r)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	return err
}

//---------------------------------------------------------------------------
// prompt and read a line
//---------------------------------------------------------------------------
func PromptReadLine(prompt string, r *bufio.Reader) (string, error) {
	var err error
	var str string

	fmt.Print(prompt)

	line, _, err := r.ReadLine()
	if err != nil {
		log.Println(err)
		return str, err
	}

	str = strings.Replace(string(line), "\r", "", -1)

	return str, err
}

func PromptReadLineWithDefault(prompt string, def string, r *bufio.Reader) (string, error) {
	var err error
	var str string

	str, err = PromptReadLine(prompt, r)
	if err != nil {
		return str, err
	}

	if str == "" {
		str = def
	}

	return str, err
}

func PromptReadLineUntilInput(prompt string, r *bufio.Reader) (string, error) {
	var err error
	var str string

	for {
		str, err = PromptReadLine(prompt, r)
		if err != nil || str != "" {
			break
		}
	}

	return str, err
}

//---------------------------------------------------------------------------
// parse command
//---------------------------------------------------------------------------
func (sc *ServerConfig) ParseMonitorCommand(cmdstr string, r *bufio.Reader) error {
	var err error

	//fmt.Println(cmdstr)
	toks := strings.Fields(cmdstr)
	if len(toks) < 1 {
		return err
	}
	ntok := len(toks)
	//fmt.Println(toks)

	baseUrl, err := url.Parse("http://localhost:8080/command")
	if err != nil {
		log.Println(err)
		return err
	}

	params := url.Values{}
	params.Add("op", toks[0])

	var method string
	switch toks[0] {
	// GET ops
	case "show":
		if ntok < 2 {
			fmt.Printf("usage: show [config|dir|network|channel|ring|array]\n")
			return err
		}

		method = "GET"
		params.Add("obj", toks[1])

		switch toks[1] {
		case "config":
		case "network":
		case "channel":
		case "array":
		case "ring":
			id := "0"
			if ntok < 3 {
				id, err = PromptReadLineWithDefault("Ring id? [0]: ", id, r)
				if err != nil {
					return err
				}
			} else {
				id = toks[2]
			}
			params.Add("id", id)
		case "dir":
			path := "record"
			if ntok < 3 {
				path, err = PromptReadLineWithDefault("dir path? [record]: ", path, r)
				if err != nil {
					return err
				}
			} else {
				path = toks[2]
			}
			params.Add("path", path)
		default:
			fmt.Printf("I can't %s for %s\n", toks[0], toks[1])
			return err
		}

	// POST ops
	case "start":
		if ntok < 2 {
			fmt.Printf("usage: start [http_reader|tcp_caster/server|dir_reader|file_reader/writer] [params ...]\n")
			return err
		}

		method = "POST"
		params.Add("obj", toks[1])

		switch toks[1] {
		case "http_reader":
			url := "http://imoment:imoment@192.168.0.91/axis-cgi/mjpg/video.cgi"
			if ntok < 3 {
				url, err = PromptReadLineWithDefault("url to read? [axis]: ", url, r)
				if err != nil {
					return err
				}
			} else {
				url = toks[2]
			}
			params.Add("url", url)
		case "dir_reader":
			if ntok > 2 {
				params.Add("file", toks[2])
			} else {
				params.Add("file", "static/image/*.jpg")
			}
		case "file_reader":
			fallthrough
		case "file_writer":
			file := "record/output.mjpg"
			if ntok < 3 {
				file, err = PromptReadLineWithDefault("file to handle? [record/output.mjpg]: ", file, r)
				if err != nil {
					return err
				}
			} else {
				file = toks[2]
			}
			params.Add("file", file)
		case "tcp_caster":
			fallthrough
		case "tcp_server":
			port := "8087"
			if ntok < 3 {
				port, err = PromptReadLineWithDefault("port to handle? [8087]: ", port, r)
				if err != nil {
					return err
				}
			} else {
				port = toks[2]
			}
			params.Add("port", port)
			id := "0"
			if ntok < 4 {
				id, err = PromptReadLineWithDefault("Ring id? [0]: ", id, r)
				if err != nil {
					return err
				}
			} else {
				id = toks[3]
			}
			params.Add("id", id)
		default:
			fmt.Printf("I can't %s for %s\n", toks[0], toks[1])
			return err
		}

	case "stop":
		if ntok < 2 {
			fmt.Printf("usage: stop [ring|array]\n")
			return err
		}

		method = "POST"
		params.Add("obj", toks[1])

		switch toks[1] {
		case "ring":
			id := "0"
			if ntok < 3 {
				id, err = PromptReadLineWithDefault("Ring id? [0]: ", id, r)
				if err != nil {
					return err
				}
			}
			params.Add("id", id)
		case "array":
			yn, err := PromptReadLineWithDefault("Are you sure? [N/y]: ", "N", r)
			if yn != "y" {
				return err
			}
		default:
			fmt.Printf("I can't %s for %s\n", toks[0], toks[1])
			return err
		}

	case "help":
		if ntok < 2 {
			fmt.Printf("usage: help [show|start|stop]\n")
			return err
		}

		switch toks[1] {
		case "show":
			fmt.Printf("usage: show [config|dir|network|ring|array]\n")
		case "start":
			fmt.Printf("usage: start [http_reader|dir_reader|file_reader/writer|tcp_caster/server]\n")
		case "stop":
			fmt.Printf("usage: stop [ring|array]\n")
		default:
			fmt.Printf("I can't %s for %s\n", toks[0], toks[1])
		}
		return err

	case "test":
		fmt.Printf("for TEST\n")
		return err

	default:
		fmt.Printf("usage: [show|start|stop|help|quit]\n")
		return err
	}

	// make and send request
	baseUrl.RawQuery = params.Encode()

	var res *http.Response
	if method == "POST" {
		res, err = http.Post(fmt.Sprint(baseUrl), "text/plain", nil)
	} else {
		res, err = http.Get(fmt.Sprint(baseUrl))
	}
	if err != nil {
		log.Println(err)
		return err
	}

	body, err := ioutil.ReadAll(res.Body)
	fmt.Printf("%s\n", string(body))

	return err
}

//---------------------------------------------------------------------------
//	multipart reader entry, mainly from camera
//---------------------------------------------------------------------------
func (sc *ServerConfig) StreamReader(url string, ring *sr.StreamRing) error {
	log.Printf("start %s for %s\n", ph.STR_HTTP_READER, url)
	defer log.Printf("end %s for %s\n", ph.STR_HTTP_READER, url)

	var err error
	var res *http.Response

	// WHY: different behavior?
	if strings.Contains(url, "axis") {
		res, err = http.Get(url)
	} else {
		client := ph.NewClient()
		res, err = client.Get(url)
	}
	if err != nil {
		log.Println(sb.RedString(err))
		return err
	}
	defer res.Body.Close()

	boundary, err := ph.GetTypeBoundary(res.Header.Get("Content-Type"))
	if err != nil {
		log.Println(err)
		return err
	}
	ring.Boundary = boundary

	mr := multipart.NewReader(res.Body, ring.Boundary)

	err = ph.ReadMultipartToRing(mr, ring)

	return err
}

//---------------------------------------------------------------------------
// http caster client
//---------------------------------------------------------------------------
func (sc *ServerConfig) StreamCaster(url string) error {
	log.Printf("start %s for %s\n", ph.STR_HTTP_CASTER, url)
	defer log.Printf("end %s for %s\n", ph.STR_HTTP_CASTER, url)

	hp := ph.NewProtoHttp()
	client := ph.NewClient()
	return hp.RequestPost(client)

	// TODO: sending data
}

//---------------------------------------------------------------------------
// http player client
//---------------------------------------------------------------------------
func (sc *ServerConfig) StreamPlayer(url string, ring *sr.StreamRing) error {
	log.Printf("start %s for %s\n", ph.STR_HTTP_PLAYER, url)
	defer log.Printf("end %s for %s\n", ph.STR_HTTP_PLAYER, url)

	var err error
	var res *http.Response

	// WHY: different behavior?
	if strings.Contains(url, "axis") {
		res, err = http.Get(url)
	} else {
		client := ph.NewClient()
		res, err = client.Get(url)
	}
	if err != nil {
		log.Println(sb.RedString(err))
		return err
	}

	ring.Boundary, err = ph.GetTypeBoundary(res.Header.Get("Content-Type"))
	mr := multipart.NewReader(res.Body, ring.Boundary)

	err = ph.ReadMultipartToRing(mr, ring)

	return err
}

//---------------------------------------------------------------------------
// http server entry
//---------------------------------------------------------------------------
func (sc *ServerConfig) StreamServer(ring *sr.StreamRing) error {
	log.Printf("start %s\n", ph.STR_HTTP_SERVER)
	defer log.Printf("end %s\n", ph.STR_HTTP_SERVER)

	http.HandleFunc("/", sc.IndexHandler)
	http.HandleFunc("/hello", sc.HelloHandler)     // view
	http.HandleFunc("/media", sc.MediaHandler)     // on-demand
	http.HandleFunc("/stream", sc.StreamHandler)   // live
	http.HandleFunc("/search", sc.SearchHandler)   // server info
	http.HandleFunc("/command", sc.CommandHandler) // server control & monitor

	http.Handle("/websocket", websocket.Handler(sc.WebsocketHandler))

	// CAUTION: don't use /static not /static/ as the prefix
	http.Handle("/static/", http.StripPrefix("/static/", FileServer("./static")))

	//var wg sync.WaitGroup
	wg := sync.WaitGroup{}

	wg.Add(1)
	// HTTP server
	go sc.ServeHttp(&wg)

	wg.Add(1)
	// HTTPS server
	go sc.ServeHttps(&wg)

	wg.Add(1)
	// HTTP2 server
	go sc.ServeHttp2(&wg)

	/*
		wg.Add(1)
		// WS server
		go sc.ServeWs(&wg)

		wg.Add(1)
		// WSS server
		go sc.ServeWss(&wg)
	*/

	wg.Wait()

	return nil
}

//---------------------------------------------------------------------------
// index file handler
//---------------------------------------------------------------------------
func (sc *ServerConfig) IndexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("handle /index %s to %s\n", r.URL.Path, r.Host)
	defer r.Body.Close()

	if strings.Contains(r.URL.Path, "favicon.ico") {
		http.ServeFile(w, r, "static/favicon.ico")
		return
	}

	ph.WriteTemplatePage(w, index_tmpl, sc)
}

//---------------------------------------------------------------------------
// hello file handler (default: hello.html)
//---------------------------------------------------------------------------
func (sc *ServerConfig) HelloHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("handle /hello %s to %s\n", r.URL.Path, r.Host)
	defer r.Body.Close()

	hello_page := "static/hello.html"

	host := strings.Split(r.Host, ":")
	sc.Port = host[1]

	if r.TLS != nil {
		sc.Addr = "https://localhost"
	}

	_, err := os.Stat(hello_page)
	if err != nil {
		ph.WriteTemplatePage(w, hello_tmpl, sc)
		log.Printf("hello serve %s\n", "hello_tmpl")
	} else {
		http.ServeFile(w, r, hello_page)
		log.Printf("hello serve %s\n", hello_page)
	}
}

//---------------------------------------------------------------------------
// handle /media to provide on-demand media file request
//---------------------------------------------------------------------------
func (sc *ServerConfig) MediaHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("handle /media %s for %s to %s\n", r.Method, r.RequestURI, r.Host)
	defer r.Body.Close()

	_, err := os.Stat(r.URL.Path[1:])
	if err != nil {
		ph.WriteResponseMessage(w, 404, r.URL.Path+" is Not Found")
	} else {
		ph.WriteResponseMessage(w, 200, r.URL.Path)
	}
}

//---------------------------------------------------------------------------
// websocket handler
//---------------------------------------------------------------------------
func (sc *ServerConfig) WebsocketHandler(ws *websocket.Conn) {
	log.Printf("handle /websocket %s\n", ws.RemoteAddr())
	defer ws.Close()

	data := make([]byte, sb.MBYTE)
	for {
		n, err := ws.Read(data)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(n)
	}

	err := websocket.Message.Send(ws, "Not yet implemented")
	if err != nil {
		log.Println(err)
		return
	}
}

//---------------------------------------------------------------------------
// handle /search access
//---------------------------------------------------------------------------
func (sc *ServerConfig) SearchHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("handle /search %s for %s to %s\n", r.Method, r.RequestURI, r.Host)
	defer r.Body.Close()

	var err error

	err = ph.WriteResponseMessage(w, 200, "/search: Not yet implemented")
	if err != nil {
		log.Println(err)
	}

	return
}

//---------------------------------------------------------------------------
// handle /command access
//---------------------------------------------------------------------------
func (sc *ServerConfig) CommandHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("handle /command %s for %s to %s\n", r.Method, r.RequestURI, r.Host)
	defer r.Body.Close()

	var err error
	var str string

	query := r.URL.Query()
	op := query.Get("op")
	obj := query.Get("obj")

	ring := sc.Array[0]

	switch r.Method {
	// monitor part
	case "GET":
		switch op {
		case "show":
			switch obj {
			case "dir":
				path := query.Get("path")
				res, _ := exec.Command("ls", "-al", "./"+path).CombinedOutput()
				str = string(res)
			case "network":
				str = sb.ShowNetInterfaces()
			case "config":
				str = fmt.Sprint(sc)
			case "ring":
				id := query.Get("id")
				i, err := strconv.Atoi(id)
				if err == nil && i < len(sc.Array) {
					ring = sc.Array[i]
				}
				str = fmt.Sprint(ring)
			case "array":
				for i := range sc.Array {
					str += fmt.Sprintf("[%d] %s\n", i, sc.Array[i].BaseString())
				}
			default:
				str = "what obj? [config|network|ring|array]"
			}
		default:
			str = "what op? [show]"
		}

	// control part
	case "POST":
		switch op {
		case "start":
			switch obj {
			case "http_reader":
				url := query.Get("url")
				go sc.StreamReader(url, ring)
				str = "order to start http_reader " + url
			case "dir_reader":
				file := query.Get("file")
				fp := pf.NewProtoFile(file)
				//go fp.DirReader(ring, true)
				go fp.DirReader(sc.Array[1], true)
				str = "order to start dir_reader " + file
			case "file_reader":
				file := query.Get("file")
				fp := pf.NewProtoFile(file)
				go fp.StreamReader(ring)
				str = "order to start file_reader " + file
			case "file_writer":
				file := query.Get("file")
				fp := pf.NewProtoFile(file)
				go fp.StreamWriter(ring)
				str = "order to start file_writer " + file
			case "tcp_server":
				port := query.Get("port")
				id := query.Get("id")
				i, _ := strconv.Atoi(id)
				i = i % len(sc.Array)
				tp := pt.NewProtoTcp("localhost", port, "T-Rx")
				go tp.StreamServer(sc.Array[i])
				str = fmt.Sprintf("order to start tcp_server (%s, %s)", port, id)
			case "tcp_caster":
				port := query.Get("port")
				tp := pt.NewProtoTcp("localhost", port, "T-Rx")
				go tp.StreamCaster()
				str = "order to start tcp_caster " + port
			default:
				str = "what obj to start? [http_reader|dir_reader|file_reader/writer|tcp_caster/server]"
			}
		case "stop":
			switch obj {
			case "ring":
				id := query.Get("id")
				i, err := strconv.Atoi(id)
				if err == nil && i < len(sc.Array) {
					ring = sc.Array[i]
				} else {
					str = "invalid ring number: " + id
					break
				}
				err = ring.SetStatusIdle()
				if err != nil {
					str = fmt.Sprint(err)
				} else {
					str = "set to stop the ring: " + id
				}
			case "array":
				for i := range sc.Array {
					sc.Array[i].SetStatusIdle()
				}
				str = "stopped all rings"
			default:
				str = "what obj to stop? [ring]"
			}
		default:
			str = "what op? [start|stop]"
		}

	default:
		str = "what's this?"
	}

	err = ph.WriteResponseMessage(w, http.StatusOK, str)
	if err != nil {
		log.Println(err)
	}
}

//---------------------------------------------------------------------------
// handle /stream access
//---------------------------------------------------------------------------
func (sc *ServerConfig) StreamHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("handle /stream %s for %s to %s\n", r.Method, r.RequestURI, r.Host)
	defer r.Body.Close()

	var err error
	/*
		query := r.URL.Query()
		trk := query.Get("track")
	*/
	ring := sc.Array[0]

	switch r.Method {
	case "POST": // for Caster
		ring.Boundary, err = ph.GetTypeBoundary(r.Header.Get("Content-Type"))

		err = ph.ResponsePost(w, ring.Boundary)
		if err != nil {
			log.Println(err)
			break
		}

		mr := multipart.NewReader(r.Body, ring.Boundary)

		err = ph.ReadMultipartToRing(mr, ring)
		if err != nil {
			log.Println(err)
			break
		}

	case "GET": // for Player
		err = ph.ResponseGet(w, ring.Boundary)
		if err != nil {
			log.Println(err)
			break
		}

		err = ph.WriteRingInMultipart(w, ring)
		if err != nil {
			log.Println(err)
			break
		}

	default:
		log.Println("Unknown request method: ", r.Method)
	}

	return
}

//---------------------------------------------------------------------------
// serve http access
//---------------------------------------------------------------------------
func (sc *ServerConfig) ServeHttp(wg *sync.WaitGroup) {
	log.Println("start HTTP server at http://" + sc.Host + ":" + sc.Port)
	defer log.Println("end HTTP server at http://" + sc.Host + ":" + sc.Port)

	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + sc.Port,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

//---------------------------------------------------------------------------
// serve https tls access
//---------------------------------------------------------------------------
func (sc *ServerConfig) ServeHttps(wg *sync.WaitGroup) {
	log.Println("start HTTPS server at https://" + sc.Host + ":" + sc.PortS)
	defer log.Println("end HTTPS server at https://" + sc.Host + ":" + sc.PortS)

	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + sc.PortS,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServeTLS("sec/cert.pem", "sec/key.pem"))
}

//---------------------------------------------------------------------------
// serve http2 tls access
//---------------------------------------------------------------------------
func (sc *ServerConfig) ServeHttp2(wg *sync.WaitGroup) {
	log.Println("start HTTP2 server at https://" + sc.Host + ":" + sc.Port2)
	defer log.Println("end HTTP2 server at https://" + sc.Host + ":" + sc.Port2)

	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + sc.Port2,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	http2.ConfigureServer(srv, &http2.Server{})
	log.Fatal(srv.ListenAndServeTLS("sec/cert.pem", "sec/key.pem"))
}

//---------------------------------------------------------------------------
// serve ws access
// http://www.ajanicij.info/content/websocket-tutorial-go
//---------------------------------------------------------------------------
func (sc *ServerConfig) ServeWs(wg *sync.WaitGroup) {
	log.Println("start WS server at https://" + sc.Host + ":" + sc.Port)
	defer log.Println("end WS server at https://" + sc.Host + ":" + sc.Port)

	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + sc.Port,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

//---------------------------------------------------------------------------
// serve wss access
//---------------------------------------------------------------------------
func (sc *ServerConfig) ServeWss(wg *sync.WaitGroup) {
	log.Println("start WSS server at https://" + sc.Host + ":" + sc.PortS)
	defer log.Println("end WSS server at https://" + sc.Host + ":" + sc.PortS)

	defer wg.Done()

	srv := &http.Server{
		Addr: ":" + sc.PortS,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServeTLS("sec/cert.pem", "sec/key.pem"))
}

//---------------------------------------------------------------------------
// file server for path
//---------------------------------------------------------------------------
func FileServer(path string) http.Handler {
	log.Println("start File server for " + path)
	defer log.Println("end File server for " + path)

	return http.FileServer(http.Dir(path))
}

// ---------------------------------E-----N-----D--------------------------------
