#
# Makefile for happy package
#
PACKAGE=protofile

all: usage

edit e:
	vi $(PACKAGE).go

et:
	vi $(PACKAGE)_test.go

build b:
	go build

test t:
	go test -v

buildtest bt:
	go build
	go test -v

btf:
	go build
	go test -run TestReadMultipartFile

view v:
	#shotwell output.jpg
	hexer output.mjpg

clobber clean:
	rm -f output.*

make m:
	vi Makefile

usage:
	@echo ""
	@echo "usage: make [edit|build|test|view]"
	@echo ""
