.PHONY: build test lint install clean coverage

build:
	go build ./...

test:
	go test ./...
	go vet ./...

lint:
	golangci-lint run

install:
	go install ./...

coverage:
	go test ./... -coverprofile=coverage.out

clean:
	rm -f coverage.out
	rm -rf dist
