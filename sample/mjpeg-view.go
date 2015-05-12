/*
	https://gist.githubusercontent.com/gregworley/1294450/raw/9ba9fd49ed800b76c8674ecb0117b8479380aad1/gistfile1.go

	X11 viewer for a MJPEG stream, such as the one obtained from the
	Android app https://market.android.com/details?id=com.pas.webcam
*/
package main

import (
	//"exp/gui/x11"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

var (
	//stream = flag.String("stream", "http://10.0.0.212:8080/videofeed", "MJPEG stream")
	//stream = flag.String("stream", "http://imoment:imoment@192.168.0.91/axis-cgi/jpg/image.cgi", "JPEG Still Image")
	stream = flag.String("stream", "http://imoment:imoment@192.168.0.91/axis-cgi/mjpg/video.cgi", "JPEG Still Image")
	flush  = flag.Bool("flush", true, "flush x11 after each frame")
)

func init() {
	log.SetOutput(os.Stdout)
	//log.SetPrefix("TRACE: ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	flag.Parse()
	res, err := http.Get(*stream)
	if err != nil {
		log.Fatalf("GET of %q: %v", *stream, err)
	}
	log.Printf("Content: %v", res.Header.Get("Content-Type"))

	// Content-Type: multipart/x-mixed-replace; boundary=myboundary
	_, params, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	fmt.Printf("%v %v\n", params, res.Header.Get("Content-Type"))
	if err != nil {
		log.Fatalf("ParseMediaType: %v", err)
	}
	//params := res.Header["Content-Type"]
	boundary := params["boundary"]
	if !strings.HasPrefix(boundary, "--") {
		log.Printf("expected boundary to start with --, got %q", boundary)
	}
	r := multipart.NewReader(res.Body, boundary)
	framec := make(chan image.Image, 10)
	go decodeFrames(framec, r)

	/*
			win, err := x11.NewWindow()
			if err != nil {
				log.Fatalf("NewWindow: %v", err)
			}
		eventc := win.EventChan()
	*/
	for {
		select {
		/*
			case evt, ok := <-eventc:
				if !ok {
					return
				}
				log.Printf("Event: %#v", evt)
		*/
		case im := <-framec:
			log.Printf("frame %v %q\n", im.Bounds(), im.ColorModel())
			/*
				sim := win.Screen()
				draw.Draw(sim, im.Bounds(), im, sim.Bounds().Min, draw.Over)
				if *flush {
					win.FlushImage()
				}
			*/
		}
	}
}

func decodeFrames(c chan image.Image, r *multipart.Reader) {
	for {
		p, err := r.NextPart()
		if err != nil {
			log.Fatalf("NextPart: %v", err)
		}
		im, err := jpeg.Decode(p)
		if err != nil {
			log.Fatalf("jpeg Decode: %v", err)
		}
		c <- im
	}
}
