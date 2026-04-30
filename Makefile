.PHONY: build test lint ci

build:
	go build -o finalcountdown .

test:
	go test ./...

lint:
	go vet ./...

ci: build test lint
