package transport

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/kuking/go-frank/api"
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

	assertWait("is disconnected", func() bool { return sl.State == DISCONNECTED }, 100*time.Millisecond, t)
	assertClosedConnection(ctx)
}

func TestSyncLink_GoFuncSend_DetectsClosedConnection(t *testing.T) {
	ctx := setup(t)
	defer teardown(ctx)
	sl := ctx.repl.NewSyncLinkSend(ctx.sendPipe, "host:1234", ctx.sendStream, "repl-1")
	_ = ctx.recvPipe.Close()
	go sl.goFuncSend()

	assertWait("is disconnected", func() bool { return sl.State == DISCONNECTED }, 100*time.Millisecond, t)
	assertClosedConnection(ctx)
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

	initialSendHandShakeDone(ctx)
	feedStream(ctx.sendStream, 100)

	prevAbsPos := uint64(0)
	for i := 0; i < 100; i++ {
		wireDataMsg := verifyReceivesDataMessage(i, ctx)
		if wireDataMsg.AbsPos < prevAbsPos {
			t.Fatal("prevAbsPos does not seems to get updated")
		}
		prevAbsPos = wireDataMsg.AbsPos
	}

	closeAndVerify(sl, ctx)
}

func TestSyncLink_GoFuncSend_ProcessesNACKs(t *testing.T) {
	ctx := setup(t)
	defer teardown(ctx)
	sl := ctx.repl.NewSyncLinkSend(ctx.sendPipe, "host:1234", ctx.sendStream, "repl-1")
	go sl.goFuncSend()

	initialSendHandShakeDone(ctx)

	feedStream(ctx.sendStream, 100)
	var replayAbsPos uint64
	for i := 0; i < 100; i++ {
		wireDataMsg := verifyReceivesDataMessage(i, ctx)
		if i == 50 {
			replayAbsPos = wireDataMsg.AbsPos
		}
	}

	wireAcksMsg := WireAcksMsg{
		Version: WireVersion,
		Message: WireNACKN,
		AbsPos:  replayAbsPos,
	}
	if err := binary.Write(ctx.recvPipe, binary.LittleEndian, &wireAcksMsg); err != nil {
		t.Fatal(err)
	}
	// now, 49 messages should be received
	for i := 50; i < 100; i++ {
		verifyReceivesDataMessage(i, ctx)
	}

	closeAndVerify(sl, ctx)
}

func TestSyncLink_GoFuncSend_ProcessesACKs(t *testing.T) {
	ctx := setup(t)
	defer teardown(ctx)
	sl := ctx.repl.NewSyncLinkSend(ctx.sendPipe, "host:1234", ctx.sendStream, "repl-1")
	go sl.goFuncSend()

	initialSendHandShakeDone(ctx)

	var wireDataMsg *WireDataMsg
	feedStream(ctx.sendStream, 100)
	for i := 0; i < 100; i++ {
		wireDataMsg = verifyReceivesDataMessage(i, ctx)
	}

	if wireDataMsg == nil {
		t.Fatal("impossible, but to keep the linter happy")
	}

	// HWM for replicator Id is 0
	assertWait("HWM is Zero", func() bool { return sl.Stream.GetRepHWM(sl.repId) == 0 }, 100*time.Millisecond, t)

	wireAcksMsg := WireAcksMsg{
		Version: WireVersion,
		Message: WireACK,
		AbsPos:  wireDataMsg.AbsPos,
	}
	if err := binary.Write(ctx.recvPipe, binary.LittleEndian, &wireAcksMsg); err != nil {
		t.Fatal(err)
	}

	// HWM for replicator Id it is now wireDataMsg.AbsPos (what we sent in the wireAckMsg)
	assertWait("HWM is Update", func() bool { return sl.Stream.GetRepHWM(sl.repId) == wireDataMsg.AbsPos }, 500*time.Millisecond, t)

	closeAndVerify(sl, ctx)
}

func TestSyncLink_GoFuncRecv_Close(t *testing.T) {
	ctx := setup(t)
	defer teardown(ctx)
	sl := ctx.repl.NewSyncLinkRecv(ctx.recvPipe, "host:1234", ctx.prefix)
	go sl.goFuncRecv()

	sl.Close()

	assertWait("is disconnected", func() bool { return sl.State == DISCONNECTED }, 100*time.Millisecond, t)
	assertClosedConnection(ctx)
}

