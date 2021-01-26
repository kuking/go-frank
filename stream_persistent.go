package go_frank

func (s *mmapStream) Consume(clientName string) Stream {
	subId := s.SubscriberIdForName(clientName)
	provider := &mmapStreamProviderForSubscriber{
		subId:        subId,
		waitApproach: UntilNoMoreData,
		mmapStream:   s,
	}
	pullFn := func() (read interface{}, closed bool) {
		return s.pullBySubId(subId, provider.waitApproach)
	}
	return &streamImpl{
		provider: provider,
		pull:     pullFn,
	}
}

type mmapStreamProviderForSubscriber struct {
	subId        int
	waitApproach WaitApproach
	mmapStream   *mmapStream
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
	return ms.mmapStream.pullBySubId(ms.subId, ms.waitApproach)
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

func (ms *mmapStreamProviderForSubscriber) Wait(waitApproach WaitApproach) {
	ms.waitApproach = waitApproach
}
