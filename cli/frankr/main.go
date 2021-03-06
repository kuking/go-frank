package main

import (
	"flag"
	"fmt"
	"github.com/kuking/go-frank/persistent"
	"github.com/kuking/go-frank/serialisation"
	"github.com/kuking/go-frank/transport"
	"log"
	"os"
	"strconv"
	"strings"
)

func doArgsParsing() bool {

	if len(os.Args) <= 1 {
		fmt.Printf("Usage of %v: Frank Stream Bus replicator\n\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Print(`Examples:

  $ ./frankr send replicator_name@streams/file-stream hostname:port
  $ ./frankr accept streams/ bind_host:port optional_host 192.168.0.0/24
  $ ./frankr ps
  $ ./frankr top streams/
  $ ./frankr stop <pid>
`)
		return false
	}

	if len(os.Args) == 4 && os.Args[1] == "send" {
		parsed := strings.Split(os.Args[2], "@")
		if len(parsed) != 2 {
			fmt.Println("replicator_id and base_stream has to be provided, i.e. sub1@streams/number-one")
			return false
		}
		replName := parsed[0]
		basePath := parsed[1]
		hostName := os.Args[3]
		fmt.Println("so sending:", replName, "@", basePath, "->", hostName)
		r := transport.NewReplicator()

		stream, err := persistent.MmapStreamOpen(basePath, serialisation.ByteArraySerialiser{})
		if err != nil {
			log.Fatal(err)
		}
		err = r.ConnectTCP(stream, replName, hostName)
		if err != nil {
			log.Fatal(err)
		}
		r.WaitAll(false, false)
		return true
	} else if len(os.Args) > 3 && os.Args[1] == "accept" {
		basePath := os.Args[2]
		binding := os.Args[3]
		var accepted []string
		if len(os.Args) > 4 {
			accepted = os.Args[4:]
			fmt.Println("Warning: listening filters is not implemented yet.")
		} else {
			accepted = []string{"*"}
		}

		log.Printf("Accepting: %v; streams in: %v (accepting -not implemented- %v)\n", binding, basePath, accepted)
		r := transport.NewReplicator()
		go r.ListenTCP(binding, basePath)
		r.WaitAll(false, false)
		return true
	} else if len(os.Args) == 2 && os.Args[1] == "ps" {
		fmt.Println("so listing all the replication processes, just by name")
	} else if len(os.Args) == 3 && os.Args[1] == "top" {
		basePath := os.Args[2]
		fmt.Println("top streams in path", basePath)
		return true
	} else if len(os.Args) == 3 && os.Args[1] == "stop" {
		pid, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("could not parse the pid number, err:", err)
			return false
		}
		fmt.Println("trying to stop pid", pid)
		return true
	} else {
		fmt.Println("The arguments provided could not be understood.")
		return false
	}

	return false
}

func main() {
	if !doArgsParsing() {
		os.Exit(-1)
	}
}