func TestSyncLink_GoFuncRecv_DetectsClosedConnection(t *testing.T) {
	ctx := setup(t)
	defer teardown(ctx)
	sl := ctx.repl.NewSyncLinkRecv(ctx.recvPipe, "host:1234", ctx.prefix)
	_ = ctx.recvPipe.Close()
	go sl.goFuncRecv()

	assertWait("is disconnected", func() bool { return sl.State == DISCONNECTED }, 100*time.Millisecond, t)
	assertClosedConnection(ctx)
}

func TestSyncLink_GoFuncRecv_ReceivesHelloAndStatus(t *testing.T) {
	ctx := setup(t)
	defer teardown(ctx)
	sl := ctx.repl.NewSyncLinkRecv(ctx.recvPipe, "host:1234", ctx.prefix)
	go sl.goFuncRecv()

	assertWait("is connected", func() bool { return sl.State == CONNECTED }, 100*time.Millisecond, t)

	wireHelloMsg := WireHelloMsg{
		Version:      WireVersion,
		Message:      WireHELLO,
		StreamUniqId: ctx.sendStream.GetUniqId(),
		PartSize:     ctx.sendStream.GetPartSize(),
		FirstPart:    ctx.sendStream.GetFirstPart(),
	}
	if err := binary.Write(ctx.sendPipe, binary.LittleEndian, &wireHelloMsg); err != nil {
		t.Fatal(err)
	}
	assertWait("is pulling", func() bool { return sl.State == PULLING }, 500*time.Millisecond, t)

	if ctx.sendStream.GetUniqId() == sl.Stream.GetUniqId() {
		t.Fatal("source and replica should have differnet uniqIds")
	}
	if ctx.sendStream.GetUniqId() != sl.Stream.GetReplicaOf() {
		t.Fatal("created stream should hold 'replicaOf' the source")
	}

	if sl.Stream.IsClosed() {
		t.Fatal("Stream should not be closed!")
	}
	wireStatusMsg := WireStatusMsg{
		Version:    WireVersion,
		Message:    WireSTATUS,
		FirstPart:  ctx.sendStream.GetFirstPart(),
		PartsCount: ctx.sendStream.GetPartsCount(),
		Closed:     1, // yes purposely closed to tell it happened
	}
	if err := binary.Write(ctx.sendPipe, binary.LittleEndian, &wireStatusMsg); err != nil {
		t.Fatal(err)
	}

	assertWait("stream closed as directed by status", func() bool { return sl.Stream.IsClosed() }, 500*time.Millisecond, t)

	forceCloseAndVerify(sl, ctx)
}

func TestSyncLink_GoFuncRecv_ReceivesData(t *testing.T) {
	ctx := setup(t)
	defer teardown(ctx)
	sl := ctx.repl.NewSyncLinkRecv(ctx.recvPipe, "host:1234", ctx.prefix)
	go sl.goFuncRecv()

	initialRecvHandShakeDone(ctx)

	feedStream(ctx.sendStream, 100) // because it is easy to obtain correct AbsPos, etc.
	subId := ctx.sendStream.SubscriberIdForName("sub1")
	lastAbsPos := uint64(0)
	for {
		elem, absPos, closed := ctx.sendStream.PullBySubId(subId, api.UntilNoMoreData)
		if closed {
			break
		}
		elemB := elem.([]byte)
		wireDataMsg := WireDataMsg{
			Version: WireVersion,
			Message: WireDATA,
			AbsPos:  absPos,
			Length:  uint16(len(elemB)),
		}
		if err := binary.Write(ctx.sendPipe, binary.LittleEndian, &wireDataMsg); err != nil {
			t.Fatal(err)
		}
		if n, err := ctx.sendPipe.Write(elemB); n != len(elemB) || err != nil {
			t.Fatal()
		}
		lastAbsPos = absPos
	}

	assertEqualStreams(ctx.sendStream, sl.Stream, ctx)
	assertWait("write > lastAbsPos", func() bool { return lastAbsPos < sl.Stream.WritePos() }, 500*time.Millisecond, t)

	forceCloseAndVerify(sl, ctx)
}

// -------------------------------------------------------------------------------------------------------------------

func feedStream(stream *persistent.MmapStream, qty int) {
	for i := 0; i < qty; i++ {
		elem := [8]byte{}
		binary.LittleEndian.PutUint64(elem[:], uint64(i))
		stream.Feed(elem[:])
	}
}

