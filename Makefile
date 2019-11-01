BUILD_FLAGS = "-v"

.PHONY: build clean test

all: build

build:
	go build ./...

clean:
	go clean -i ./...

test:
	go test -cover -race ./...