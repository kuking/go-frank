package benchmarks

import (
	"bytes"
	"encoding/binary"
	"github.com/kuking/go-frank/persistent"
	"github.com/kuking/go-frank/serialisation"
	"github.com/kuking/go-frank/transport"
	"os/user"
	"path"
	"testing"
)

// this assumes you have a persistent-stream in the root folder and an empty folder streams/
// $ rm persistent-stream.* streams/* && ./franki -miop 10 pub_bench
func testIntegrationWire(t *testing.T) {
	usr, _ := user.Current()
	root := path.Join(usr.HomeDir, "/KUKINO/go/go-frank")
	bind := "localhost:1234"
	r := transport.NewReplicator()
	stream, err := persistent.MmapStreamOpen(path.Join(root, "persistent-stream"), serialisation.ByteArraySerialiser{})
	if err != nil {
		t.Fatal(err)
	}
	go r.ListenTCP(bind, path.Join(root, "/streams/"))
	if err = r.ConnectTCP(stream, "repl1", bind); err != nil {
		t.Fatal(err)
	}
	r.WaitAll(true, true)
}

func Benchmark_WireDataMessage_BinaryWrite(b *testing.B) {
	wireDataMsg := transport.WireDataMsg{
		Version: transport.WireVersion,
		Message: transport.WireDATA,
		AbsPos:  123456,
		Length:  1234,
	}
	buffer := bytes.NewBuffer(make([]byte, 100))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Write(buffer, binary.LittleEndian, &wireDataMsg)
		buffer.Reset()
	}
}

func Benchmark_WireDataMessageNA_OwnWriter(b *testing.B) {
	wireDataMsgNA := transport.WireDataMsgNA{}
	wireDataMsgNA.SetVersion(transport.WireVersion)
	wireDataMsgNA.SetMessage(transport.WireDATA)
	wireDataMsgNA.SetAbsPos(123456)
	wireDataMsgNA.SetLength(1234)
	buffer := bytes.NewBuffer(make([]byte, 100))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wireDataMsgNA.Write(buffer)
		buffer.Reset()
	}
}
