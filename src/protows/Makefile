#
# Makefile for happy package
#
PACKAGE=protows

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
