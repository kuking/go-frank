
all: clean build test coverage benchmark

clean:
	go clean -testcache -testcache

build:
	go build ./...

test:
	go test ./...

coverage:
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

benchmark:
	go test -run=Benchmark -bench=.

memory: clean
	go tool compile "-m" stream.go # -S
