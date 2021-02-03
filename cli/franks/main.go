package main

import (
	"fmt"
	"github.com/kuking/go-frank/transport"
)

func main() {
	s := transport.Synchroniser{}
	err := s.ListenTCP(":4500")
	fmt.Println(err)
}
