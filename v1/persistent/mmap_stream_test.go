package persistent

import (
	"fmt"
	"github.com/kuking/go-frank/v1/base"
	"github.com/kuking/go-frank/v1/serialisation"
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

	s, err := MmapStreamCreate(base, 64*1024*1024, &serialisation.ByteArraySerialiser{})
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

func TestSimpleCreateCloseOpenFeedCloseConsumeDelete(t *testing.T) {
	prefix, _ := ioutil.TempDir("", "MMAP-")
	baseDir := prefix + "/a-stream"
	defer cleanup(prefix)

	var s *MmapStream
	var err error
	if s, err = MmapStreamCreate(baseDir, 64*1024, &serialisation.ByteArraySerialiser{}); err != nil {
		t.Fatal()
	}
	if err = s.CloseFile(); err != nil {
		t.Fatal()
	}

	s, err = MmapStreamOpen(baseDir, &serialisation.ByteArraySerialiser{})
	for i := 0; i < 20_000; i++ {
		s.Feed([]byte(fmt.Sprintf("!!%v!!%v!!", i, i)))
	}

	subId := s.SubscriberIdForName("sub-1")
	waitduty := base.NewDefaultFastSpinThenWait()
	for i := 0; i < 20_000; i++ {
		val, _, _ := s.PullBySubId(subId, -1, waitduty)
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
