BINARY  := bin/ssh-battos
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X github.com/batt0s/ssh.battos/tui.Version=$(VERSION)

.PHONY: run local build test vet fmt tidy demo docker clean

## run: start the SSH server
run:
	go run .

## local: run the TUI in the current terminal (no SSH)
local:
	go run . --local

## build: compile a versioned binary to bin/
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

## test: run the test suite
test:
	go test ./...

## vet: run go vet
vet:
	go vet ./...

## fmt: format all Go code
fmt:
	gofmt -w .

## tidy: tidy go.mod/go.sum
tidy:
	go mod tidy

## demo: build and record demo.gif with vhs
demo: build
	vhs demo.tape

## docker: build the container image
docker:
	docker build -t ssh-battos .

## clean: remove build artifacts
clean:
	rm -rf bin demo.gif
