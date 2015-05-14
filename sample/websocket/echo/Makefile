.PHONY: all edit build run make usage

all: usage

edit e:
	vi mirror.go

build b:
	go build -o mirror mirror.go

run r:
	@echo ""
	@echo "make (run) [rc|rs]"
	@echo ""

rc:
	./mirror

rs:
	./mirror -d

kill k:
	killall mirror

readme md:
	vi README.md

make m:
	vi Makefile

usage:
	@echo ""
	@echo "usage: make [edit|build|run]"
	@echo ""
