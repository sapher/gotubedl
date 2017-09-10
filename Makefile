.PHONY: all

all: build

build:
	go build

test:
	go test

run-dev:
	go generate
	go build