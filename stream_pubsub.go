package go_frank

// uri formats
//
// file+text://
// Reads, writes text files, each element is one line; end of line characters are stripped when reading, added when
// writing. Options: append=true (only used in publishers). Useful for reading CSV files, one line JSONs.
//
// file+mm://
// Memory mapped file; multiple producer, multiple consumer stream. Parameters:
// - offset: resets reading offset to the provided element offset
// - subscriberId: uniq id used to remember subscriber offset position
// - extendSize: how big are going to be allowed to be each extend (i.e. 10M) only used when creating a new mm file
//
// A memory mapped files is actually multiple files: i.e. file+mm://events
// - events.001, events.001, events.002 are the file extends
// - events.idx offset indexes for fast skipping
// - events.cli clients states
//
// concurrency warranties are valid as far as the underlying operating system honours the memory...
// https://github.com/edsrzf/mmap-go
// https://github.com/riobard/go-mmap
//
// mem://name
// In memory stream, behaves similarly as EmptyStream(n int). Useful for parametrised initialisation. If a name is
// provided, it will be registered and shared between different instances. Parameters:
// - size: number of elements to hold
// - multicast: boolean, enables multicasting in pub-sub.
//
// udp+recv://bind_host:listen_port
// Each UDP message received will be inserted in the stream as one element of []byte content, insecure, unauthenticated.
//
// udp+send://host:port
// Each stream element is sent to the provided host:port as an unique udp message, the payload is encoded using the
// binary package in little endian. If another encoding is required, the element should be converted to []byte before
// publishing.
//
// tcp://bind_host:bind_port
// Networked stream subscription with capacity to replay, re-subscribe and hold client states. Parameters like file+mm,
// offset, subscriberId are supported.
//

func Subscribe(uri string) Stream {
	return nil
}

func (s *streamImpl) Publish(uri string) { //Publisher struct

}
