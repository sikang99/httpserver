.PHONY:	all tls-client make usage

PROGRAMS= \
	tls-client \
	tls-simple \
	tls-example \
	mjpeg-view \
	mjpeg-stream \
	mjpeg-proxy \
	websocket-chat \
	file-upload \
	goupload \
	url-parse \
	mixed-write \


PROGRAM=mixed-write

all: usage

edit e:
	vi $(PROGRAM).go

html h:
	vi static/$(PROGRAM).html

build b:
	go build -o $(PROGRAM) $(PROGRAM).go

run r:
	./$(PROGRAM)
	#./$(PROGRAM) -port=8081

kill k:
	killall $(PROGRAM)

image:
	curl -I http://imoment:imoment@192.168.0.91/axis-cgi/jpg/image.cgi

image-save:
	curl http://imoment:imoment@192.168.0.91/axis-cgi/jpg/image.cgi > static/axis.jpeg
	shotwell static/axis.jpeg

video:
	curl -I http://imoment:imoment@192.168.0.91/axis-cgi/mjpg/video.cgi

video-save:
	curl http://imoment:imoment@192.168.0.91/axis-cgi/mjpg/video.cgi > static/axis.mjpeg

clean:
	rm -f $(PROGRAM)

clobber:
	rm -f $(PROGRAMS)

git:
	cd .. && make gp

readme md:
	vi README.md

make m:
	vi Makefile

usage:
	@echo ""
	@echo "Makefile for '$(PROGRAM)'"
	@echo "usage: make [edit|html|build|run|kill|git]"
	@echo ""
