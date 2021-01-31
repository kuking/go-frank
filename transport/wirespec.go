package transport

const (
	WireVersion byte = 1
	WireHELLO   byte = 1
	WireACK     byte = 2
	WireNACK1   byte = 3
	WireNACKN   byte = 4
	WireDATA    byte = 5
)

// One wire communication (UDP/TCP) is established per replication link, all structs are sent in little endian

// Replication flow
// ORIGIN          ->       REPLICA
// send/recv    WireHELLO   send/recv - Any party can initiate a replication, any party can hang after first this
// recv         WireNACKN   send      - Replica indicates from where to start to receive
// recv         WireNACK1   send      - Replica might indicate to retransmit only one element
// send         WireDATA    recv      - Replica receives one payload
// recv         WireACK     send      - Replica informs Origin his confirmed high-water-mark

// WireHELLO
type WireHelloMsg struct {
	Version      byte // = WireVersion
	Message      byte // = WireHELLO
	Intention    byte // 0=Pull copy, 1=Push copy
	StreamUniqId uint64
}

type WireAcksMsg struct {
	Version byte // = WireVersion
	Message byte // = WireACK / WireNACK1 / WireNACKN
	AbsPos  uint64
}

type WireDataMsg struct {
	Version byte   // = WireVersion
	Message byte   // = WireDATA
	AbsPos  uint64 // = Message
	Length  uint64 // = Length -- from here on, it can be read directly into mmap
	// Data    []byte
}
