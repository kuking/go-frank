package go_frank

func (s *mmapStream) Consume(clientName string) Stream {
	subId := s.SubscriberIdForName(clientName)
	fn := func() (read interface{}, closed bool) {
		return s.pullBySubId(subId)
	}
	si := streamImpl{
		provider: &mmapStreamProviderSubscriberAware{
			subId:      subId,
			mmapStream: s,
		},
		pull: fn,
	}
	return &si
}

type mmapStreamProviderSubscriberAware struct {
	subId      int
	mmapStream *mmapStream
}

func (ms *mmapStreamProviderSubscriberAware) Feed(elem interface{}) {
	ms.mmapStream.Feed(elem)
}
func (ms *mmapStreamProviderSubscriberAware) Close() {
	ms.mmapStream.Close()
}

func (ms *mmapStreamProviderSubscriberAware) IsClosed() bool {
	return ms.mmapStream.IsClosed()
}

func (ms *mmapStreamProviderSubscriberAware) Pull() (elem interface{}, closed bool) {
	return ms.mmapStream.Pull()
}
func (ms *mmapStreamProviderSubscriberAware) Reset() uint64 {
	return ms.mmapStream.Reset(ms.subId)
}

func (ms *mmapStreamProviderSubscriberAware) CurrAbsPos() uint64 {
	//XXX: implement
	return 0
}

func (ms *mmapStreamProviderSubscriberAware) PeekLimit() uint64 {
	//XXX: implement
	return 0
}

func (ms *mmapStreamProviderSubscriberAware) Peek(absPos uint64) interface{} {
	//XXX: implement
	return nil
}
