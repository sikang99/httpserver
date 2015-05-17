package base

type Mjpeg struct {
	HeaderType string
	MediaType  string
	Desc       string
}

func MjpegNew() (m *Mjpeg) {
	return &Mjpeg{
		HeaderType: "multipart/x-mixed-replace",
		MediaType:  "image/jpeg",
		Desc:       "MJPEG Stream",
	}
}

func (m Mjpeg) Open() {

}

func (m Mjpeg) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (m Mjpeg) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (m Mjpeg) Close() {

}
