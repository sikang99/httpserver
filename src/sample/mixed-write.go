package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
)

type proxy struct {
	io.Writer
}

func main() {
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
