package registry

import (
	"testing"
)

func TestInMemoryRegistry_Crud(t *testing.T) {
	imr := NewInMemoryRegistry()
	if err := imr.Register("hola", "hola"); err != nil {
		t.Fatal(err)
	}
	if err := imr.Register("hola", "2"); err == nil {
		t.Fatal("it should fail on double entry")
	}
	if l := imr.List(); len(l) != 1 || l[0] != "hola" {
		t.Fatal()
	}
	imr.Unregister("asd") // ok to unregister non-existent things
	if l := imr.List(); len(l) != 1 {
		t.Fatal()
	}
	imr.Unregister("hola")
	if l := imr.List(); len(l) != 0 {
		t.Fatal("it should be empty now")
	}
}

// more tests in extra_tests to avoid circular references
