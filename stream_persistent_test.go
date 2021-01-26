package go_frank

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestOpenPersistentStream_SimpleConsume(t *testing.T) {
	prefix, _ := ioutil.TempDir("", "MMAP-")
	base := prefix + "/a-stream"
	defer cleanup(prefix)

	ps1, err := OpenCreatePersistentStream(base, 64*1024, ByteArraySerialiser{})
	if err != nil {
		t.Fatal()
	}
	ps2, err := OpenCreatePersistentStream(base, 64*1024, ByteArraySerialiser{})
	if err != nil {
		t.Fatal()
	}

	ps1.Feed("hello")
	ps1.Feed("lala")

	s := ps2.Consume("client-1")
	if asString(s.Pull().Get()) != "hello" || asString(s.Pull().Get()) != "lala" {
		t.Fatal()
	}
}

func TestOpenCreatePersistentStream_MultiConsumerMultiProducer(t *testing.T) {
	prefix, _ := ioutil.TempDir("", "MMAP-")
	base := prefix + "/a-stream"
	defer cleanup(prefix)

	ps1, err := OpenCreatePersistentStream(base, 64*1024, ByteArraySerialiser{})
	if err != nil {
		t.Fatal()
	}
	ps2, err := OpenCreatePersistentStream(base, 64*1024, ByteArraySerialiser{})
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
