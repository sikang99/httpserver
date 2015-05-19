package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/textproto"
	"os"
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
	files := []string{"static/01.txt", "static/02.txt", "static/03.txt"}
	files = append(files, "static/04.txt")

	mr := multipart.NewWriter(os.Stdout)
	mr.SetBoundary("myboundary")

	b := new(bytes.Buffer)

	for i := range files {
		fdata, _ := ioutil.ReadFile(files[i])
		fsize := len(fdata)

		part, _ := mr.CreatePart(textproto.MIMEHeader{
			"Content-Type":   {"text/plain"},
			"Content-Length": {strconv.Itoa(fsize)},
		})

		b.Write(fdata)
		b.WriteTo(part)
	}
}
