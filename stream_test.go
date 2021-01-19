package go_frank

import (
	"testing"
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
	result := givenArrayStream().AsArray()
	if len(result) != 6 || result[2] != "are" {
		t.Fatal()
	}
}

func TestReduce(t *testing.T) {
	result, ok := givenArrayStream().
		Reduce(func(l, r string) string { return l + " " + r }).
		First()
	if !ok || result != "Hello how are you doing ?" {
		t.Fatal()
	}
}

func TestFirstLast(t *testing.T) {
	result, ok := givenArrayStream().First()
	if !ok || result != "Hello" {
		t.Fatal()
	}
	result, ok = givenArrayStream().Last()
	if !ok || result != "?" {
		t.Fatal()
	}
}

func TestCount(t *testing.T) {
	count := givenArrayStream().Count()
	if count != 6 {
		t.Fatal()
	}
}

func TestMap(t *testing.T) {
	result, ok := givenArrayStream().
		Map(func(s string) int { return len(s) }).
		Reduce(func(l, r int) int { return l + r }).
		First()
	if !ok || result != 20 {
		t.Fatal()
	}
}

func TestSumInt64(t *testing.T) {
	result, ok := givenArrayStream().
		Map(func(s string) int64 { return int64(len(s)) }).
		SumInt64().
		First()
	if !ok || result != int64(20) {
		t.Fatal()
	}
}

func TestFilter(t *testing.T) {
	result, ok := givenArrayStream().
		Filter(func(s string) bool { return len(s) == 3 }).
		Reduce(func(l, r string) string { return l + " " + r }).
		First()
	if !ok || result != "Hello doing ?" {
		t.Fatal(result)
	}
}

func givenArrayStream() *streamImpl {
	return ArrayStream([]interface{}{"Hello", "how", "are", "you", "doing", "?"})
}
