package main

import (
	"flag"
	"fmt"
	"github.com/kuking/go-frank/transport"
	"os"
)

var doHelp bool
var doSetReplica bool
var doDelReplica bool
var baseFile string
var replName string
var hostname string
var acceptNewPush bool
var binding string
var doReceive bool
var doSend bool

func doArgsParsing() bool {

	flag.StringVar(&baseFile, "bs", "persistent-stream", "Base file path")
	flag.StringVar(&replName, "rn", "repl-1", "Replicator name")
	flag.StringVar(&hostname, "host", "", "Hostname")
	flag.BoolVar(&doSetReplica, "set", false, "set replica configuration")
	flag.BoolVar(&doDelReplica, "del", false, "delete replica configuration")
	flag.BoolVar(&acceptNewPush, "accept", false, "Accept new replica pushes")
	flag.StringVar(&binding, "bind", ":4500", "Listening host:port to bind to")
	flag.BoolVar(&doReceive, "receive", false, "Receive replicas (already accepted ones, or new ones if open to accept)")
	flag.BoolVar(&doSend, "send", false, "Send replicas (already configured to be replicated.)")
	flag.BoolVar(&doHelp, "h", false, "Show usage")
	flag.Parse()
	if doHelp {
		fmt.Printf("Usage of %v: Frank Stream Bus replicator\n\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Print(`

Examples:

  $ ./frankr -bs stream-name -rn repl-to-hostA -host hostA -set
  $ ./frankr -bs stream-name -rn repl-to-hostA -del
  $ ./frankr -bs ./ -receive -send
  $ ./
`)
		return false
	}

	if doSetReplica && doDelReplica {
		fmt.Println("Can not set and delete replica configuration at the same time")
		return false
	} else if (doSetReplica || doDelReplica) && (doReceive && doSetReplica && acceptNewPush) {
		fmt.Println("Replica configuration and replication server settings can not be configured at the same time")
		return false
	} else if !doSetReplica && !doDelReplica && !doReceive && !doSend {
		fmt.Println("No command provided, i.e. -set -del -send -receive")
		return false
	}
	return true
}

func main() {
	if !doArgsParsing() {
		os.Exit(-1)
	}
	s := transport.Replicator{}
	err := s.ListenTCP(":4500")
	fmt.Println(err)
}
