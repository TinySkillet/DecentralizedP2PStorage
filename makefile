DEFAULT_TARGET: build

.PHONY: build run test

build:
	@go build -o bin/p2p

run: 
	@./bin/p2p

test: 
	@go test ./... -v

