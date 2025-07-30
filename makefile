build:
	@go build -o bin/p2p

run: build
	@./bin/p2p

test: 
	@go test ./... -v

