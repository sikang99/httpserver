#
# Makefile for package
#
PACKAGE=image

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
	shotwell clock.png
	#shotwell output.png

clean:
	rm -f output.* clock.*

make m:
	vi Makefile

usage:
	@echo ""
	@echo "usage: make [edit|build|test]"
	@echo ""
