PROGRAM=viewer

all: usage

edit e:
	vi $(PROGRAM).go

build b:
	go build -o $(PROGRAM) $(PROGRAM).go

run r:
	./$(PROGRAM) *.png

test t:
	go test -v

clean c:
	rm -f $(PROGRAM)

make m:
	vi Makefile

usage:
	@echo ""
	@echo "usage: make [edit|build|run|test]"
	@echo ""
