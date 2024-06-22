package transport

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"testing"
	"time"
)

func TestWireDataMsgNA(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	wireDataMsgNA := WireDataMsgNA{}
	for i := 0; i < len(wireDataMsgNA); i++ {
		wireDataMsgNA[i] = byte(rand.Int())
	}
	var wireDataMsg WireDataMsg
	if err := binary.Read(bytes.NewBuffer(wireDataMsgNA[:]), binary.LittleEndian, &wireDataMsg); err != nil {
		t.Fatal()
	}
	if wireDataMsg.Version != wireDataMsgNA.Version() ||
		wireDataMsg.Message != wireDataMsgNA.Message() ||
		wireDataMsg.AbsPos != wireDataMsgNA.AbsPos() ||
		wireDataMsg.Length != wireDataMsgNA.Length() {
		t.Fatal()
	}

	wireDataMsgNA.SetVersion(byte(rand.Int()))
	wireDataMsgNA.SetMessage(byte(rand.Int()))
	wireDataMsgNA.SetAbsPos(rand.Uint64())
	wireDataMsgNA.SetLength(uint16(rand.Int()))

	if err := binary.Read(bytes.NewBuffer(wireDataMsgNA[:]), binary.LittleEndian, &wireDataMsg); err != nil {
		t.Fatal()
	}
	if wireDataMsg.Version != wireDataMsgNA.Version() ||
		wireDataMsg.Message != wireDataMsgNA.Message() ||
		wireDataMsg.AbsPos != wireDataMsgNA.AbsPos() ||
		wireDataMsg.Length != wireDataMsgNA.Length() {
		t.Fatal()
	}

}
