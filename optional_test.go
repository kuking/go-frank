package go_frank

import (
	"fmt"
	"testing"
)

func TestOptionalEmpty(t *testing.T) {
	opt := EmptyOptional()
	if opt.IsPresent() || !opt.isEmpty() {
		t.Fatal()
	}
	if fmt.Sprint(opt.Get()) != "tried to Get an empty option" {
		t.Fatal()
	}
	ok := false
	opt.IfPresent(func() { ok = true })
	if ok {
		t.Fatal()
	}
	if opt.OrElse("lele") != "lele" {
		t.Fatal()
	}
	if opt.Map(func(s string) string { return s + "postfix" }).IsPresent() {
		t.Fatal()
	}
}

func TestOptionalOf(t *testing.T) {
	opt := OptionalOf("hello")
	if !opt.IsPresent() || opt.isEmpty() {
		t.Fatal()
	}
	if opt.Get() != "hello" {
		t.Fatal()
	}
	ok := false
	opt.IfPresent(func(_ interface{}) { ok = true })
	if !ok {
		t.Fatal()
	}
	if opt.OrElse("lele") != "hello" {
		t.Fatal()
	}
	if opt.Map(func(s string) string { return s + "postfix" }).Get() != "hellopostfix" {
		t.Fatal()
	}
}
