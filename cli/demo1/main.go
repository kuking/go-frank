package main

import (
	"fmt"
	frank "github.com/kuking/go-frank"
	"github.com/kuking/go-frank/api"
	"github.com/kuking/go-frank/extras"
	"github.com/kuking/go-frank/serialisation"
	"reflect"
	"time"
)

func main() {

	// registers two memory streams in the registry
	_ = frank.Register("input", frank.EmptyStream(1024))
	_ = frank.Register("output", frank.EmptyStream(1024))

	// creates a persistent stream and registers it
	p, _ := frank.PersistentStream("persistent-stream", 65*1024*1024, serialisation.Int64Serialiser{})
	_ = frank.Register("persistent", p)

	// virtual stream 1: consumes from 'input' stream (memory) transforms it, and publishes it into the persistent one
	_ = frank.SubscribeNE("input").
		Wait(api.UntilClosed).
		EnsureType(reflect.Int64).
		Map(func(i int64) int64 { return i * i }).
		Publish("persistent?clientName=calculator")

	// mini-virtual stream 2: subscribes to the persistent stream with client name 'output' and sends its output to the
	// memory stream 'output'
	_ = frank.SubscribeNE("persistent?clientName=output").Wait(api.UntilClosed).Publish("output")

	// feeds 100 numbers into the input
	_ = extras.Int64Generator(0, 100).Publish("input")

	// then the persistent output has to have 1000 powers
	fmt.Print("Persistent stream client 'one': ")
	frank.SubscribeNE("persistent?clientName=one").Wait(api.WaitingUpto1s).ForEach(func(i int64) { fmt.Print(i, " ") })
	fmt.Println()
	//p.Close()

	fmt.Print("Inmemory stream name 'output': ")
	frank.SubscribeNE("output").Wait(api.WaitingUpto1s).ForEach(func(i int64) { fmt.Print(i, " ") })
	fmt.Println()
	fmt.Println("Closing")

	p.Close()
	time.Sleep(time.Millisecond * 250) // gives time for consumes to close before the persistent stream dissapears

	// cleanup
	_ = p.Delete()
}
