
all: clean build test coverage benchmark binary

clean:
	rm -f franki
	go clean -testcache -testcache

build:
	go build ./...

binary:
	go build -o franki cli/franki/main.go

test:
	go test ./...

coverage:
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

benchmark:
	go test -run=Benchmark -bench=. ./benchmarks

memory: clean
	go tool compile "-m" stream.go # -S

