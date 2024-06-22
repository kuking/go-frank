
all: clean build test coverage benchmark binary

clean:
	rm -f coverage.out
	go clean -testcache -testcache

build:
	go build ./...

binary:
	# go build -o franki cli/franki/main.go
	# go build -o frankr cli/frankr/main.go

test:
	go test ./...

coverage:
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

benchmark:
	go test -run=Benchmark -bench=. ./...
