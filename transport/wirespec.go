package transport

const (
	WireVersion byte = 1
	WireHELLO   byte = 1
	WireDESC    byte = 2
	WireACK     byte = 3
	WireNACK1   byte = 4
	WireNACKN   byte = 5
	WireDATA    byte = 6
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

type WireDescriptionMsg struct {
	Version    byte // = WireVersion
	Message    byte // = WireDESC
	PartSize   uint64
	FirstPart  uint64
	PartsCount uint64
	Write      uint64
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
	Length  uint16 // = Length -- from here on, it can be read directly into mmap
	// Data    []byte
}
