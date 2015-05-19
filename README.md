# Static http/https server

Simple http/https test server for some services and tatic files.

## Install

    $ go get github.com/sikang99/httpserver

## Usage

To serve the current directory on port 8080:

    $ httpserver

To use a different port specify with the `-port` flag:

    $ httpserver -port=5000

To serve a different directory use the `-root` flag:

    $ httpserver -root=./public

## Options

* `-port` Defines the TCP port to listen on. (Defaults to 8080).
* `-root` Defines the directory to serve. (Defaults to the current directory).


## Design

Internal service structure

	http  --> server(8001)     + --> file server
	https --> server(8001) --> + --> streaming server
	http2 --> server(8002)     + --> monitor server




## References

### V4L2

- [v4l2grab - Grabbing JPEGs from V4L2 devices](http://www.twam.info/linux/v4l2grab-grabbing-jpegs-from-v4l2-devices)

### Streaming

- [mattn/go-mjpeg](https://github.com/mattn/go-mjpeg) - 
- [saljam/mjpeg](https://github.com/saljam/mjpeg) - MJPEG streaming for Go
- [saljam/webcam](https://github.com/saljam/webcam) - WebRTC based one-way camera streaming
- [hlubek/webcamproxy](https://github.com/hlubek/webcamproxy) - Proxy for a webcam stream over websockets in Go (golang)
- [BlueDragonX/go-proxy-example](https://github.com/BlueDragonX/go-proxy-example) - An example TCP proxy in Go

### MIME/Multipart

- [technoweenie/multipartstreamer](https://github.com/technoweenie/multipartstreamer) - Encode large files in MIME multipart format without reading the entire content into memory
- [tdewolff/minify](https://github.com/tdewolff/minify) - Go minifiers for web formats
- [Looking for good net/http package tutorials, blog posts, etc (self.golang)](http://www.reddit.com/r/golang/comments/364z8t/looking_for_good_nethttp_package_tutorials_blog/)
- [HTTP-POST file multipart programming in Go language](http://stackoverflow.com/questions/7223616/http-post-file-multipart-programming-in-go-language)

### HTTP

- [aljemala/tls-client](https://gist.github.com/michaljemala/d6f4e01c4834bf47a9c4) - SSL Client Authentication Golang sample
- [Serving server generated PNGs over HTTP in golang](http://41j.com/blog/2015/03/serving-server-generated-pngs-over-http-in-golang/)
- [Urban4M/go-workgroup](https://github.com/Urban4M/go-workgroup) - go-workgroup - wraps sync.WaitGroup
- [Gorilla Websockets, golang simple websockets example](http://41j.com/blog/2014/12/gorilla-websockets-golang-simple-websockets-example/)
- [The http.HandlerFunc wrapper technique in #golang](https://medium.com/@matryer/the-http-handlerfunc-wrapper-technique-in-golang-c60bf76e6124)
- [Basic Encryption in Golang](http://golangcast.tv/articles/basic-encryption-in-golang)
- [jameycribbs/ivy](https://github.com/jameycribbs/ivy) - A simple, file-based Database Management System (DBMS) for Go
- [goware/httpmock](https://github.com/goware/httpmock) - HTTP mocking in Go made easy
- [Using Object-Oriented Web Servers in Go](http://blog.codeship.com/using-object-oriented-web-servers-go/)
- [sparks/rter](https://github.com/sparks/rter) - rtER: Real-Time Emergency Response
- [Golang. What to use? http.ServeFile(..) or http.FileServer(..)?](http://stackoverflow.com/questions/28793619/golang-what-to-use-http-servefile-or-http-fileserver)
- [HTTPS and Go](https://www.kaihag.com/https-and-go/)
- [Static http file server in Go](https://www.chrismytton.uk/2013/07/17/golang-static-http-file-server/)
- Forked from [chrismytton/httpserver](https://github.com/chrismytton/httpserver).
- Inspired by [nodeapps/http-server](https://github.com/nodeapps/http-server).


## License

MIT

