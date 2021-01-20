package go_frank

import "testing"

func TestPositioningAtStart(t *testing.T) {
	s := EmptyStream(1024)
	if s.CurrAbsPos() != uint64(0) {
		t.Fatal()
	}
	if s.PeekLimit() != uint64(0) {
		t.Fatal()
	}
	if s.Peek(0) != nil || s.Peek(123) != nil {
		t.Fatal()
	}
}

func TestPositioningSimple(t *testing.T) {
	s := EmptyStream(1024)
	for i := 0; i < 512; i++ {
		s.Feed(i)
	}
	if s.CurrAbsPos() != uint64(0) {
		t.Fatal()
	}
	if s.PeekLimit() != uint64(512) {
		t.Fatal()
	}
	for i := 0; i < int(s.PeekLimit()); i++ {
		if s.Peek(uint64(i)) != i {
			t.Fatal(i)
		}
	}
	if s.Peek(512) != nil {
		t.Fatal()
	}
}

func TestPositioningOnNonHeads(t *testing.T) {
	// positioning operations on non-head Stream should fail
	s := EmptyStream(1024)
	for i := 0; i < 512; i++ {
		s.Feed(i)
	}
	s = s.FilterNA(func(v interface{}) bool { return v.(int) == 100000 })

	if s.Pull().isPresent() {
		t.Fatal()
	}
	if s.CurrAbsPos() != 0 || s.PeekLimit() != 0 || s.Peek(0) != nil {
		t.Fatal()
	}
}

func TestPositionOnRollOverStream(t *testing.T) {
	s := EmptyStream(1024)
	for i := 0; i < 1024+1024+512; i++ {
		s.Feed(i)
		if i >= 512 {
			s.Pull()
		}
	}
	if s.CurrAbsPos() != uint64(2048) {
		t.Fatal()
	}
	if s.PeekLimit() != uint64(2048+512) {
		t.Fatal()
	}
	for i := s.CurrAbsPos(); i < s.PeekLimit(); i++ {
		if s.Peek(i) != int(i) {
			t.Fatal(i)
		}
	}
}
