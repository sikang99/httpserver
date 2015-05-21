/*
	multipart stream generator
*/
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
)

type proxy struct {
	io.Writer
}

func oldmain() {
	outer := multipart.NewWriter(os.Stdout)

	buf := new(bytes.Buffer)
	delay := proxy{buf}

	inner := multipart.NewWriter(delay)

	for i := 0; i < 3; i++ {
		part, _ := outer.CreatePart(textproto.MIMEHeader{
			"Content-Type": {"multipart/mixed; boundary=" + inner.Boundary()},
		})

		buf.WriteTo(part)
		delay.Writer = part
	}
}

func main() {
	/*
		files := []string{"static/01.txt", "static/02.txt", "static/03.txt"}
		files = append(files, "static/ubuntu.jpg")
		files = append(files, "static/video.mjpg")
		files = append(files, "static/favicon.ico")
		files = append(files, "static/tiger.svg")
	*/

	files, err := filepath.Glob("static/*")
	if err != nil {
		log.Println(err)
	}
	if files == nil {
		log.Println("no matched files")
		return
	}
	fmt.Println(files)

	mw := multipart.NewWriter(os.Stdout)
	mw.SetBoundary("myboundary")

	b := new(bytes.Buffer)

	for i := range files {
		fmt.Println(files[i], len(files[i]))

		fdata, err := ioutil.ReadFile(files[i])
		if err != nil {
			log.Println(err)
		}
		fsize := len(fdata)

		ctype := mime.TypeByExtension(filepath.Ext(files[i]))
		if ctype == "" {
			ctype = http.DetectContentType(fdata)
		}

		part, err := mw.CreatePart(textproto.MIMEHeader{
			"Content-Type":   {ctype},
			"Content-Length": {strconv.Itoa(fsize)},
		})
		if err != nil {
			log.Println(err)
		}

		//b.Write(fdata)  // prepare data in the buffer
		b.WriteTo(part) // output the part in multipart format
	}
}
