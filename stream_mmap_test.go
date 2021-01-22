package go_frank

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestSimpleCreateOpenFeedDelete(t *testing.T) {
	prefix, _ := ioutil.TempDir("", "MMAP-")
	base := prefix + "/lala"
	defer cleanup(prefix)

	s, err := MmapStreamCreate(base, 64*1024*1024, &GobSerialiser{})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 1024*1024; i++ {
		fmt.Println(i)
		s.Feed(i)
	}
	if err = s.CloseFile(); err != nil {
		t.Fatal()
	}
}

func cleanup(base string) {
	err := os.RemoveAll(base + ".*")
	if err != nil {
		fmt.Println(err)
	}
}
