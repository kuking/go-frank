# go-frank streaming framework

Two flavours: in-memory and persistent, both are multiple-producers, multiple-consumers; raw -fast- stream replicator.

It is fast (see [PERF.md](PERF.md)), it won't use typical mutexes, locks, channels, etc. It incorporates ideas
from: [Aeron](https://github.com/real-logic/Aeron),
and [Mechanical Sympathy blog](https://mechanical-sympathy.blogspot.com).

It tries to find a balance between maximum possible performance, which comes at high cost i.e. flyweights,
non-allocation, buffer pools, etc. and a practical simple framework that performs very fast with low-latency.

**Status:** Mostly works, you are welcome to participate.

## Fluid syntax
Extracts from [demo1](cli/demo1/main.go):
```go
func textFile() {
	lines := frank.TextFileStream("README.md").Count()
	chars := frank.TextFileStream("README.md").
		Map(func(line string) int { return len(line) + 1 }).
		Sum().
		First()
	fmt.Printf("README.md has %v lines and %v characters.\n", lines, chars)

	title := frank.TextFileStream("README.md").
		Filter(func(s string) bool { return len(s) < 1 || s[0] != '#' }).
		Map(func(s string) string { return strings.TrimSpace(s[1:]) }).
		First()
	fmt.Printf("README.md title is: %v\n", title)
}
```