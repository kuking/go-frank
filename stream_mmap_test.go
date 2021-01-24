package go_frank

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func testSimpleCreateOpenFeedDelete(t *testing.T) {
	prefix, _ := ioutil.TempDir("", "MMAP-")
	//prefix, _ := os.Getwd()
	//prefix += "/TEST"
	base := prefix + "/a-stream"
	defer cleanup(prefix)

	t0 := time.Now()

	s, err := MmapStreamCreate(base, 64*1024*1024, &ByteArraySerialiser{})
	if err != nil {
		t.Fatal(err)
	}
	//gob.Register(map[string]string{})
	//value := map[string]string{"a": "b"}
	value := []byte("Hello, how are you doing?")

	count := 100 * 1024 * 1024
	for i := 0; i < count; i++ {
		if i%1_000_000 == 0 {
			_, _ = os.Stdout.WriteString(".")
			_ = os.Stdout.Sync()
		}
		s.Feed(value)
		//s.Feed(i)
	}
	fmt.Println()

	dt := time.Now().Sub(t0)
	rt := time.Duration(int64(dt) / int64(count))
	iops := 1_000_000_000 / rt.Nanoseconds()
	fmt.Printf("Took: %v to store %v MElems, avg. %v/write, %v K.IOPS, %v Mb, %v Mb/s\n",
		dt.Truncate(time.Millisecond), count/1024/1024, rt, iops/1024, s.descriptor.Write/1024/1024,
		s.descriptor.Write/1024/uint64(dt.Milliseconds()))

	if err = s.CloseFile(); err != nil {
		t.Fatal()
	}

	// Took: 29.527s to store 10 MElems, avg. 2.815µs/write, 355239 IOPS, 470 Mb. (gob serialiser, intel)
	// Took: 975ms to store 10 MElems, avg. 93ns/write, 10752688 IOPS, 320 Mb. (ByteArraySerialiser. amd)
	// Took: 9.354s to store 100 MElems, avg. 89ns/write, 10972 K.IOPS, 3300 Mb, 361 Mb/s (ByteArraySerialiser, intel)

}

func TestMmapStreamSubscriberForID(t *testing.T) {
	prefix, _ := ioutil.TempDir("", "MMAP-")
	base := prefix + "/a-stream"
	defer cleanup(prefix)

	s, err := MmapStreamCreate(base, 1024*1024*1024, &ByteArraySerialiser{})
	if err != nil {
		t.Fatal(err)
	}

	ids := map[int]int{}
	for i := 0; i < 1000; i++ {
		id := s.SubscriberIdForName(fmt.Sprint(i))
		if id < 0 || id > 64 {
			t.Fatal("id should be between 0 and 64")
		}
		ids[id]++
	}

	for _, v := range ids {
		if v < 2 {
			t.Fatal("it is expected to recycle old sub-ids")
		}
	}

	if s.SubscriberIdForName("HELLO") != s.SubscriberIdForName("HELLO") {
		t.Fatal("for the same named-subscriber, the subId should be the same")
	}
}

func TestSimpleCreateCloseOpenFeedCloseConsumeDelete(t *testing.T) {
	prefix, _ := ioutil.TempDir("", "MMAP-")
	base := prefix + "/a-stream"
	defer cleanup(prefix)

	var s *mmapStream
	var err error
	if s, err = MmapStreamCreate(base, 64*1024, &ByteArraySerialiser{}); err != nil {
		t.Fatal()
	}
	if err = s.CloseFile(); err != nil {
		t.Fatal()
	}

	s, err = MmapStreamOpen(base, &ByteArraySerialiser{})
	for i := 0; i < 20_000; i++ {
		s.Feed([]byte(fmt.Sprintf("!!%v!!%v!!", i, i)))
	}

	subId := s.SubscriberIdForName("sub-1")
	for i := 0; i < 20_000; i++ {
		val := s.pullBySubId(subId)
		if fmt.Sprintf("!!%v!!%v!!", i, i) != string(val.([]byte)) {
			t.Fatal(fmt.Sprintf("%v should be eq to %v", val, i))
		}
	}

	if err = s.CloseFile(); err != nil {
		t.Fatal(err)
	}
}

func cleanup(prefix string) {
	err := os.RemoveAll(prefix)
	if err != nil {
		fmt.Println(err)
	}
}
