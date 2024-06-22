package base

import (
	"testing"
)

func TestSkip(t *testing.T) {
	if givenInt64StreamGenerator(1024).Skip(512).Count() != 512 {
		t.Fatal()
	}
	if givenInt64StreamGenerator(1024).Skip(1024).Count() != 0 {
		t.Fatal()
	}
	arr := givenInt64StreamGenerator(2048).Skip(1024).AsArray()
	for i := 0; i < 1024; i++ {
		if arr[i] != int64(i)+1024+1 { // +1 generator is 1 based
			t.Fatal()
		}
	}
}
