# Makefile for http server supporting http, https, http2

.PHONY:	edit build run rebuild install clean make

all: usage

PROGRAM=httpserver
EDITOR=vim

edit e:
	$(EDITOR) src/server/$(PROGRAM).go

ew:
	$(EDITOR) src/base/wrapper.go

build b:
	go build -o $(PROGRAM) src/server/$(PROGRAM).go
	@ls -alF --color=auto

run r:
	@echo ""
	@echo "make (run) [rh|rh2|rc|rc2|rs|rm]"
	@echo "    rh  : run server and access to it with http"
	@echo "    rh2 : run server and access to it with http2"
	@echo "    rs  : run server"
	@echo "    rc  : run client for http"
	@echo "    rc2 : run client for http2"
	@echo "    rm  : run monitor"
	@echo ""

rh:
	@chromium-browser http://localhost:8080/hello
	./$(PROGRAM) -d -port=8080 -ports=8081 -port2=8082

rh2:
	@chromium-browser --allow-running-insecure-content https://localhost:8082/static
	./$(PROGRAM) -d -port=8080 -ports=8081 -port2=8082

rc:
	./$(PROGRAM) -url http://localhost:8080/hello
	./$(PROGRAM) -url http://localhost:8080/media/gopher.jpeg
	./$(PROGRAM) -url http://localhost:8080/static/image/gopher.jpg
	./$(PROGRAM) -url http://localhost:8080/index.html

rc2:
	./$(PROGRAM) -url https://localhost:8082/hello
	./$(PROGRAM) -url https://localhost:8082/media/gopher.jpeg
	./$(PROGRAM) -url https://localhost:8082/static/image/gopher.jpg
	./$(PROGRAM) -url https://localhost:8082/index.html

rm:
	./$(PROGRAM) -m -port=8080 -ports=8081 -port2=8082

rs:
	./$(PROGRAM) -d -port=8080 -ports=8081 -port2=8082

rt:
	curl -I http://localhost:8080/stream

rebuild:
	rm -f ./$(PROGRAM)
	go build $(PROGRAM).go
	@ls -alF --color=auto

install i:
	go install

kill k:
	killall httpserver

clean:
	rm -f ./$(PROGRAM)

# ---------------------------------------------------------------------------
git g:
	@echo ""
	@echo "make (git) [gv|gh|gd|gp|gs]"
	@echo "    gv  : git view"
	@echo "    gp  : git push"
	@echo "    gs  : git status"
	@echo ""

gv:
	LANG=C chromium-browser https://github.com/sikang99/$(PROGRAM)

gh:
	ssh -T git@github.com

gd:
	git push -u https://sikang99@github.com/sikang99/$(PROGRAM) master

gp:
	git init
	git add * .gitignore
	git commit -m "define mjpeg struct"
	git push -u https://sikang99@github.com/sikang99/$(PROGRAM) master

gs:
	git status
	git log --oneline -5

gencert:
	openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 9999 -nodes

# ---------------------------------------------------------------------------
readme md:
	$(EDITOR) README.md

make m:
	$(EDITOR) Makefile

usage:
	@echo ""
	@echo "Makefile for '$(PROGRAM)', by Stoney Kang, 2015/05/15"
	@echo ""
	@echo "usage: make [edit|readme|build|run|test|rebuild|clean|git]"
	@echo "    edit(e)    : edit source"
	@echo "    build(b)   : compile source"
	@echo "    run(r)     : execute $(PROGRAM)"
	@echo "    install(i) : install $(PROGRAM) to $(GOPATH)/bin"
	@echo "    git(g)     : git $(PROGRAM) to github.com"
	@echo "    gencert    : make certificates for https"
	@echo ""
