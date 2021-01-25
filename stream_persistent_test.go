package go_frank

import (
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

//func TestOpenCreatePersistentStream_MultiConsumerMultiProducer(t *testing.T) {
//	prefix, _ := ioutil.TempDir("", "MMAP-")
//	base := prefix + "/a-stream"
//	defer cleanup(prefix)

//ps1, err := OpenCreatePersistentStream(base, 64*1024, ByteArraySerialiser{})
//if err != nil {
//	t.Fatal()
//}
//ps2, err := OpenCreatePersistentStream(base, 64*1024, ByteArraySerialiser{})
//if err != nil {
//	t.Fatal()
//}

//for i:=0; i<1000; i++ {
//	ps1.Feed()
//}

//}
