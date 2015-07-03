#
# Makefile for httpserver package
#
PACKAGE=protowsm

all: usage

#--------------------------------------------------------------------------------------
edit:
	@echo ""
	@echo "usage: make (edit) [e|et|em]"
	@echo ""

e:
	vi $(PACKAGE).go

et:
	vi $(PACKAGE)_test.go

em:
	vi echo.html

build b:
	go build

#--------------------------------------------------------------------------------------
test t:
	@echo ""
	@echo "usage: make (test) [bte1|bte2|btf1|btf2|btf3|btf4]"
	@echo ""

bt:
	go build
	go test -v

#--------------------------------------------------------------------------------------
bte:
	@go build
	go test -run TestEcho

bte1:
	@go build
	go test -run TestEchoReceive

bte2:
	@go build
	go test -run TestEchoSendReceive

#--------------------------------------------------------------------------------------
btf:
	go build
	go test -run TestStream

btf1:
	go build
	go test -run TestStreamServe

btf2:
	go build
	go test -run TestStreamCastServe

btf3:
	go build
	go test -run TestStreamServePlay

btf4:
	go build
	go test -run TestStreamCastServePlay

#--------------------------------------------------------------------------------------
gencert:
	openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 9999 -nodes

view v:
	shotwell output.jpg

clobber clean:
	rm -f output.*

make m:
	vi Makefile

usage:
	@echo ""
	@echo "usage: make [edit|build|test|view]"
	@echo ""
