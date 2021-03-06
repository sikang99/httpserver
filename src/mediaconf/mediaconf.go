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
	"github.com/fatih/color"

	pf "stoney/httpserver/src/protofile"
	ph "stoney/httpserver/src/protohttp"
	pt "stoney/httpserver/src/prototcp"
	pw "stoney/httpserver/src/protows"

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
		}

		if cmdstr == "" {
			continue
		}

		// for just enter(return)
		if cmdstr == "." {
			// exec previous command
			cmdstr = prestr
		} else {
			// remember current command
			prestr = cmdstr
		}

		err = ParseMonitorCommand(cmdstr, r)
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

	prompt = fmt.Sprintf("%s [%s]: ", prompt, def)
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
func ParseMonitorCommand(cmdstr string, r *bufio.Reader) error {
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
			fmt.Printf("usage: show [config|dir|network|channel|ring|array|actor]\n")
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
				id, err = PromptReadLineWithDefault("\tring id", id, r)
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
				path, err = PromptReadLineWithDefault("\tdir path", path, r)
				if err != nil {
					return err
				}
			} else {
				path = toks[2]
			}
			params.Add("path", path)
		case "actor":
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
				url, err = PromptReadLineWithDefault("\turl to read", url, r)
				if err != nil {
					return err
				}
			} else {
				url = toks[2]
			}
			if url == "q" { // command cancel
				return err
			}
			params.Add("url", url)
			id := "0"
			if ntok < 4 {
				id, err = PromptReadLineWithDefault("\tring id", id, r)
				if err != nil {
					return err
				}
			} else {
				id = toks[3]
			}
			params.Add("id", id)
		case "dir_reader":
			file := "static/image/*.jpg"
			if ntok < 3 {
				file, err = PromptReadLineWithDefault("\tpatten to read", file, r)
				if err != nil {
					return err
				}
			} else {
				file = toks[2]
			}
			if file == "q" { // command cancel
				return err
			}
			params.Add("file", file)
			id := "0"
			if ntok < 4 {
				id, err = PromptReadLineWithDefault("\tring id", id, r)
				if err != nil {
					return err
				}
			} else {
				id = toks[3]
			}
			params.Add("id", id)
		case "file_reader":
			fallthrough
		case "file_writer":
			file := "record/output.mjpg"
			if ntok < 3 {
				file, err = PromptReadLineWithDefault("\tfile to handle", file, r)
				if err != nil {
					return err
				}
			} else {
				file = toks[2]
			}
			if file == "q" { // command cancel
				return err
			}
			params.Add("file", file)
			id := "0"
			if ntok < 4 {
				id, err = PromptReadLineWithDefault("\tring id", id, r)
				if err != nil {
					return err
				}
			} else {
				id = toks[3]
			}
			params.Add("id", id)
		case "tcp_caster":
			fallthrough
		case "tcp_server":
			port := "8087"
			if ntok < 3 {
				port, err = PromptReadLineWithDefault("\tport to handle", port, r)
				if err != nil {
					return err
				}
			} else {
				port = toks[2]
			}
			if port == "q" { // command cancel
				return err
			}
			params.Add("port", port)
			id := "0"
			if ntok < 4 {
				id, err = PromptReadLineWithDefault("\tring id", id, r)
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
			fmt.Printf("usage: stop [ring|array|actor]\n")
			return err
		}

		method = "POST"
		params.Add("obj", toks[1])

		switch toks[1] {
		case "actor":
			id := "0"
			if ntok < 3 {
				id, err = PromptReadLineWithDefault("\tactor id", id, r)
				if err != nil {
					return err
				}
			} else {
				id = toks[2]
			}
			params.Add("id", id)
		default:
			fmt.Printf("I can't %s for %s\n", toks[0], toks[1])
			return err
		}

	case "close":
		if ntok < 2 {
			fmt.Printf("usage: close [ring|array]\n")
			return err
		}

		method = "POST"
		params.Add("obj", toks[1])

		switch toks[1] {
		case "ring":
			id := "0"
			if ntok < 3 {
				id, err = PromptReadLineWithDefault("\tring id", id, r)
				if err != nil {
					return err
				}
			} else {
				id = toks[2]
			}
			params.Add("id", id)
		case "array":
			yn, err := PromptReadLineWithDefault("\tAre you sure?", "N", r)
			if yn != "y" {
				return err
			}
		default:
			fmt.Printf("I can't %s for %s\n", toks[0], toks[1])
			return err
		}

	case "help":
		if ntok < 2 {
			fmt.Printf("usage: help [show|start|stop|close]\n")
			return err
		}

		switch toks[1] {
		case "show":
			fmt.Printf("usage: show [config|dir|network|ring|array]\n")
		case "close":
			fmt.Printf("usage: close [ring|array]\n")
		case "start":
			fmt.Printf("usage: start [http_reader|dir_reader|file_reader/writer|tcp_caster/server]\n")
		case "stop":
			fmt.Printf("usage: stop [actor] [id]\n")
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
	fmt.Printf("%s\n", color.BlueString(string(body)))

	return err
}

