package main

import (
	"fmt"
	frank "github.com/kuking/go-frank"
	"github.com/kuking/go-frank/api"
	"github.com/kuking/go-frank/extras"
	"reflect"
)

func main() {

	// registers an empty in memory stream with name 'inmemory' in the registry
	_ = frank.Register("input", frank.EmptyStream(1024))
	_ = frank.Register("output", frank.EmptyStream(1024))

	// creates a persistent stream and register it into the registry
	//p, _ := frank.PersistentStream("persistent-stream", 65*1024*1024, &serialisation.GobSerialiser{})
	//_ = frank.Register("persistent", p)

	// virtual stream: consumes from 'inmemory' stream and pushes after transformation back into the 'persistent' stream
	_ = frank.SubscribeNE("input").
		Wait(api.UntilClosed).
		EnsureType(reflect.Int64).
		Map(func(i int64) int64 { return i * i }).
		Publish("output")
	//Publish("persistent?clientName=powerTwoCalculator")

	// feeds 100 numbers into the input
	_ = extras.Int64Generator(0, 100).Publish("input")

	// then the persistent output has to have 1000 powers
	//frank.SubscribeNE("persistent?clientName=one").Wait(api.WaitingUpto1s).ForEach(func(i int64) { fmt.Print(i, " ") })
	frank.SubscribeNE("output").Wait(api.WaitingUpto1s).ForEach(func(i int64) { fmt.Print(i, " ") })

	fmt.Println("Closing")

	// cleanup
	//_ = p.Delete()
}
