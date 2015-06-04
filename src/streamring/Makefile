#
# Makefile for package
#
PACKAGE=streamring

all: usage

edit e:
	vi $(PACKAGE).go

et:
	vi $(PACKAGE)_test.go

eb:
	vi $(PACKAGE)_bench.go

build b:
	go build

bench tb:
	go test -bench=.

test t:
	go test -v

buildtest bt:
	go build
	go test -v

make m:
	vi Makefile

usage:
	@echo ""
	@echo "usage: make [edit|build|test]"
	@echo ""
