#
# WebSocket Echo Test
# http://jan.newmarch.name/go/websockets/chapter-websockets.html
#
all: usage

edit e:
	@echo ""
	@echo "make (edit) [ec|es|ess]"
	@echo ""
	
ec:
	vi echoclient.go

es:
	vi echoserver.go

ess:
	vi echoservertls.go


build b:
	@echo ""
	@echo "make (build) [bc|bs|bss|ba]"
	@echo ""

bc:
	go build -o echoclient echoclient.go

bs:
	go build -o echoserver echoserver.go

bss:
	go build -o echoservertls echoservertls.go

ba:
	go build -o echoclient echoclient.go
	go build -o echoserver echoserver.go
	go build -o echoservertls echoservertls.go

run r:
	@echo ""
	@echo "make (run) [rc|rcs|rs|rss]"
	@echo ""
	
rc:
	./echoclient ws://localhost:8080/echo

rcs:
	./echoclient wss://localhost:8080/echo

rs:
	./echoserver &

rss:
	./echoservertls &

ps:
	@ps -ef | grep echo | grep -v grep

clobber:
	rm -f echoclient echoserver echoservertls

kill k:
	killall echoserver echoservertls

gencert:
	openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 9999 -nodes

make m:
	vi Makefile

usage:
	@echo ""
	@echo "usage: make [edit|build|run|kill]"
	@echo ""
