# Makefile for http server supporting http, https, http2

.PHONY:	edit build run rebuild install clean make

all: usage

PROGRAM=httpserver
EDITOR=vim

edit e:
	$(EDITOR) src/server/$(PROGRAM).go

ew:
	$(EDITOR) src/base/wrapper.go

em:
	$(EDITOR) src/base/mjpeg.go


build b:
	go build -o $(PROGRAM) src/server/$(PROGRAM).go
	@ls -alF --color=auto

run r:
	@echo ""
	@echo "make (run) [rh|rh2|rc|rc2|rs|rm]"
	@echo "    rh  : run server and access to it with http"
	@echo "    rh2 : run server and access to it with http2"
	@echo "    rs  : run server"
	@echo "    rp  : run player for http"
	@echo "    rp2 : run player for http2"
	@echo "    rc : run caster for http"
	@echo "    rm  : run monitor"
	@echo ""

rh:
	@chromium-browser http://localhost:8080/hello
	./$(PROGRAM) -m=http_server -port=8080 -ports=8081 -port2=8082

rhs:
	@chromium-browser --allow-running-insecure-content https://localhost:8081/hello
	./$(PROGRAM) -m=http_server -port=8080 -ports=8081 -port2=8082

rh2:
	@chromium-browser --allow-running-insecure-content https://localhost:8082/hello
	./$(PROGRAM) -m=http_server -port=8080 -ports=8081 -port2=8082

rp:
	./$(PROGRAM) -m=http_player -url http://localhost:8080/stream

rp1:
	./$(PROGRAM) -m=http_player -url http://localhost:8080/hello
	./$(PROGRAM) -m=http_player -url http://localhost:8080/media/gopher.jpeg
	./$(PROGRAM) -m=http_player -url http://localhost:8080/static/image/gopher.jpg
	./$(PROGRAM) -m=http_player -url http://localhost:8080/index.html

rp2:
	./$(PROGRAM) -m=http_player -url https://localhost:8082/hello
	./$(PROGRAM) -m=http_player -url https://localhost:8082/media/gopher.jpeg
	./$(PROGRAM) -m=http_player -url https://localhost:8082/static/image/gopher.jpg
	./$(PROGRAM) -m=http_player -url https://localhost:8082/index.html

rc:
	./$(PROGRAM) -m=http_caster -url http://localhost:8080/stream

rr:
	./$(PROGRAM) -m=http_reader -url http://imoment:imoment@192.168.0.91/axis-cgi/mjpg/video.cgi

rra:
	./$(PROGRAM) -m=http_reader -url http://localhost:8050/agilecam\?action=stream\&channel=100\&source=1

rm:
	./$(PROGRAM) -m=http_monitor -port=8080 -ports=8081 -port2=8082

rs:
	./$(PROGRAM) -m=http_server -port=8080 -ports=8081 -port2=8082

rf:
	./$(PROGRAM) -m=file_reader -port=8080 -ports=8081 -port2=8082

rcam:
	/home/stoney/coding/imt-cam/imt-shot -s localhost:8087
	

# --- TCP
rts:
	./$(PROGRAM) -m=tcp_caster -port=8080 -ports=8081 -port2=8082

rtr:
	./$(PROGRAM) -m=tcp_server -port=8080 -ports=8081 -port2=8082

# --- WebSocket
rws:
	./$(PROGRAM) -m=ws_caster -port=8080 -ports=8081 -port2=8082

rwr:
	./$(PROGRAM) -m=ws_server -port=8080 -ports=8081 -port2=8082

rdt:
	echo -n "test out the server" | nc localhost 8080

rt:
	curl -I http://localhost:8080/stream

ping:
	ping -c 3 192.168.0.91

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

clobber:
	@make clean
	@cd src/sample && make clobber
	@cd src/streamimage && make clobber

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
	@make clobber
	git init
	git add * .gitignore
	git commit -m "refactor prototcp package with bufio and raw io"
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
