# Makefile for http server supporting http, https, http2

.PHONY:	edit build run rebuild install clean make

all: usage

PROGRAM=httpserver
EDITOR=vim

edit e:
	$(EDITOR) $(PROGRAM).go

readme md:
	$(EDITOR) README.md

build b:
	#go build $(PROGRAM).go
	go build 
	@ls -alF --color=auto

run r:
	@chromium-browser http://127.0.0.1:8000/hello
	./$(PROGRAM) -port=8000 -sport=8001

test t:
	@chromium-browser -insecure https://127.0.0.1:8001/static
	./$(PROGRAM) -port=8000 -sport=8001

rclient rc:
	./$(PROGRAM) -url http://localhost:8000/hello -port=8000 -sport=8001
	./$(PROGRAM) -url http://localhost:8000/index
	./$(PROGRAM) -url http://localhost:8000/media

rserver rs:
	./$(PROGRAM) -d -port=8000 -sport=8001

rtest rt:
	./$(PROGRAM) -t

rebuild:
	rm -f ./$(PROGRAM)
	go build $(PROGRAM).go
	@ls -alF --color=auto

install i:
	go install

clean:
	rm -f ./$(PROGRAM)

make m:
	$(EDITOR) Makefile

# ---------------------------------------------------------------------------
git-hub gh:
	ssh -T git@github.com


git-pull gd:
	git push -u https://sikang99@github.com/sikang99/$(PROGRAM) master

git-push gu:
	git init
	git add *
	git commit -m "add config.go"
	git push -u https://sikang99@github.com/sikang99/$(PROGRAM) master
	chromium-browser https://github.com/sikang99/$(PROGRAM)

git-status gs:
	git status
	git log --oneline -5

gencert:
	openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 9999 -nodes

# ---------------------------------------------------------------------------
usage:
	@echo ""
	@echo "Makefile for '$(PROGRAM)', by Stoney Kang, 2015/04/18"
	@echo ""
	@echo "usage: make [edit|readme|build|run|test|rebuild|clean|git]"
	@echo "	edit    : edit source"
	@echo "	build   : compile source"
	@echo "	run     : execute $(PROGRAM)"
	@echo "	test    : test $(PROGRAM) options"
	@echo "	install : install $(PROGRAM) to $(GOPATH)/bin"
	@echo "	git-push : upload $(PROGRAM) to github.com"
	@echo "	git-pull : fetch $(PROGRAM) from github.com"
	@echo "	gencert  : make certificates for https"
	@echo ""