//---------------------------------------------------------------------------
//	multipart reader entry, mainly from camera
//---------------------------------------------------------------------------
func (sc *ServerConfig) StreamReader(ring *sr.StreamRing, url string) error {
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
func (sc *ServerConfig) StreamPlayer(ring *sr.StreamRing, url string) error {
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

	var err error

	wp := pw.NewProtoWs()

	err = wp.HandleRequest(ws, sc.Array[0])
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
					str = fmt.Sprintf("[%d] %s", i, ring)
				} else {
					str = fmt.Sprintf("error> id(%s)", id)
				}
			case "array":
				for i := range sc.Array {
					str += fmt.Sprintf("[%d] %s\n", i, sc.Array[i].BaseString())
				}
			case "actor":
				for key, actor := range sc.Actors {
					if actor.Status == sb.STATUS_IDLE {
						delete(sc.Actors, key)
					}
					str += fmt.Sprintf("%s\n", actor)
				}
			default:
				str = "what obj? [config|network|ring|array|actor]"
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
				id := query.Get("id")
				i, err := strconv.Atoi(id)
				if err == nil && i < len(sc.Array) {
					np := ph.NewProtoHttpWithUrl(url)
					sc.Actors[np.Base.Id] = np.Base
					go sc.StreamReader(sc.Array[i], url)
					str = fmt.Sprintf("order to start %s (%s -> %s)", obj, url, id)
				} else {
					str = fmt.Sprintf("error: %s (%s -> %s)", obj, url, id)
				}
			case "dir_reader":
				file := query.Get("file")
				id := query.Get("id")
				i, err := strconv.Atoi(id)
				if err == nil && i < len(sc.Array) {
					np := pf.NewProtoFile(file)
					sc.Actors[np.Base.Id] = np.Base
					go np.DirReader(sc.Array[i], true)
					str = fmt.Sprintf("order to start %s (%s, %s)", obj, file, id)
				} else {
					str = fmt.Sprintf("error: %s (%s -> %s)", obj, file, id)
				}
			case "file_reader":
				file := query.Get("file")
				id := query.Get("id")
				i, err := strconv.Atoi(id)
				if err == nil && i < len(sc.Array) {
					np := pf.NewProtoFile(file)
					sc.Actors[np.Base.Id] = np.Base
					go np.StreamReader(sc.Array[i])
					str = fmt.Sprintf("order to start %s (%s, %s)", obj, file, id)
				} else {
					str = fmt.Sprintf("error: %s (%s -> %s)", obj, file, id)
				}
			case "file_writer":
				id := query.Get("id")
				file := query.Get("file")
				i, err := strconv.Atoi(id)
				if err == nil && i < len(sc.Array) {
					np := pf.NewProtoFile(file)
					sc.Actors[np.Base.Id] = np.Base
					go np.StreamWriter(sc.Array[i])
					str = fmt.Sprintf("order to start %s (%s, %s)", obj, file, id)
				} else {
					str = fmt.Sprintf("error: %s (%s -> %s)", obj, file, id)
				}
			case "tcp_server":
				port := query.Get("port")
				id := query.Get("id")
				i, err := strconv.Atoi(id)
				if err == nil && i < len(sc.Array) {
					np := pt.NewProtoTcp("localhost", port, "T-Rx")
					sc.Actors[np.Base.Id] = np.Base
					go np.StreamServer(sc.Array[i])
					str = fmt.Sprintf("order to start %s (%s, %s)", obj, port, id)
				} else {
					str = fmt.Sprintf("error: %s (%s -> %s)", obj, port, id)
				}
			case "tcp_caster":
				port := query.Get("port")
				id := query.Get("id")
				i, err := strconv.Atoi(id)
				if err == nil && i < len(sc.Array) {
					np := pt.NewProtoTcp("localhost", port, "T-Rx")
					sc.Actors[np.Base.Id] = np.Base
					go np.StreamCaster()
					str = fmt.Sprintf("order to start %s (%s, %s)", obj, port, id)
				} else {
					str = fmt.Sprintf("error: %s (%s -> %s)", obj, port, id)
				}
			default:
				str = "what obj to start? [http_reader|dir_reader|file_reader/writer|tcp_caster/server]"
			}

		case "stop":
			switch obj {
			case "actor":
				id := query.Get("id")
				actor := sc.Actors[id]
				if actor != nil {
					actor.SetStatusClose()
					str = fmt.Sprintf("%s %s is closed", obj, id)
				} else {
					str = fmt.Sprintf("%s %s not exist", obj, id)
				}
			default:
				str = "what obj to stop? [ring|array]"
			}

		case "close":
			switch obj {
			case "ring":
				id := query.Get("id")
				i, err := strconv.Atoi(id)
				if err == nil && i < len(sc.Array) {
					ring = sc.Array[i]
				} else {
					str = "error: invalid ring number: " + id
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
				str = "closed all rings (array)"
			default:
				str = "what obj to close? [ring|array]"
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
