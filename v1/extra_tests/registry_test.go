package extra_tests

import (
	"fmt"
	frank "github.com/kuking/go-frank/v1"
	"github.com/kuking/go-frank/v1/registry"
	"github.com/kuking/go-frank/v1/serialisation"
	"io/ioutil"
	"os"
	"testing"
)

func TestInMemoryRegistry_Obtain(t *testing.T) {
	prefix, _ := ioutil.TempDir("", "MMAP-")
	base := prefix + "/a-stream"
	defer cleanup(prefix)

	s1 := frank.EmptyStream(123)
	s2, _ := frank.PersistentStream(base, 64*1024, serialisation.ByteArraySerialiser{})
	s2.Feed("record")

	imr := registry.NewInMemoryRegistry()
	_ = imr.Register("one", s1)
	_ = imr.Register("two", s2)

	if res, err := imr.Obtain("one?lala=123"); err != nil || res != s1 {
		t.Fatal(err)
	}
	if res, err := imr.Obtain("two?sn=lala"); err != nil || serialisation.AsString(res.First().Get()) != "record" {
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
