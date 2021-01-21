package go_frank

import (
	"github.com/edsrzf/mmap-go"
	"os"
)

const (
	mmapStreamFileVersion uint64 = 1

	nonPersistentSubId uint64 = 1 // all subId with this id are cleared upon file opening
)

type mmapStreamDescriptor struct {
	version        uint64
	partFileSize   uint64
	PartFilesCount uint64
	WLock          uint64
	WAlloc         uint64

	// 64 persistent subscribers
	SubId   [64]uint64
	SubRPos [64]uint64

	ObjIdx [1024 * 2]uint64
}

type mmapPartFileDescriptor struct {
	version    uint64
	partFileNo uint64
	ElemOfs    [32]uint64
	ElemNo     [32]uint64
	// i.e. (ElemOfs[0]=528, ElemNo[0]=123), (ElemOfs[1]=4096, ElemNo[1]=230)
	// At part file offset 528 (just after header), starts the element 123, at offset 4096, starts the element 230.
	// by definition, the first offset is the first byte after the file descriptor, the element index is filled upon
	// the first element passing the offset
}

func mmapInit(filename string, size int) error {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	if _, err = f.Seek(int64(size)-1, 0); err != nil {
		return err
	}
	if _, err = f.Write([]byte{0}); err != nil {
		return err
	}
	return f.Close()
}

func mmapOpen(filename string) (mmap.MMap, error) {
	f, err := os.OpenFile(filename, os.O_RDWR, 0644)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	return mmap.Map(f, mmap.RDWR, 0)
}
