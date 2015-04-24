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

## References

- [Basic Encryption in Golang](http://golangcast.tv/articles/basic-encryption-in-golang)
- [BlueDragonX/go-proxy-example](https://github.com/BlueDragonX/go-proxy-example) - An example TCP proxy in Go
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

