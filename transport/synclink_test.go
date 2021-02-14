package transport

import (
	"encoding/binary"
	"fmt"
	"github.com/kuking/go-frank/misc"
	"github.com/kuking/go-frank/persistent"
	"github.com/kuking/go-frank/serialisation"
	"io"
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"
)

func TestSyncLink_GoFuncSend_Close(t *testing.T) {
	ctx := setup(t)
	defer teardown(ctx)
	sl := ctx.repl.NewSyncLinkSend(ctx.sendPipe, "host:1234", ctx.sendStream, "repl-1")
	go sl.goFuncSend()

	sl.Close()

	assertWait(func() bool { return sl.State == DISCONNECTED }, 100*time.Millisecond, t)
	assertIsClosed(ctx.sendPipe, t)
}

func TestSyncLink_GoFuncSend_DetectsClosedConnection(t *testing.T) {
	ctx := setup(t)
	defer teardown(ctx)
	sl := ctx.repl.NewSyncLinkSend(ctx.sendPipe, "host:1234", ctx.sendStream, "repl-1")
	_ = ctx.recvPipe.Close()
	go sl.goFuncSend()

	assertWait(func() bool { return sl.State == DISCONNECTED }, 100*time.Millisecond, t)
	assertIsClosed(ctx.sendPipe, t)
}

func TestSyncLink_GoFuncSend_DoHelloAndStatus(t *testing.T) {
	ctx := setup(t)
	defer teardown(ctx)
	sl := ctx.repl.NewSyncLinkSend(ctx.sendPipe, "host:1234", ctx.sendStream, "repl-1")
	go sl.goFuncSend()

	var wireHelloMsg WireHelloMsg
	if err := binary.Read(ctx.recvPipe, binary.LittleEndian, &wireHelloMsg); err != nil {
		t.Fatal(err)
	}
	if wireHelloMsg.Version != WireVersion ||
		wireHelloMsg.Message != WireHELLO ||
		wireHelloMsg.StreamUniqId != ctx.sendStream.GetUniqId() ||
		wireHelloMsg.PartSize != ctx.sendStream.GetPartSize() ||
		wireHelloMsg.FirstPart != ctx.sendStream.GetFirstPart() {
		t.Fatal("WireHelloMsg does not seems correct")
	}

	var wireStatusMsg WireStatusMsg
	if err := binary.Read(ctx.recvPipe, binary.LittleEndian, &wireStatusMsg); err != nil {
		t.Fatal(err)
	}
	if wireStatusMsg.Version != WireVersion ||
		wireStatusMsg.Message != WireSTATUS ||
		wireStatusMsg.FirstPart != ctx.sendStream.GetFirstPart() ||
		wireStatusMsg.PartsCount != ctx.sendStream.GetPartsCount() ||
		(!misc.Uint32Bool(wireStatusMsg.Closed) && ctx.sendStream.IsClosed()) ||
		(misc.Uint32Bool(wireStatusMsg.Closed) && !ctx.sendStream.IsClosed()) {
		t.Fatal("WireStatusMsg does not seems correct")
	}

	closeAndVerify(sl, ctx)
}

func TestSyncLink_GoFuncSend_SendsStream(t *testing.T) {
	ctx := setup(t)
	defer teardown(ctx)
	sl := ctx.repl.NewSyncLinkSend(ctx.sendPipe, "host:1234", ctx.sendStream, "repl-1")
	go sl.goFuncSend()

	initialHandShakeDone(ctx)
	feedStream(ctx.sendStream, 100)

	var wireDataMsg WireDataMsg
	buf := make([]byte, 10)
	prevAbsPos := uint64(0)
	for i := 0; i < 100; i++ {
		if err := binary.Read(ctx.recvPipe, binary.LittleEndian, &wireDataMsg); err != nil {
			t.Fatal(err)
		}
		if wireDataMsg.Version != WireVersion ||
			wireDataMsg.Message != WireDATA ||
			wireDataMsg.AbsPos < prevAbsPos ||
			wireDataMsg.Length != 8 { // fixed, it is an uint64
			t.Fatal("WireDataMsg does not seems correct")
		}
		if n, err := ctx.recvPipe.Read(buf[0:8]); n != 8 || err != nil {
			t.Fatal(err)
		}
		if binary.LittleEndian.Uint64(buf[0:8]) != uint64(i) {
			t.Fatal("value is not correct")
		}
	}

	closeAndVerify(sl, ctx)
}

