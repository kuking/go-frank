package go_frank

func (p *mmapStream) Consume(clientName string) Stream {
	subId := p.SubscriberIdForName(clientName)
	fn := func(s *streamImpl) (read interface{}, closed bool) {
		return p.pullBySubId(subId)
	}
	s := streamImpl{
		ringBuffer: nil,
		ringRead:   0,
		ringWrite:  0,
		ringWAlloc: 0,
		closedFlag: 0,
		pull:       fn,
		prev:       nil,
	}

	return &s

}
