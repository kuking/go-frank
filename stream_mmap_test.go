package go_frank

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"unsafe"
)

func TestMmap(t *testing.T) {
	if err := mmapInit("a-file", 1024*1024); err != nil {
		t.Fatal(err)
	}
	mm, err := mmapOpen("a-file")
	if err != nil {
		t.Fatal(err)
	}

	descriptor := mmapStreamDescriptor{
		Version:    mmapStreamFileVersion,
		PartSize:   1024 * 1024 * 1024,
		PartsCount: 1,
		Write:      0,
		WAlloc:     0,
		Closed:     0,
		SubId:      [64]uint64{},
		SubRPos:    [64]uint64{},
	}

	buf := bytes.NewBuffer(mm)
	buf.Reset()
	mmWriter := bufio.NewWriter(buf)
	//if err = binary.Write(mmWriter, binary.LittleEndian, &descriptor); err != nil {
	//	t.Fatal(err)
	//}

	enc := gob.NewEncoder(mmWriter)
	if err = enc.Encode(descriptor); err != nil {
		t.Fatal(err)
	}

	if err = mmWriter.Flush(); err != nil {
		t.Fatal(err)
	}

	mm[1024] = '!'
	ptr := (*uint64)(unsafe.Pointer(&mm[1024]))
	loaded := atomic.LoadUint64(ptr)
	fmt.Println(loaded)

	if err = mm.Flush(); err != nil {
		t.Fatal(err)
	}
	if err = mm.Unmap(); err != nil {
		t.Fatal()
	}
}

func TestSimpleCreateOpenFeedDelete(t *testing.T) {
	//base, _ := ioutil.TempFile(os.TempDir(), "lala-")
	base := "lala"
	defer cleanup(base)

	s, err := MmapStreamCreate(base, 64*1024, &GobSerialiser{})
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
