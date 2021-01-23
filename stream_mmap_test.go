package go_frank

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func tTestSimpleCreateOpenFeedDelete(t *testing.T) {
	prefix, _ := ioutil.TempDir("", "MMAP-")
	//prefix, _ := os.Getwd()
	//prefix += "/TEST"
	base := prefix + "/lala"
	defer cleanup(prefix)

	t0 := time.Now()

	s, err := MmapStreamCreate(base, 64*1024*1024, &ByteArraySerialiser{})
	if err != nil {
		t.Fatal(err)
	}

	//gob.Register(map[string]string{})
	//value := map[string]string{"a": "b"}
	value := []byte("hello how are you doing?")

	count := 10 * 1024 * 1024
	for i := 0; i < count; i++ {
		if i%1000000 == 0 {
			fmt.Println(i)
		}
		s.Feed(value)
		//s.Feed(i)
	}

	dt := time.Now().Sub(t0)
	rt := time.Duration(int64(dt) / int64(count))
	rps := 1_000_000_000 / rt.Nanoseconds()
	fmt.Printf("Took: %v to store %v MElems, avg. %v/write, %v IOPS, %v Mb.\n",
		dt.Truncate(time.Millisecond), count/1024/1024, rt, rps, s.descriptor.Write/1024/1024)

	if err = s.CloseFile(); err != nil {
		t.Fatal()
	}

	// Took: 29.527s to store 10 MElems, avg. 2.815µs/write, 355239 IOPS, 470 Mb. (gob serialiser, intel)
	// Took: 975ms to store 10 MElems, avg. 93ns/write, 10752688 IOPS, 320 Mb. (ByteArraySerialiser. amd)
}

func cleanup(prefix string) {
	err := os.RemoveAll(prefix)
	if err != nil {
		fmt.Println(err)
	}
}
