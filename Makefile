# Makefile for http server supporting http, https, http2

.PHONY:	edit build run rebuild install clean make

all: usage

PROGRAM=httpserver
EDITOR=vim

edit e:
	$(EDITOR) server/$(PROGRAM).go

ew:
	$(EDITOR) base/wrapper.go

readme md:
	$(EDITOR) README.md

build b:
	#go build $(PROGRAM).go
	go build -o $(PROGRAM) server/$(PROGRAM).go
	@ls -alF --color=auto

# testing http
run r:
	@chromium-browser http://127.0.0.1:8000/hello
	./$(PROGRAM) -d -port=8000 -ports=8001

# testing http2
test t:
	chromium-browser --allow-running-insecure-content https://127.0.0.1:8002/static
	./$(PROGRAM) -d -port=8000 -ports=8001 -port2=8002

rclient rc:
	./$(PROGRAM) -url http://localhost:8000/hello
	./$(PROGRAM) -url https://localhost:8001/media
	./$(PROGRAM) -url https://localhost:8002/index.html
	#./$(PROGRAM) -url https://localhost:8002/README.md

rserver rs:
	./$(PROGRAM) -d -port=8000 -sport=8001

rmonitor rm:
	./$(PROGRAM) -m

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

make m:
	$(EDITOR) Makefile

# ---------------------------------------------------------------------------
git-view gview gv:
	LANG=C chromium-browser https://github.com/sikang99/$(PROGRAM)

git-hub gh:
	ssh -T git@github.com


git-pull gpull gd:
	git push -u https://sikang99@github.com/sikang99/$(PROGRAM) master

git-push gpush gp:
	git init
	git add * .gitignore
	git commit -m "add samples for tls coding"
	git push -u https://sikang99@github.com/sikang99/$(PROGRAM) master
	#chromium-browser https://github.com/sikang99/$(PROGRAM)

git-status gs:
	git status
	git log --oneline -5

gencert:
	openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 9999 -nodes

# ---------------------------------------------------------------------------
usage:
	@echo ""
	@echo "Makefile for '$(PROGRAM)', by Stoney Kang, 2015/04/24"
	@echo ""
	@echo "usage: make [edit|readme|build|run|test|rebuild|clean|git]"
	@echo "	edit(e)    : edit source"
	@echo "	build(b)   : compile source"
	@echo "	run(r)     : execute $(PROGRAM)"
	@echo "	run(rm)    : $(PROGRAM) monitor options"
	@echo "	run(rc)    : $(PROGRAM) client options"
	@echo "	run(rs)    : $(PROGRAM) server options"
	@echo "	install(i) : install $(PROGRAM) to $(GOPATH)/bin"
	@echo "	git-push(gu) : upload $(PROGRAM) to github.com"
	@echo "	git-pull(gp) : fetch $(PROGRAM) from github.com"
	@echo "	git-view(gv) : browse $(PROGRAM) at github.com"
	@echo "	gencert  : make certificates for https"
	@echo ""
