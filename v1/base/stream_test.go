package base

import (
	api2 "github.com/kuking/go-frank/v1/api"
	"runtime"
	"testing"
	"time"
)

func TestSimplest(t *testing.T) {
	s := EmptyStream(256)
	s.Feed("Hello")
	go s.Close()

	result := s.AsArray()
	if len(result) != 1 || result[0] != "Hello" {
		t.Fatal()
	}
}

func TestSimple(t *testing.T) {
	result := GivenStringArrayStream().AsArray()
	if len(result) != 6 || result[2] != "are" {
		t.Fatal()
	}
}

func TestGenerator(t *testing.T) {
	count := 0
	stream := StreamGenerator(func() api2.Optional {
		count++
		if count <= 10*1024 {
			return api2.OptionalOf(count)
		}
		return api2.EmptyOptional()
	})
	runtime.Gosched() // to let the ring-buffer fill, so coverage
	time.Sleep(1 * time.Millisecond)
	runtime.Gosched()
	if stream.Count() != 10*1024 {
		t.Fatal()
	}
}

func TestArrayStream(t *testing.T) {
	if ArrayStream([]string{"hi", "hello"}).Count() != 2 {
		t.Fatal()
	}
	if ArrayStream("hi").Count() != 1 {
		t.Fatal()
	}
	if !ArrayStream(nil).IsEmpty() {
		t.Fatal()
	}
}

// -------------------------------------------------------------------------------------------------------------------

func GivenStringArrayStream() api2.Stream {
	return ArrayStream([]interface{}{"Hello", "how", "are", "you", "doing", "?"})
}
