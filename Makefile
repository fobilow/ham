PHONY: build install
BUILD_VER :=$(shell date "+%Y.%m.%d.%H%M%S")
HASH :=$(shell git rev-parse --short HEAD)
build:
	go build -o ham cmd/ham/main.go
install:
	go build -ldflags="-X main.Version=$(HASH)-$(BUILD_VER)" -o $$GOPATH/bin/ham cmd/ham/main.go