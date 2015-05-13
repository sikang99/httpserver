.PHONY:	all tls-client make usage

#PROGRAM=tls-client
#PROGRAM=tls-simple
#PROGRAM=tls-example
#PROGRAM=file-upload
#PROGRAM=mjpeg-view
PROGRAM=goupload

all: usage

edit e:
	vi $(PROGRAM).go

html h:
	vi static/$(PROGRAM).html

build b:
	go build -o $(PROGRAM) $(PROGRAM).go

run r:
	./$(PROGRAM)

image:
	curl -I http://imoment:imoment@192.168.0.91/axis-cgi/jpg/image.cgi

image-save:
	curl http://imoment:imoment@192.168.0.91/axis-cgi/jpg/image.cgi > static/axis.jpeg

video:
	curl -I http://imoment:imoment@192.168.0.91/axis-cgi/mjpg/video.cgi

video-save:
	curl http://imoment:imoment@192.168.0.91/axis-cgi/mjpg/video.cgi > static/axis.mjpeg

clean:
	rm -f $(PROGRAM)

clobber:
	rm -f tls-simple tls-client tls-simple file-upload mjpeg-view goupload

git:
	make clobber
	cd .. && make gp

readme md:
	vi README.md

make m:
	vi Makefile

usage:
	@echo ""
	@echo "Makefile for $(PROGRAM)"
	@echo "usage: make [edit|html|build|runi|git]"
	@echo ""
