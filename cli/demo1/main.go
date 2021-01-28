package main

import (
	"fmt"
	frank "github.com/kuking/go-frank"
	"github.com/kuking/go-frank/extras"
)

func main() {

	extras.Int64Generator(0, 100).
		Publish("hundred")

	frank.SubscribeNE("hundred").
		Map(func(i int64) int64 { return i * i }).
		Publish("hundred_power_two")

	fmt.Println(frank.SubscribeNE("hundred_power_two").Last())

}
