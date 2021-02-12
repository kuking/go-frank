package persistent

import (
	"github.com/kuking/go-frank/api"
	"github.com/kuking/go-frank/base"
	"github.com/kuking/go-frank/serialisation"
)

func OpenCreatePersistentStream(basePath string, partSize uint64, serialiser serialisation.StreamSerialiser) (ps api.PersistentStream, err error) {
	ps, err = MmapStreamOpen(basePath, serialiser)
	if err == nil {
		return
	} else {
		ps, err = MmapStreamCreate(basePath, partSize, serialiser)
		return
	}
}

func (s *MmapStream) Consume(subscriberName string) api.Stream {
	subId := s.SubscriberIdForName(subscriberName)
	provider := &mmapStreamProviderForSubscriber{
		subId:        subId,
		waitApproach: api.UntilNoMoreData,
		mmapStream:   s,
	}
	pullFn := func() (read interface{}, closed bool) {
		read, _, closed = s.PullBySubId(subId, provider.waitApproach)
		return
	}
	return base.NewStreamImpl(provider, pullFn)
}

//FIXME
func (s *MmapStream) Publish(uri string) {
	base.LocalRegistry.Register(uri, s)
}

type mmapStreamProviderForSubscriber struct {
	subId        int
	waitApproach api.WaitApproach
	mmapStream   *MmapStream
}

func (ms *mmapStreamProviderForSubscriber) Feed(elem interface{}) {
	ms.mmapStream.Feed(elem)
}
func (ms *mmapStreamProviderForSubscriber) Close() {
	ms.mmapStream.Close()
}

func (ms *mmapStreamProviderForSubscriber) IsClosed() bool {
	return ms.mmapStream.IsClosed()
}

func (ms *mmapStreamProviderForSubscriber) Pull() (elem interface{}, closed bool) {
	elem, _, closed = ms.mmapStream.PullBySubId(ms.subId, ms.waitApproach)
	return
}

func (ms *mmapStreamProviderForSubscriber) Reset() uint64 {
	return ms.mmapStream.Reset(ms.subId)
}

func (ms *mmapStreamProviderForSubscriber) CurrAbsPos() uint64 {
	//XXX: implement
	return 0
}

func (ms *mmapStreamProviderForSubscriber) PeekLimit() uint64 {
	//XXX: implement
	return 0
}

func (ms *mmapStreamProviderForSubscriber) Peek(absPos uint64) interface{} {
	//XXX: implement
	return nil
}

func (ms *mmapStreamProviderForSubscriber) Wait(waitApproach api.WaitApproach) {
	ms.waitApproach = waitApproach
}
