package base

import (
	"github.com/kuking/go-frank/api"
	"github.com/kuking/go-frank/registry"
)

// uri formats
//
// file+text://
// Reads, writes text files, each element is one line; end of line characters are stripped when reading, added when
// writing. Options: append=true (only used in publishers). Useful for reading CSV files, one line JSONs.
//
// file+frank://
// A Frank memory mapped file; multiple producer, multiple consumer stream. Parameters:
// - subscriberId=string: uniq id used to continue, track offset position in the stream
// - reset=boolean: resets the stream position to its earliest possible
// - partSize=megabytes: how big are going to be each part file, only configurable at creation time
//
// concurrency warranties are valid as far as the underlying operating system honours sharing memory buffers.
//
// mem://name
// In memory stream, behaves similarly as EmptyStream(n int). Useful for parametrised initialisation. If a name is
// provided, it will be registered and shared between different instances. Parameters:
// - size=n: number of elements to hold
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

var LocalRegistry = registry.NewInMemoryRegistry()

func Subscribe(uri string) (api.Stream, error) {
	return LocalRegistry.Obtain(uri)
}

// Consumes the contents of this Stream and publishes it into the provided URI Steam. Consumption will follow the
// Stream' Wait approach; the output Stream will be left open.
func (s *StreamImpl) Publish(uri string) error {
	out, err := LocalRegistry.Obtain(uri)
	if err != nil {
		return err
	}
	go publishSpooler(s, out, false)
	return nil
}

// Consumes the contents of this Stream and publishes it into the provided URI Steam. Consumption will follow the
// Stream' Wait approach; the output Stream will be Closed when no more elements are available in the Stream.
func (s *StreamImpl) PublishClose(uri string) error {
	out, err := LocalRegistry.Obtain(uri)
	if err != nil {
		return err
	}
	go publishSpooler(s, out, true)
	return nil
}

func publishSpooler(source, sink api.Stream, closeAtEnd bool) {
	source.ForEach(func(elem interface{}) { sink.Feed(elem) })
	if closeAtEnd {
		sink.Close()
	}
}
