package base

import (
	"fmt"
	"github.com/kuking/go-frank/api"
	"testing"
	"time"
)

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
	go s.Close()

	s = s.FilterNA(func(v interface{}) bool { return v.(int) != 100000 })

	if v := s.Pull(); v.IsPresent() {
		t.Fatal(v)
	}
	if s.CurrAbsPos() != 512 || s.PeekLimit() != 512 || s.Peek(0) != nil {
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

func TestReset(t *testing.T) {
	s := givenInt64ArrayStream(1024)
	for i := 0; i < 512; i++ {
		s.Pull()
	}
	// it can be reset now
	if s.Reset() != 0 {
		t.Fatal()
	}
	for i := 0; i < 1024; i++ {
		s.Pull()
		s.Feed(i)
	}
	// after the ring buffer has loop around, it can not be reset anymore
	if s.Reset() != s.CurrAbsPos() {
		t.Fatal()
	}

	// a closed stream can be reset if previous conditions apply
	s = givenInt64ArrayStream(1024)
	s.AsArray()
	if !s.IsClosed() {
		t.Fatal("this should be closed!")
	}
	if s.Reset() != 0 {
		t.Fatal()
	}

	// a derived (not root) stream can not be reset
	s = givenInt64ArrayStream(1024).MapInt64(func(i int64) int64 { return i + 1 })
	s.Pull()
	if s.Reset() != 0 {
		t.Fatal()
	}
}

func TestWaitApproachInMemory(t *testing.T) {
	s := EmptyStream(1024)
	s.Wait(api.WaitingUpto10ms)
	t0 := time.Now()
	if s.Count() != 0 {
		t.Fatal()
	}
	dur := time.Now().Sub(t0)
	if dur.Milliseconds() < 10 {
		t.Fatal(fmt.Sprintf("it should have waited at least 10ms, but it wait: %v", dur))
	}

	s.Feed("1")
	s.Wait(api.UntilNoMoreData)
	t0 = time.Now()
	if s.Count() != 1 {
		t.Fatal()
	}
	dur = time.Now().Sub(t0)
	if dur.Nanoseconds() > 1_000_000 {
		t.Fatal(fmt.Sprintf("it should have been lots faster, took: %v", dur))
	}
}