func verifyReceivesDataMessage(n int, ctx *context) *WireDataMsg {
	var wireDataMsg WireDataMsg
	buf := make([]byte, 10)
	if err := binary.Read(ctx.recvPipe, binary.LittleEndian, &wireDataMsg); err != nil {
		ctx.t.Fatal(err)
	}
	if wireDataMsg.Version != WireVersion ||
		wireDataMsg.Message != WireDATA ||
		wireDataMsg.Length != 8 { // fixed, it is an uint64
		ctx.t.Fatal("WireDataMsg does not seems correct")
	}
	if n, err := ctx.recvPipe.Read(buf[0:8]); n != 8 || err != nil {
		ctx.t.Fatal(err)
	}
	if binary.LittleEndian.Uint64(buf[0:8]) != uint64(n) {
		ctx.t.Fatal("value is not correct")
	}
	return &wireDataMsg
}

func initialSendHandShakeDone(ctx *context) {
	var wireHelloMsg WireHelloMsg
	if err := binary.Read(ctx.recvPipe, binary.LittleEndian, &wireHelloMsg); err != nil {
		ctx.t.Fatal(err)
	}
	var wireStatusMsg WireStatusMsg
	if err := binary.Read(ctx.recvPipe, binary.LittleEndian, &wireStatusMsg); err != nil {
		ctx.t.Fatal(err)
	}
}

func initialRecvHandShakeDone(ctx *context) {
	wireHelloMsg := WireHelloMsg{
		Version:      WireVersion,
		Message:      WireHELLO,
		StreamUniqId: ctx.sendStream.GetUniqId(),
		PartSize:     ctx.sendStream.GetPartSize(),
		FirstPart:    ctx.sendStream.GetFirstPart(),
	}
	if err := binary.Write(ctx.sendPipe, binary.LittleEndian, &wireHelloMsg); err != nil {
		ctx.t.Fatal(err)
	}
	wireStatusMsg := WireStatusMsg{
		Version:    WireVersion,
		Message:    WireSTATUS,
		FirstPart:  ctx.sendStream.GetFirstPart(),
		PartsCount: ctx.sendStream.GetPartsCount(),
		Closed:     0,
	}
	if err := binary.Write(ctx.sendPipe, binary.LittleEndian, &wireStatusMsg); err != nil {
		ctx.t.Fatal(err)
	}
}

func assertEqualStreams(left *persistent.MmapStream, right *persistent.MmapStream, ctx *context) {
	leftSubId := left.SubscriberIdForName("left-subscriber")
	rightSubId := right.SubscriberIdForName("right-subscriber")
	for {
		leftElem, leftAbsPos, leftClosed := left.PullBySubId(leftSubId, api.UntilNoMoreData)
		rightElem, rightAbsPos, rightClosed := right.PullBySubId(rightSubId, api.UntilNoMoreData)
		if leftClosed && rightClosed {
			return // happy
		}
		if leftClosed || rightClosed {
			ctx.t.Fatal("streams have different quantity of elements")
		}
		if leftAbsPos != rightAbsPos {
			ctx.t.Fatal("streams got un-synchronised with AbsPos")
		}
		if !bytes.Equal(leftElem.([]byte), rightElem.([]byte)) {
			ctx.t.Fatal("elements are not equal")
		}
	}
}

// needed as Buffer Peek blocks
func forceCloseAndVerify(sl *SyncLink, ctx *context) {
	_ = sl.conn.Close()
	closeAndVerify(sl, ctx)
}

func closeAndVerify(sl *SyncLink, ctx *context) {
	sl.Close()
	assertWait("is disconnected", func() bool { return sl.State == DISCONNECTED }, 500*time.Millisecond, ctx.t)
	assertClosedConnection(ctx)
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

func assertClosedConnection(ctx *context) {
	assertIsClosed(ctx.sendPipe, ctx.t)
	assertIsClosed(ctx.recvPipe, ctx.t)
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

func assertWait(explanation string, expr func() bool, maxWait time.Duration, t *testing.T) {
	t0 := time.Now()
	for {
		if expr() {
			return
		}
		if time.Now().Sub(t0) > maxWait {
			t.Fatalf("failed waiting for: %v (for %v).", explanation, maxWait)
		}
		time.Sleep(time.Millisecond * 5)
	}
}
