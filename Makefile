.PHONY: build test lint clean

build:
	go build -o build/bavarix ./cmd/bavarix

test:
	go test ./... -v -race

lint:
	golangci-lint run ./...

clean:
	rm -rf build/
