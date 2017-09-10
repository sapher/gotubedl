.PHONY: all

all: build

build: dist

test:
	go test

dist:
	mkdir dist

run-dev:
	go generate
	go build