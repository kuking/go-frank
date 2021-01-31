package persistent

const (
	mmapStreamFileVersion    uint64 = 1
	mmapStreamMaxClients     int    = 64
	mmapStreamMaxReplicators int    = 16
	mmapStreamHeaderSize     int    = 2048
	mmapPartHeaderSize       int    = 1024

	// Entry Header
	//  1 Byte  = EndOfPart | Valid | SkipToNext
	//  1 Byte  = Version (1)
	// 2 bytes  = little endian uint16 length of payload (yes, maximum 64kb)
	// variable = payload
	entryHeaderSize int  = 1 + 1 + 4
	entryVersion    byte = 1
	entryIsEoP      byte = 0x11
	entryIsValid    byte = 0x22
	entrySkip       byte = 0x33 // mark as 'this will never be complete' after certain timeout
)

// In memory structure
type mmapStreamDescriptor struct {
	Version    uint64
	UniqId     uint64
	PartSize   uint64
	FirstPart  uint64 // to enable infinite streams, to cleanup old parts
	PartsCount uint64
	Write      uint64
	Closed     uint32

	// 64 persistent subscribers
	SubId   [mmapStreamMaxClients]uint64 // an unique id
	SubRPos [mmapStreamMaxClients]uint64
	SubTime [mmapStreamMaxClients]int64 // last time a subscriber was active (reading/writing), updated rarely but helps to cleanup

	// replicators state, replicator for UniqId 'X' is the subscriber: 'Replicator:X'
	RepUniqId [mmapStreamMaxReplicators]uint64
	RepHWMPos [mmapStreamMaxReplicators]uint64
	RepHost   [mmapStreamMaxReplicators][128]byte
}

// In memory structure
type mmapPartFileDescriptor struct {
	Version uint64
	UniqId  uint64 // same for descriptor and parts
	PartNo  uint64
	ElemOfs [32]uint64
	ElemNo  [32]uint64
	// All offsets are global, first 1kb of the part file contains the header and padding zeroes for simplicity.
	// ElemOfs/ElemNo form an pseudo-index, so the part file can partially seek without sequential reading.
	// by definition ElemNo[0] is the first element in the part file, ElemNo[31] is the last one. And the ones in
	// between are spread equally.
}
