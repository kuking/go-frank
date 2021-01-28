package main

import (
	frank "github.com/kuking/go-frank"
	"github.com/kuking/go-frank/extras"
	"github.com/kuking/go-frank/serialisation"
)

func main() {

	extras.Int64Generator(0, 100).
		Publish("hundred")

	p, _ := frank.PersistentStream("persistent-stream", 65*1024*1024, serialisation.ByteArraySerialiser{})
	p.Publish("persistent")

	frank.SubscribeNE("hundred").
		Map(func(i int64) int64 { return i * i }).
		Publish("hundred_power_two")

	frank.SubscribeNE("hundred").Publish("persistent?clientName=pub1")

	//fmt.Println(frank.SubscribeNE("hundred_power_two").Last())

}
