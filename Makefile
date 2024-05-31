PHONY: build install

build:
	go build -o ham cmd/ham/main.go
install:
	go build -o $$GOPATH/bin/ham cmd/ham/main.go