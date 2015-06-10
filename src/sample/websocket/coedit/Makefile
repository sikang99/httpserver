PROGRAM=collabedit

all: usage

edit e:
	@echo ""
	@echo "usage: make (edit) [main|css|js|html]"
	@echo ""

main em:
	vi src/main.go

css ec:
	vi public/main.css

js ej:
	vi public/main.js

html eh:
	vi public/index.html


build b:
	go build -o $(PROGRAM) src/*.go

run r:
	./$(PROGRAM)

clean c:
	rm -f $(PROGRAM)

make m:
	vi Makefile

usage:
	@echo ""
	@echo "usage: make [edit|build|run]"
	@echo ""
