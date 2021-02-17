package transport

import (
	"encoding/binary"
	"io"
)

// Non Allocation versions of the wire structs, at the moment only WireDataMsg is implemented as it is the only one used
// many thousand of times per second, all the other message types are once a second-ish sent, no need to optimised.
// Benchmark_WireDataMessage_BinaryWrite = 130ns (+1 alloc) Benchmark_WireDataMessageNA_OwnWriter=6.4ns (no alloc)
type WireDataMsgNA [1 + 1 + 8 + 2]byte

func (m *WireDataMsgNA) Version() byte {
	return m[0]
}

func (m *WireDataMsgNA) SetVersion(v byte) {
	m[0] = v
}

func (m *WireDataMsgNA) Message() byte {
	return m[1]
}

func (m *WireDataMsgNA) SetMessage(v byte) {
	m[1] = v
}

func (m *WireDataMsgNA) AbsPos() uint64 {
	return binary.LittleEndian.Uint64(m[2:10])
}

func (m *WireDataMsgNA) SetAbsPos(v uint64) {
	binary.LittleEndian.PutUint64(m[2:10], v)
}

func (m *WireDataMsgNA) Length() uint16 {
	return binary.LittleEndian.Uint16(m[10:12])
}

func (m *WireDataMsgNA) SetLength(v uint16) {
	binary.LittleEndian.PutUint16(m[10:12], v)
}

func (m *WireDataMsgNA) Write(w io.Writer) (err error) {
	var n int
	n, err = w.Write(m[:])
	if n != len(m) {
		err = io.ErrShortWrite
	}
	return
}

func (m *WireDataMsgNA) Read(r io.Reader) (err error) {
	var n int
	n, err = io.ReadFull(r, m[:])
	if n != len(m) {
		err = io.ErrShortWrite
	}
	return
}
