# Static http server

Ridiculously simple http/https server for static files.

## Install

    $ go get github.com/sikang99/httpserver

## Usage

To serve the current directory on port 8080:

    $ httpserver

To use a different port specify with the `-port` flag:

    $ httpserver -port 5000

To serve a different directory use the `-root` flag:

    $ httpserver -root public

## Options

* `-port` Defines the TCP port to listen on. (Defaults to 8080).
* `-root` Defines the directory to serve. (Defaults to the current directory).

## References

- [HTTPS and Go](https://www.kaihag.com/https-and-go/)
- [Static http file server in Go](https://www.chrismytton.uk/2013/07/17/golang-static-http-file-server/)
- Forked from [chrismytton/httpserver](https://github.com/chrismytton/httpserver).
- Inspired by [nodeapps/http-server](https://github.com/nodeapps/http-server).

## License

MIT

