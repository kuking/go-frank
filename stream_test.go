package go_frank

import (
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
	result := givenStringArrayStream().AsArray()
	if len(result) != 6 || result[2] != "are" {
		t.Fatal()
	}
}

func TestGenerator(t *testing.T) {
	count := 0
	stream := StreamGenerator(func() Optional {
		count++
		if count <= 10*1024 {
			return OptionalOf(count)
		}
		return EmptyOptional()
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

func givenInt64StreamGenerator(total int) Stream {
	count := int64(0)
	return StreamGenerator(func() Optional {
		count++
		if count <= int64(total) {
			return OptionalOf(count)
		}
		return EmptyOptional()
	})
}

func givenIntArray(count int) []interface{} {
	elems := make([]interface{}, count)
	for i := 0; i < len(elems); i++ {
		elems[i] = i
	}
	return elems
}

func givenInt64Array(count int) []interface{} {
	elems := make([]interface{}, count)
	for i := 0; i < len(elems); i++ {
		elems[i] = int64(i)
	}
	return elems
}

func givenStringArrayStream() Stream {
	return ArrayStream([]interface{}{"Hello", "how", "are", "you", "doing", "?"})
}

func givenInt64ArrayStream(l int) Stream {
	arr := make([]interface{}, l)
	for i := 0; i < l; i++ {
		arr[i] = int64(i)
	}
	return ArrayStream(arr)
}
