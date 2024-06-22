package persistent

import (
	"github.com/kuking/go-frank/v1/api"
	"github.com/kuking/go-frank/v1/base"
	"github.com/kuking/go-frank/v1/serialisation"
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
	waitDuty := base.NewDefaultFastSpinThenWait()
	subId := s.SubscriberIdForName(subscriberName)
	provider := &mmapStreamProviderForSubscriber{
		subId:       subId,
		waitTimeOut: api.UntilNoMoreData,
		waitDuty:    waitDuty,
		mmapStream:  s,
	}
	pullFn := func() (read interface{}, closed bool) {
		read, _, closed = s.PullBySubId(subId, provider.waitTimeOut, provider.waitDuty)
		return
	}
	return base.NewStreamImpl(provider, pullFn)
}

// FIXME
func (s *MmapStream) Publish(uri string) {
	base.LocalRegistry.Register(uri, s)
}

type mmapStreamProviderForSubscriber struct {
	subId       int
	waitTimeOut api.WaitTimeOut
	waitDuty    api.WaitDuty
	mmapStream  *MmapStream
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
	elem, _, closed = ms.mmapStream.PullBySubId(ms.subId, ms.waitTimeOut, ms.waitDuty)
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

func (ms *mmapStreamProviderForSubscriber) WaitTimeOut(waitTimeOut api.WaitTimeOut) {
	ms.waitTimeOut = waitTimeOut
}

func (ms *mmapStreamProviderForSubscriber) WaitDuty(waitDuty api.WaitDuty) {
	ms.waitDuty = waitDuty
}
