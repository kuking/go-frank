package go_frank

import (
	"fmt"
	"testing"
)

func TestOptionalEmpty(t *testing.T) {
	opt := EmptyOptional()
	if opt.isPresent() || !opt.isEmpty() {
		t.Fatal()
	}
	if fmt.Sprint(opt.Get()) != "tried to Get an empty option" {
		t.Fatal()
	}
	ok := false
	opt.ifPresent(func() { ok = true })
	if ok {
		t.Fatal()
	}
	if opt.orElse("lele") != "lele" {
		t.Fatal()
	}
	if opt.Map(func(s string) string { return s + "postfix" }).isPresent() {
		t.Fatal()
	}
}

func TestOptionalOf(t *testing.T) {
	opt := OptionalOf("hello")
	if !opt.isPresent() || opt.isEmpty() {
		t.Fatal()
	}
	if opt.Get() != "hello" {
		t.Fatal()
	}
	ok := false
	opt.ifPresent(func(_ interface{}) { ok = true })
	if !ok {
		t.Fatal()
	}
	if opt.orElse("lele") != "hello" {
		t.Fatal()
	}
	if opt.Map(func(s string) string { return s + "postfix" }).Get() != "hellopostfix" {
		t.Fatal()
	}
}
