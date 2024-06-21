package highlevel

func (s *streamImpl[T]) ForEach(op func(T)) {
	for {
		read, found := s.pull()
		if !found {
			return
		}
		op(read)
	}
}
