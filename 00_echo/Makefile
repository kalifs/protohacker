GO ?= go
BINARY ?= bin/echo

.PHONY: build run clean

build:
	$(GO) build -o $(BINARY) ./main.go

run:
	$(GO) run ./main.go

clean:
	rm -f $(BINARY)
