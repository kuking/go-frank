package persistent

import (
	"fmt"
	"github.com/kuking/go-frank/serialisation"
	"io/ioutil"
	"testing"
)

func TestMmapStreamSubscriberForID(t *testing.T) {
	prefix, _ := ioutil.TempDir("", "MMAP-")
	base := prefix + "/a-stream"
	defer cleanup(prefix)

	s, err := MmapStreamCreate(base, 1024*1024*1024, &serialisation.ByteArraySerialiser{})
	if err != nil {
		t.Fatal(err)
	}

	ids := map[int]int{}
	for i := 0; i < 1000; i++ {
		id := s.SubscriberIdForName(fmt.Sprint(i))
		if id < 0 || id > 64 {
			t.Fatal("id should be between 0 and 64")
		}
		if serialisation.FromNTString(s.descriptor.SubName[id][:]) != fmt.Sprint(i) {
			t.Fatal()
		}
		ids[id]++
	}

	for _, v := range ids {
		if v < 2 {
			t.Fatal("it is expected to recycle old sub-ids")
		}
	}

	if s.SubscriberIdForName("123") != s.SubscriberIdForName("123") {
		t.Fatal()
	}
	idx := s.SubscriberIdForName("123")
	if serialisation.FromNTString(s.descriptor.SubName[idx][:]) != "123" {
		t.Fatal()
	}

	if s.SubscriberIdForName("234") == s.SubscriberIdForName("345") {
		t.Fatal()
	}

	if s.SubscriberIdForName("HELLO") != s.SubscriberIdForName("HELLO") {
		t.Fatal("for the same named-subscriber, the subId should be the same")
	}
}

func TestMmapStream_GetRepSubIds(t *testing.T) {
	prefix, _ := ioutil.TempDir("", "MMAP-")
	base := prefix + "/a-stream"
	defer cleanup(prefix)

	s, _ := MmapStreamCreate(base, 1024*1024*1024, &serialisation.ByteArraySerialiser{})

	if len(s.GetReplicatorIds()) != 0 {
		t.Fatal()
	}

	repId, subId, created := s.ReplicatorIdForNameHost("hello", "a-host")
	if repId == -1 || subId == -1 || !created {
		t.Fatal()
	}

	repId2, subId2, created2 := s.ReplicatorIdForNameHost("hello", "another-host")
	if repId != repId2 || subId != subId2 || created2 {
		t.Fatal()
	}

}
