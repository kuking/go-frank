package go_frank

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/edsrzf/mmap-go"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"os"
)

const (
	mmapStreamFileVersion uint64 = 1

	nonPersistentSubId uint64 = 1 // all subId with this id are cleared upon file opening
)

type mmapStreamDescriptor struct {
	Version    uint64
	UniqId     uint64
	PartSize   uint64
	FirstPart  uint64 // to enable infinite streams, to cleanup old parts
	PartsCount uint64
	Write      uint64
	WAlloc     uint64
	Closed     uint32

	// 64 persistent subscribers
	SubId   [64]uint64 // an unique id
	SubRPos [64]uint64
	SubTime [64]int64 // last time a subscriber was active (reading/writing), updated rarely but helps to cleanup
}

//needs watchdog: Wlock!=WAlloc and values stay quiet for 1s, producer has die.

type mmapPartFileDescriptor struct {
	Version uint64
	UniqId  uint64 // same for descriptor and parts
	PartNo  uint64
	ElemOfs [32]uint64
	ElemNo  [32]uint64
	// All offsets are global, first 1kb of the part file contains the header and padding zeroes for simplicity.
	// ElemOfs/ElemNo form an pseudo-index, so the part file can partially seek without sequential reading.
	// by definition ElemNo[0] is the first element in the part file, ElemNo[31] is the last one. And the ones in
	// between are spread equally.
}

type mmapPart struct {
	filename   string
	serialiser StreamSerialiser
	partSize   uint64
	mmap       mmap.MMap
	descriptor *mmapPartFileDescriptor
}

func createMmapPart(baseFilename string, uniqId, partNo, partSize uint64) (err error) {
	fdpPath := baseFilename + fmt.Sprintf(".%03x", partNo)
	if err = mmapInit(fdpPath, 1024+int(partSize)); err != nil {
		return
	}

	mm, err := mmapOpen(fdpPath)
	if err != nil {
		return
	}
	fdpInMM := (*mmapPartFileDescriptor)(unsafe.Pointer(&mm[0]))
	*fdpInMM = mmapPartFileDescriptor{
		Version: mmapStreamFileVersion,
		UniqId:  uniqId,
		PartNo:  partNo,
		ElemOfs: [32]uint64{},
		ElemNo:  [32]uint64{},
	}
	return mm.Unmap()
}

func openMmapPart(baseFilename string, partSize, partNo uint64, serialiser StreamSerialiser) (mp *mmapPart, err error) {
	mp = &mmapPart{
		filename:   baseFilename + fmt.Sprintf(".%03x", partNo),
		partSize:   partSize,
		serialiser: serialiser,
	}
	mp.mmap, err = mmapOpen(mp.filename)
	if err != nil {
		return
	}
	mp.descriptor = (*mmapPartFileDescriptor)(unsafe.Pointer(&mp.mmap[0]))
	return
}

func (mp *mmapPart) WriteAt(absOfs uint64, elem interface{}, elemLength uint64) {
	localOfs := 1024 + int(absOfs%mp.partSize)
	binary.LittleEndian.PutUint64(mp.mmap[localOfs:], elemLength)
	if err := mp.serialiser.Encode(elem, mp.mmap[localOfs+8:localOfs+8]); err != nil {
		panic(fmt.Sprintf("could not write in part, err: %v", err))
	}
}

func (mp *mmapPart) Close() error {
	return mp.mmap.Unmap()
}

type mmapStream struct {
	serialiser     StreamSerialiser
	baseFilename   string
	descriptorMmap mmap.MMap
	descriptor     *mmapStreamDescriptor // pointing to the descriptor'Mmap
	subPart        [32]*mmapPart         // subscribers mmPart for readers
	writerPart     *mmapPart             // writer mmpart
	partLClock     sync.Mutex            // lock only used when loading parts or creating to avoid races on create/load
}

func MmapStreamCreate(baseFilename string, partSize uint64, serialiser StreamSerialiser) (s *mmapStream, err error) {
	if partSize < 64*1024 {
		return nil, errors.New("part file should be at least 64k")
	}
	rand.Seed(time.Now().Unix())
	fdfPath := baseFilename + ".fdf"
	if err = mmapInit(fdfPath, 2048); err != nil {
		return nil, err
	}
	mm, err := mmapOpen(fdfPath)
	if err != nil {
		return nil, err
	}
	fdfInMM := (*mmapStreamDescriptor)(unsafe.Pointer(&mm[0]))
	*fdfInMM = mmapStreamDescriptor{
		Version:    mmapStreamFileVersion,
		UniqId:     rand.Uint64(),
		PartSize:   partSize,
		FirstPart:  0,
		PartsCount: 0,
		Write:      0,
		WAlloc:     0,
		Closed:     0,
		SubId:      [64]uint64{},
		SubRPos:    [64]uint64{},
		SubTime:    [64]int64{},
	}
	if err = mm.Unmap(); err != nil {
		return nil, err
	}
	return MmapStreamOpen(baseFilename, serialiser)
}

