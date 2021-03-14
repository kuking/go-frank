package main

import (
	"fmt"
	frank "github.com/kuking/go-frank"
	"github.com/kuking/go-frank/api"
	"github.com/kuking/go-frank/extras"
	"github.com/kuking/go-frank/serialisation"
	"reflect"
	"strings"
)

func registryPersistentDemo() {
	// registers two memory streams in the registry
	_ = frank.Register("input", frank.EmptyStream(1024))
	_ = frank.Register("output", frank.EmptyStream(1024))

	// creates a persistent stream and registers it
	p, _ := frank.PersistentStream("persistent-stream", 65*1024*1024, serialisation.Int64Serialiser{})
	_ = frank.Register("persistent", p)

	// virtual stream 1: consumes from 'input' stream (memory), until it is closed, calculates the power of the number,
	// and publishes the output into the persistent stream (using subscriber name 'calculator')
	_ = frank.SubscribeNE("input").
		EnsureType(reflect.Int64).
		Map(func(i int64) int64 { return i * i }).
		PublishClose("persistent?sn=calculator")

	// mini-virtual stream 2: subscribes to the persistent stream with subscriber name 'output' and sends the stream
	// traffic to the in-memory stream 'output', you can have multiple consumers so this can be done multiple times
	_ = frank.SubscribeNE("persistent?sn=output").
		Wait(api.UntilClosed). // persistent streams won't consume until closed by default as they are long-lived streams
		PublishClose("output")

	// feeds 25 numbers into the input, and closes the input stream.
	_ = extras.Int64Generator(0, 35).PublishClose("input")

	// Finally, the in memory output stream has some output
	fmt.Print("Inmemory stream name 'output': ")
	frank.SubscribeNE("output").ForEach(func(i int64) { fmt.Print(i, " ") })
	fmt.Println()

	// and the persistent one, by subscribing with subscriber name 'one' has some numbers too
	fmt.Print("Persistent stream subscriber 'one': ")
	frank.SubscribeNE("persistent?sn=one").ForEach(func(i int64) { fmt.Print(i, " ") })
	fmt.Println()

	// another subscriber can replay the stream, i.e. subscriber 'two'
	fmt.Print("Persistent stream subscriber 'two': ")
	frank.SubscribeNE("persistent?sn=two").ForEach(func(i int64) { fmt.Print(i, " ") })
	fmt.Println()

	// and the persistent one, by subscribing with subscriber name 'one' has some numbers too
	fmt.Print("Resetting persistent stream subscriber 'one':  ")
	frank.SubscribeNE("persistent?sn=one").Reset()
	frank.SubscribeNE("persistent?sn=one").ForEach(func(i int64) { fmt.Print(i, " ") })
	fmt.Println()

	fmt.Println("Closing")
	p.Close()

	// cleanup
	_ = p.Delete()
}

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

func main() {
	textFile()
	//registryPersistentDemo()
}
