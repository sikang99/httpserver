.PHONY:	all tls-client make usage

#PROGRAM=tls-client
PROGRAM=tls-simple

all: usage

edit e:
	vi $(PROGRAM).go

build b:
	go build -o $(PROGRAM) $(PROGRAM).go

run r:
	./$(PROGRAM)

clean:
	rm -f $(PROGRAM)

clobber:
	rm -f tls-simple tls-client

make m:
	vi Makefile

usage:
	@echo ""
	@echo "make [tls-client]"
	@echo ""
