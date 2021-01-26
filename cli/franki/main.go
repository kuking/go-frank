package main

import (
	"bufio"
	"flag"
	"fmt"
	frank "github.com/kuking/go-frank"
	"log"
	"os"
)

var partSize uint64
var baseFile string
var clientName string
var waitApproach int64
var doHelp bool
var beQuiet bool
var cmd string

func doArgsParsing() bool {
	flag.Uint64Var(&partSize, "ps", 256, "part size in Mb.")
	flag.StringVar(&baseFile, "bs", "persistent-stream", "Base file path")
	flag.StringVar(&clientName, "cn", "client-1", "Client name")
	flag.Int64Var(&waitApproach, "wa", int64(frank.UntilNoMoreData), "Wait approach: -1 until closed, 0 until no more data, N ns wait.")
	flag.BoolVar(&beQuiet, "q", false, "Be quiet, better for performance stats")
	flag.BoolVar(&doHelp, "h", false, "Show usage")
	flag.Parse()
	if doHelp || flag.NArg() != 1 {
		fmt.Printf("Usage of %v: Frank Stream Bus persistent file utility\n\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Print(`

COMMANDS:
  - pub  : publish lines coming from stdin into the persistent stream
  - sub  : subscribe and output to stdout from the persistem stream
  - info : describes the stream state

Examples: 
  $ cat file | franki -ps 1024 -bs test pub
  $ franki -bs test -sn sub-2 sub
`)
		return false
	}

	cmd = os.Args[len(os.Args)-1]
	if cmd != "pub" && cmd != "sub" && cmd != "info" {
		fmt.Println("Unknown command:", cmd)
		return false
	}
	return true
}

func main() {
	if !doArgsParsing() {
		os.Exit(-1)
	}

	p, err := frank.OpenCreatePersistentStream(baseFile, partSize*1024*1024, frank.ByteArraySerialiser{})
	if err != nil {
		log.Fatal(err)
	}

	if cmd == "sub" {
		s := p.Consume(clientName)
		s.Wait(frank.WaitApproach(waitApproach))
		s.ForEach(func(elem []byte) {
			fmt.Println(string(elem))
		})
	} else if cmd == "pub" {
		s := p.Consume(clientName)
		s.Wait(frank.WaitApproach(waitApproach))
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			s.Feed(scanner.Bytes())
		}
	} else if cmd == "info" {
		fmt.Println("need to implement")
	}

	if err := p.CloseFile(); err != nil {
		log.Fatal(err)
	}

}
