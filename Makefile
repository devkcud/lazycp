all: build install

build:
	go build -v -o ./lcp ./main.go

install:
	cp ./lcp ~/.local/bin

uninstall:
	-rm ~/.local/bin/lcp

.PHONY: build install uninstall
