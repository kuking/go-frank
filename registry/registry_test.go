package registry

import (
	"fmt"
	frank "github.com/kuking/go-frank"
	"github.com/kuking/go-frank/serialisation"
	"io/ioutil"
	"os"
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

func TestInMemoryRegistry_Obtain(t *testing.T) {
	prefix, _ := ioutil.TempDir("", "MMAP-")
	base := prefix + "/a-stream"
	defer cleanup(prefix)

	s1 := frank.EmptyStream(123)
	s2, _ := frank.PersistentStream(base, 64*1024, serialisation.ByteArraySerialiser{})
	s2.Feed("record")

	imr := NewInMemoryRegistry()
	_ = imr.Register("one", s1)
	_ = imr.Register("two", s2)

	if res, err := imr.Obtain("one?lala=123"); err != nil || res != s1 {
		t.Fatal(err)
	}
	if res, err := imr.Obtain("two?clientName=lala");
		err != nil || serialisation.AsString(res.First().Get()) != "record" {
		t.Fatal(err)
	}
	if _, err := imr.Obtain("two?missing-client-name=lala"); err == nil {
		t.Fatal(err)
	}
}

func cleanup(prefix string) {
	err := os.RemoveAll(prefix)
	if err != nil {
		fmt.Println(err)
	}
}