func feedStream(stream *persistent.MmapStream, qty int) {
	for i := 0; i < qty; i++ {
		elem := [8]byte{}
		binary.LittleEndian.PutUint64(elem[:], uint64(i))
		stream.Feed(elem[:])
	}
}

// -------------------------------------------------------------------------------------------------------------------

func initialHandShakeDone(ctx *context) {
	var wireHelloMsg WireHelloMsg
	if err := binary.Read(ctx.recvPipe, binary.LittleEndian, &wireHelloMsg); err != nil {
		ctx.t.Fatal(err)
	}
	var wireStatusMsg WireStatusMsg
	if err := binary.Read(ctx.recvPipe, binary.LittleEndian, &wireStatusMsg); err != nil {
		ctx.t.Fatal(err)
	}
}

func closeAndVerify(sl *SyncLink, ctx *context) {
	sl.Close()
	assertWait(func() bool { return sl.State == DISCONNECTED }, 500*time.Millisecond, ctx.t)
	assertIsClosed(ctx.sendPipe, ctx.t)
	assertIsClosed(ctx.recvPipe, ctx.t)

}

// -------------------------------------------------------------------------------------------------------------------

type context struct {
	t          *testing.T
	repl       *Replicator
	prefix     string
	sendPath   string
	recvPath   string
	sendStream *persistent.MmapStream
	recvStream *persistent.MmapStream
	sendPipe   net.Conn
	recvPipe   net.Conn
}

func setup(t *testing.T) *context {
	prefix, err := ioutil.TempDir("", "MMAP-")
	if err != nil {
		panic(err)
	}
	sendPath := prefix + "/send-stream"
	recvPath := prefix + "/recv-stream"
	sendStream, _ := persistent.MmapStreamCreate(sendPath, 64*1024*1024, serialisation.ByteArraySerialiser{})
	recvStream, _ := persistent.MmapStreamCreate(recvPath, 64*1024*1024, serialisation.ByteArraySerialiser{})
	sendPipe, recvPipe := net.Pipe()
	_ = sendPipe.SetDeadline(time.Now().Add(5 * time.Second)) // no test should take longer than 5 seconds
	_ = recvPipe.SetDeadline(time.Now().Add(5 * time.Second))
	return &context{
		t:          t,
		repl:       NewReplicator(),
		prefix:     prefix,
		sendPath:   sendPath,
		recvPath:   recvPath,
		sendStream: sendStream,
		recvStream: recvStream,
		sendPipe:   sendPipe,
		recvPipe:   recvPipe,
	}
}

func teardown(ctx *context) {
	err := os.RemoveAll(ctx.prefix)
	if err != nil {
		fmt.Println(err)
	}
	_ = ctx.sendPipe.Close()
	_ = ctx.recvPipe.Close()
}

func assertIsClosed(conn net.Conn, t *testing.T) {
	one := make([]byte, 1)
	err := conn.SetDeadline(time.Now().Add(time.Millisecond * 25))
	if err == io.EOF {
		return // happy
	}
	if _, err := conn.Read(one); err == nil {
		t.Fatal("connection should be disconnected; but it kept reading")
	} else if err != io.EOF && err != io.ErrClosedPipe {
		t.Fatalf("it was expected to fail with to read by EOF (closed) intead it failed with err: %v\n", err)
	}
}

func assertWait(expr func() bool, maxWait time.Duration, t *testing.T) {
	t0 := time.Now()
	for {
		if expr() {
			return
		}
		if time.Now().Sub(t0) > maxWait {
			t.Fatalf("failed: expression did not become true after waiting %v.", maxWait)
		}
		time.Sleep(time.Millisecond * 5)
	}
}
