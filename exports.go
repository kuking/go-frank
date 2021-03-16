package go_frank

import (
	"bufio"
	"github.com/kuking/go-frank/api"
	"github.com/kuking/go-frank/base"
	"github.com/kuking/go-frank/persistent"
	"github.com/kuking/go-frank/serialisation"
	"os"
)

// Creates am empty api.Stream with the required capacity in its ring buffer; the api.Stream is not cosed. If used directly with
// a termination function, it will block waiting for the Closing signal. This constructor is meant to be used in a
// multithreading consumer/producer scenarios, not for simple consumers i.e. e. an array of elements (use Arrayapi.Stream)
// or for creating a api.Stream with a generator function (see api.StreamGenerator). Default waiting time-out is UntilClosed.
func EmptyStream(capacity int) api.Stream {
	return base.EmptyStream(capacity)
}

func ArrayStream(elems interface{}) api.Stream {
	return base.ArrayStream(elems)
}

func StreamGenerator(generator func() api.Optional) api.Stream {
	return base.StreamGenerator(generator)
}

func TextFileStream(filename string) api.Stream {
	file, err := os.Open(filename)
	if err != nil {
		return EmptyStream(1)
	}
	return TextOsFileStream(file)
}

func TextOsFileStream(file *os.File) api.Stream {
	scanner := bufio.NewScanner(file)
	return base.StreamGenerator(func() api.Optional {
		if scanner.Scan() {
			return api.OptionalOf(scanner.Text())
		}
		_ = file.Close()
		return api.EmptyOptional()
	})
}

func PersistentStream(basePath string, partSize uint64, serialiser serialisation.StreamSerialiser) (api.PersistentStream, error) {
	return persistent.OpenCreatePersistentStream(basePath, partSize, serialiser)
}

func Subscribe(uri string) (api.Stream, error) {
	return base.Subscribe(uri)
}

func SubscribeNE(uri string) api.Stream {
	s, err := base.Subscribe(uri)
	if err != nil {
		s = EmptyStream(0)
		s.Close()
	}
	return s
}

func Register(uri string, stream interface{}) error {
	return base.LocalRegistry.Register(uri, stream)
}
