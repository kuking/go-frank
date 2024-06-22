package persistent

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/edsrzf/mmap-go"
	"github.com/kuking/go-frank/v1/serialisation"
	"math"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
)

type mmapPart struct {
	filename   string
	serialiser serialisation.StreamSerialiser
	partSize   uint64
	mmap       mmap.MMap
	descriptor *mmapPartFileDescriptor
}

func createMmapPart(baseFilename string, uniqId, partNo, partSize uint64) (err error) {
	fdpPath := baseFilename + fmt.Sprintf(".%05x", partNo)
	if err = mmapInit(fdpPath, mmapPartHeaderSize+int(partSize)); err != nil {
		return
	}

	mm, err := mmapOpen(fdpPath)
	if err != nil {
		return
	}
	fdpInMM := (*mmapPartFileDescriptor)(unsafe.Pointer(&mm[0]))
	*fdpInMM = mmapPartFileDescriptor{
		Version:  mmapStreamFileVersion,
		UniqId:   uniqId,
		PartNo:   partNo,
		IndexOfs: [mmapPartIndexSize]uint64{},
	}
	return mm.Unmap()
}

func openMmapPart(baseFilename string, uniqId, partNo, partSize uint64, serialiser serialisation.StreamSerialiser) (mp *mmapPart, err error) {
	mp = &mmapPart{
		filename:   baseFilename + fmt.Sprintf(".%05x", partNo),
		partSize:   partSize,
		serialiser: serialiser,
	}
	mp.mmap, err = mmapOpen(mp.filename)
	if err != nil {
		return
	}
	mp.descriptor = (*mmapPartFileDescriptor)(unsafe.Pointer(&mp.mmap[0]))
	if mp.descriptor.UniqId != uniqId {
		return nil, errors.New("part file is from another stream, different ids!")
	}
	return
}

func (mp *mmapPart) UpdateIndexes(absOfs uint64, ofs int) {
	idx := mmapPartIndexSize * ofs / int(mp.partSize)
	idxVal := atomic.LoadUint64(&mp.descriptor.IndexOfs[idx])
	if idxVal == 0 {
		atomic.StoreUint64(&mp.descriptor.IndexOfs[idx], absOfs)
	}
}

func (mp *mmapPart) WriteAt(absOfs uint64, elem interface{}, elemLength uint16) {
	ofs := int(absOfs % mp.partSize)
	localOfs := mmapPartHeaderSize + ofs
	binary.LittleEndian.PutUint16(mp.mmap[localOfs+2:], elemLength)
	if err := mp.serialiser.Encode(elem, mp.mmap[localOfs+entryHeaderSize:localOfs+entryHeaderSize+int(elemLength)]); err != nil {
		panic(fmt.Sprintf("could not write in part, err: %v", err))
	}
	mp.mmap[localOfs+1] = entryVersion
	mp.mmap[localOfs] = entryIsValid
	mp.UpdateIndexes(absOfs, ofs)
}

func (mp *mmapPart) WriteEoP(absOfs uint64) {
	localOfs := mmapPartHeaderSize + int(absOfs%mp.partSize)
	if uint64(localOfs) >= mp.partSize+uint64(mmapPartHeaderSize) {
		return
	}
	mp.mmap[localOfs] = entryIsEoP
}

func (mp *mmapPart) ReadAt(absOfs uint64) (elem interface{}, elemLength uint16) {
	var t0 *time.Time
	localOfs := mmapPartHeaderSize + int(absOfs%mp.partSize)
	if uint64(localOfs+entryHeaderSize) > mp.partSize+uint64(mmapPartHeaderSize) {
		return nil, math.MaxUint16
	}
	if mp.mmap[localOfs] == entryIsEoP {
		return nil, math.MaxUint16
	}
	for mp.mmap[localOfs] != entryIsValid {
		if t0 == nil {
			t := time.Now()
			t0 = &t
		} else {
			if time.Now().Sub(*t0).Milliseconds() > 100 {
				panic("implement marking dead element write") //TODO
			}
		}
		runtime.Gosched()
		time.Sleep(time.Nanosecond)
	}
	if mp.mmap[localOfs+1] != entryVersion {
		panic(fmt.Sprintf("non-supported part file entry version: %v", mp.mmap[localOfs+1]))
	}

	elemLength = binary.LittleEndian.Uint16(mp.mmap[localOfs+2:])
	var err error
	elem, err = mp.serialiser.Decode(mp.mmap[localOfs+entryHeaderSize : localOfs+entryHeaderSize+int(elemLength)])
	if err != nil {
		panic(fmt.Sprintf("could not read in part, err: %v", err))
	}
	return
}

func (mp *mmapPart) Close() error {
	return mp.mmap.Unmap()
}
