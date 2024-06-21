package api

import (
	"fmt"
	"testing"
)

func TestOptionalEmpty(t *testing.T) {
	opt := Empty[string]()
	if opt.IsPresent() || !opt.IsEmpty() {
		t.Fatal()
	}
	if fmt.Sprint(opt.OrElse("not found")) != "not found" {
		t.Fatal()
	}
	ok := false
	opt.IfPresent(func(string) { ok = true })
	if ok {
		t.Fatal()
	}
	if opt.OrElse("lele") != "lele" {
		t.Fatal()
	}
	if opt.Map(func(s Optional[string]) Optional[any] {
		return Value[any](s.OrElse("") + "postfix")
	}).OrElse("") != "postfix" {
		t.Fatal()
	}
}

func TestOptionalOf(t *testing.T) {
	opt := Value("hello")
	if !opt.IsPresent() || opt.IsEmpty() {
		t.Fatal()
	}
	if opt.OrElse("") != "hello" {
		t.Fatal()
	}
	ok := false
	opt.IfPresent(func(_ string) { ok = true })
	if !ok {
		t.Fatal()
	}
	if opt.OrElse("lele") != "hello" {
		t.Fatal()
	}
	if opt.Map(func(s Optional[string]) Optional[any] {
		return Value[any](s.OrElse("") + "postfix")
	}).OrElse("") != "hellopostfix" {
		t.Fatal()
	}
}

func TestOptional_String(t *testing.T) {
	opt := Value(123)
	if fmt.Sprint(opt) != "123" {
		t.Fatal()
	}
	opt = Empty[int]()
	if fmt.Sprint(opt) != "<empty>" {
		t.Fatal()
	}
}
