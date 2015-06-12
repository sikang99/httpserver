#
# Makefile for happy package
#
PACKAGE=prototcp

all: usage

edit e:
	vi $(PACKAGE).go

et:
	vi $(PACKAGE)_test.go

build b:
	go build

test t:
	go test -v

btf1:
	go build
	go test -run TestSendReceive -timeout 60m

btf2:
	go build
	go test -run TestCastServePlay

buildtest bt:
	go build
	go test -v

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
