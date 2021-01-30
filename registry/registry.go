package registry

import (
	"errors"
	"fmt"
	"github.com/kuking/go-frank/api"
	"net/url"
	"reflect"
	"sync"
)

type Registry interface {
	Register(uri string, stream interface{}) error
	Unregister(uri string)
	Obtain(uri string) (api.Stream, error)
	List() []string
}

type InMemoryRegistry struct {
	lock     sync.Mutex
	registry map[string]interface{}
}

func NewInMemoryRegistry() Registry {
	return &InMemoryRegistry{
		lock:     sync.Mutex{},
		registry: map[string]interface{}{},
	}
}

func (i *InMemoryRegistry) Register(uri string, stream interface{}) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	// parse uri here
	if i.registry[uri] != nil {
		return errors.New(fmt.Sprintf("registry already has a stream with name: %v", uri))
	}
	i.registry[uri] = stream
	return nil
}

func (i *InMemoryRegistry) Unregister(uri string) {
	i.lock.Lock()
	defer i.lock.Unlock()
	delete(i.registry, uri)
}

func (i *InMemoryRegistry) Obtain(uri string) (api.Stream, error) {

	URI, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	name := URI.Path
	stream, ok := i.registry[name]
	if !ok {
		return nil, errors.New(fmt.Sprintf("registry does not have a stream with name: %v", name))
	}
	switch stream.(type) {
	case api.Stream:
		return stream.(api.Stream), nil
	case api.PersistentStream:
		persistent := stream.(api.PersistentStream)
		clientName := URI.Query().Get("clientName")
		if clientName == "" {
			return nil, errors.New(fmt.Sprintf("persistent stream needs a 'clientName' parameter"))
		}
		return persistent.Consume(clientName), nil
	default:
		return nil, errors.New(fmt.Sprintf("registry does not understand streams of type: %v", reflect.TypeOf(stream)))
	}
}

func (i *InMemoryRegistry) List() []string {
	i.lock.Lock()
	defer i.lock.Unlock()
	res := make([]string, 0)
	for k, _ := range i.registry {
		res = append(res, k)
	}
	return res
}
