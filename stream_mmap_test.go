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
	//prefix, _ = os.Getwd()
	//prefix += "/TEST"
	base := prefix + "/lala"
	defer cleanup(prefix)

	t0 := time.Now()

	s, err := MmapStreamCreate(base, 16*1024*1024, &GobSerialiser{})
	if err != nil {
		t.Fatal(err)
	}

	count := 10 * 1024 * 1024
	for i := 0; i < count; i++ {
		if i%1000000 == 0 {
			fmt.Println(i)
		}
		s.Feed(i)
	}

	dt := time.Now().Sub(t0)
	fmt.Printf("Took: %v to store %v Krecords, avg. %v/record, %v Kb.\n",
		dt, count/1024, time.Duration(int64(dt)/int64(count)), s.descriptor.Write/1024)

	if err = s.CloseFile(); err != nil {
		t.Fatal()
	}

	// Took: 27.843069823s to store 10240 Krecords, avg. 2.655µs/record, 227296 Kb.
}

func cleanup(prefix string) {
	err := os.RemoveAll(prefix)
	if err != nil {
		fmt.Println(err)
	}
}
