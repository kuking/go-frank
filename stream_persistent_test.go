package go_frank

import (
	"fmt"
	"github.com/kuking/go-frank/api"
	"github.com/kuking/go-frank/serialisation"
	"io/ioutil"
	"testing"
	"time"
)

func TestOpenPersistentStream_SimpleConsume(t *testing.T) {
	prefix, _ := ioutil.TempDir("", "MMAP-")
	base := prefix + "/a-stream"
	defer cleanup(prefix)

	ps1, err := OpenCreatePersistentStream(base, 64*1024, serialisation.ByteArraySerialiser{})
	if err != nil {
		t.Fatal()
	}
	ps2, err := OpenCreatePersistentStream(base, 64*1024, serialisation.ByteArraySerialiser{})
	if err != nil {
		t.Fatal()
	}

	ps1.Feed("hello")
	ps1.Feed("lala")

	s := ps2.Consume("client-1")
	if serialisation.AsString(s.Pull().Get()) != "hello" || serialisation.AsString(s.Pull().Get()) != "lala" {
		t.Fatal()
	}
}

func TestOpenCreatePersistentStream_MultiConsumerMultiProducer(t *testing.T) {
	prefix, _ := ioutil.TempDir("", "MMAP-")
	base := prefix + "/a-stream"
	defer cleanup(prefix)

	ps1, err := OpenCreatePersistentStream(base, 64*1024, serialisation.ByteArraySerialiser{})
	if err != nil {
		t.Fatal()
	}
	ps2, err := OpenCreatePersistentStream(base, 64*1024, serialisation.ByteArraySerialiser{})
	if err != nil {
		t.Fatal()
	}

	for i := 0; i < 1000; i++ {
		ps1.Feed(fmt.Sprintf("hello PS1:%v", i))
		ps2.Feed(fmt.Sprintf("hello PS2:%v", i))
	}
	ps1.Close()

	if ps1.Consume("consumer-1").Count() != 2000 {
		t.Fatal()
	}
	if !ps2.Consume("consumer-1").IsClosed() { // 2nd persistent stream sees 1st one
		t.Fatal()
	}
	if ps2.Consume("consumer-2").Count() != 2000 {
		t.Fatal()
	}
	if !ps1.Consume("consumer-2").IsClosed() {
		t.Fatal()
	}
	if !ps1.IsClosed() || !ps2.IsClosed() {
		t.Fatal()
	}
	if ps1.Consume("another").Count() != 2000 {
		t.Fatal()
	}

}

func TestWaitApproachPersistent(t *testing.T) {
	prefix, _ := ioutil.TempDir("", "MMAP-")
	base := prefix + "/a-stream"
	defer cleanup(prefix)
	p, _ := OpenCreatePersistentStream(base, 64*1024, serialisation.ByteArraySerialiser{})

	s := p.Consume("lala")
	s.Wait(api.WaitingUpto10ms)
	t0 := time.Now()
	if s.Count() != 0 {
		t.Fatal()
	}
	dur := time.Now().Sub(t0)
	if dur.Milliseconds() < 10 {
		t.Fatal(fmt.Sprintf("it should have waited at least 10ms, but it wait: %v", dur))
	}

	s.Feed("1")
	s.Wait(api.UntilNoMoreData)
	t0 = time.Now()
	if s.Count() != 1 {
		t.Fatal()
	}
	dur = time.Now().Sub(t0)
	if dur.Nanoseconds() > 1_000_000 {
		t.Fatal(fmt.Sprintf("it should have been lots faster, took: %v", dur))
	}
}

func TestPersistentWithZeroLengthElementsWorks(t *testing.T) {
	prefix, _ := ioutil.TempDir("", "MMAP-")
	base := prefix + "/a-stream"
	defer cleanup(prefix)
	p, _ := OpenCreatePersistentStream(base, 64*1024, serialisation.ByteArraySerialiser{})

	s := p.Consume("lala")
	s.Feed([]byte{})

	elemOp := s.Pull()
	if elemOp.IsEmpty() || len(elemOp.Get().([]byte)) != 0 {
		t.Fatal()
	}
}
