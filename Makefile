.PHONY: all test clean
all:
	go build
test:
	go test ./...
clean:
	go clean
