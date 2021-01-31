package persistent

import (
	"github.com/edsrzf/mmap-go"
	"os"
)

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
