package main

import (
	"fmt"
	frank "github.com/kuking/go-frank"
	"github.com/kuking/go-frank/api"
	"github.com/kuking/go-frank/extras"
	"github.com/kuking/go-frank/serialisation"
	"reflect"
)

func main() {

	// registers two memory streams in the registry
	_ = frank.Register("input", frank.EmptyStream(1024))
	_ = frank.Register("output", frank.EmptyStream(1024))

	// creates a persistent stream and registers it
	p, _ := frank.PersistentStream("persistent-stream", 65*1024*1024, serialisation.Int64Serialiser{})
	_ = frank.Register("persistent", p)

	// virtual stream 1: consumes from 'input' stream (memory), until it is closed, calculates the power of the number,
	// and publishes the output into the persistent stream (using client name 'calculator')
	_ = frank.SubscribeNE("input").
		EnsureType(reflect.Int64).
		Map(func(i int64) int64 { return i * i }).
		PublishClose("persistent?clientName=calculator")

	// mini-virtual stream 2: subscribes to the persistent stream with client name 'output' and sends the stream
	// traffic to the in-memory stream 'output', you can have multiple consumers so this can be done multiple times
	_ = frank.SubscribeNE("persistent?clientName=output").
		Wait(api.UntilClosed). // persistent streams won't consume until closed by default as they are long-lived streams
		PublishClose("output")

	// feeds 100 numbers (from 0 to 99) into the input, and closes the input stream.
	_ = extras.Int64Generator(0, 100).PublishClose("input")

	// Finally, the in memory output stream has some output
	fmt.Print("Inmemory stream name 'output': ")
	frank.SubscribeNE("output").ForEach(func(i int64) { fmt.Print(i, " ") })
	fmt.Println()

	// and the persistent one, by subscriben with client nae 'one' has some numbers too
	fmt.Print("Persistent stream client 'one': ")
	frank.SubscribeNE("persistent?clientName=one").ForEach(func(i int64) { fmt.Print(i, " ") })
	fmt.Println()

	fmt.Println("Closing")
	p.Close()
	//time.Sleep(time.Millisecond * 250) // gives time for consumes to close before the persistent stream dissapears

	// cleanup
	_ = p.Delete()
}