func MmapStreamOpen(baseFilename string, serialiser StreamSerialiser) (s *mmapStream, err error) {
	s = &mmapStream{
		serialiser:   serialiser,
		baseFilename: baseFilename,
	}
	fdfPath := baseFilename + ".fdf"
	s.descriptorMmap, err = mmapOpen(fdfPath)
	if err != nil {
		return
	}
	s.descriptor = (*mmapStreamDescriptor)(unsafe.Pointer(&s.descriptorMmap[0]))
	return
}

func (s *mmapStream) CloseFile() error {
	//XXX close parts cache
	return s.descriptorMmap.Unmap()
}

func (s *mmapStream) resolvePart(subId int, partNo uint64) *mmapPart {
	// fast answer
	if subId == -1 && s.writerPart != nil && s.writerPart.descriptor.PartNo == partNo {
		return s.writerPart
	}
	if subId >= 0 && s.subPart[subId] != nil && s.subPart[subId].descriptor.PartNo == partNo {
		return s.subPart[subId]
	}

	if s.descriptor.PartsCount <= partNo {
		// a new part is created
		s.partLClock.Lock()
		if err := createMmapPart(s.baseFilename, s.descriptor.UniqId, partNo, s.descriptor.PartSize); err != nil {
			panic(fmt.Sprintf("failed to create a part file, err: %v", err)) //XXX: maybe less strict
		}
		s.descriptor.PartsCount++
		if err := s.descriptorMmap.Flush(); err != nil {
			panic(fmt.Sprintf("failed to flush descriptor, err: %v", err)) //XXX: maybe less strict
		}
		s.partLClock.Unlock()
	}

	// the following does not need synchronisation on the assumption only one part per subscriber will be relevant (no race)
	// and only one part for the writer will be relevant
	part, err := openMmapPart(s.baseFilename, s.descriptor.PartSize, partNo, s.serialiser)
	if err != nil {
		panic(fmt.Sprintf("failed to open a part file, fatal, err: %v", err)) //XXX: maybe less strict
	}
	if subId == -1 {
		if s.writerPart != nil {
			if err := s.writerPart.Close(); err != nil {
				panic(fmt.Sprintf("failed to close a part file, fatal, err: %v", err)) //XXX: maybe less strict
			}
		}
		s.writerPart = part
	} else {
		if s.subPart[subId] != nil {
			if err := s.subPart[subId].Close(); err != nil {
				panic(fmt.Sprintf("failed to close a part file, fatal, err: %v", err)) //XXX: maybe less strict
			}
		}
		s.subPart[subId] = part
	}
	return part
}

func (s *mmapStream) Feed(elem interface{}) {

	encodedSize, err := s.serialiser.EncodedSize(elem)
	if err != nil {
		log.Println("error retrieving encoded size, won't recover from this probably, err:", err)
		return
	}
	encodedSizePlusHeader := encodedSize + 8 // uint64 for length

	for i := 0; ; i++ {
		ofsWrite := atomic.LoadUint64(&s.descriptor.Write)
		ofsWAlloc := atomic.LoadUint64(&s.descriptor.WAlloc)

		if ofsWrite == ofsWAlloc {

			newOfsStart := ofsWrite
			newOfsWAlloc := ofsWAlloc + encodedSizePlusHeader

			partNo := ofsWrite / s.descriptor.PartSize
			partLeft := s.descriptor.PartSize - (ofsWrite % s.descriptor.PartSize)
			if partLeft < encodedSizePlusHeader {
				// it will not fit in the last bit of the part, so a new one is required
				partNo++
				newOfsStart += partLeft
				newOfsWAlloc += partLeft
			}

			if atomic.CompareAndSwapUint64(&s.descriptor.WAlloc, ofsWAlloc, newOfsWAlloc) {
				mp := s.resolvePart(-1, partNo)
				mp.WriteAt(newOfsStart, elem, encodedSize)
				if !atomic.CompareAndSwapUint64(&s.descriptor.Write, ofsWrite, newOfsWAlloc) {
					panic("failed to commit allocated write")
				}
				return
			}
		}
		runtime.Gosched()
		time.Sleep(time.Duration(i) * time.Nanosecond) // notice nanos vs micros
	}
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