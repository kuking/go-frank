package base

import (
	"testing"
)

func TestFirstLast(t *testing.T) {
	result := givenStringArrayStream().First()
	if result.IsEmpty() || result.Get() != "Hello" {
		t.Fatal()
	}
	result = givenStringArrayStream().Last()
	if result.IsEmpty() || result.Get() != "?" {
		t.Fatal()
	}
}

func TestCount(t *testing.T) {
	count := givenStringArrayStream().Count()
	if count != 6 {
		t.Fatal()
	}
}

func TestEmptyFirstLast(t *testing.T) {
	if givenInt64ArrayStream(0).Last().IsPresent() {
		t.Fatal()
	}
	if givenInt64ArrayStream(0).First().IsPresent() {
		t.Fatal()
	}
}

func TestIsEmpty(t *testing.T) {
	if !givenInt64ArrayStream(0).IsEmpty() {
		t.Fatal()
	}
	if givenInt64ArrayStream(1).IsEmpty() {
		t.Fatal()
	}
}

func TestAllMatch(t *testing.T) {
	if !givenInt64ArrayStream(10).AllMatch(func(i int64) bool { return i >= 0 }) {
		t.Fatal()
	}
	// empty stream 'matches all'
	if !givenInt64ArrayStream(0).AllMatch(func(i int64) bool { return i >= 0 }) {
		t.Fatal()
	}
}

func TestNoneMatch(t *testing.T) {
	if !givenInt64ArrayStream(10).NoneMatch(func(i int64) bool { return i < 0 }) {
		t.Fatal()
	}
	// empty stream 'is true that none matches'
	if givenInt64ArrayStream(0).NoneMatch(func(i int64) bool { return i < 0 }) {
		t.Fatal()
	}
}

func TestAtLeastOne(t *testing.T) {
	if !givenInt64ArrayStream(10).AtLeastOne(func(i int64) bool { return i > 5 }) {
		t.Fatal()
	}
	if givenInt64ArrayStream(10).AtLeastOne(func(i int64) bool { return i > 50 }) {
		t.Fatal()
	}
	// empty stream 'is true that none matches'
	if givenInt64ArrayStream(0).AtLeastOne(func(i int64) bool { return i > 5 }) {
		t.Fatal()
	}
}

func TestForEach(t *testing.T) {
	var res []string
	ArrayStream(nil).ForEach(func(s string) { res = append(res, s) })
	if len(res) != 0 {
		t.Fatal()
	}
	givenStringArrayStream().ForEach(func(s string) { res = append(res, s) })
	if len(res) != 6 {
		t.Fatal()
	}
}
