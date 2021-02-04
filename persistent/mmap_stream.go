package persistent

import (
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/edsrzf/mmap-go"
	"github.com/kuking/go-frank/api"
	"github.com/kuking/go-frank/serialisation"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type mmapStream struct {
	serialiser     serialisation.StreamSerialiser
	baseFilename   string
	descriptorMmap mmap.MMap
	descriptor     *mmapStreamDescriptor           // pointing to the descriptor'Mmap
	subPart        [mmapStreamMaxClients]*mmapPart // subscribers mmPart for readers
	writerPart     *mmapPart                       // writer mmpart
	partLClock     sync.Mutex                      // lock only used when loading parts or creating to avoid races on create/load
	subIdLock      sync.Mutex                      // lock used to allocate unique subId
}

func MmapStreamCreate(baseFilename string, partSize uint64, serialiser serialisation.StreamSerialiser) (s *mmapStream, err error) {
	if partSize < 64*1024 {
		return nil, errors.New("part file should be at least 64k")
	}
	rand.Seed(time.Now().Unix())
	fdfPath := baseFilename + ".frank"
	if err = mmapInit(fdfPath, mmapStreamHeaderSize); err != nil {
		return nil, err
	}
	mm, err := mmapOpen(fdfPath)
	if err != nil {
		return nil, err
	}
	uniqId := rand.Uint64()
	fdfInMM := (*mmapStreamDescriptor)(unsafe.Pointer(&mm[0]))
	*fdfInMM = mmapStreamDescriptor{
		Version:    mmapStreamFileVersion,
		UniqId:     uniqId,
		ReplicaOf:  uniqId,
		PartSize:   partSize,
		FirstPart:  0,
		PartsCount: 0,
		Write:      0,
		Closed:     0,
		SubId:      [mmapStreamMaxClients]uint64{},
		SubRPos:    [mmapStreamMaxClients]uint64{},
		SubTime:    [mmapStreamMaxClients]int64{},
		RepUniqId:  [mmapStreamMaxReplicators]uint64{},
		RepHWMPos:  [mmapStreamMaxReplicators]uint64{},
		RepHost:    [mmapStreamMaxReplicators][128]byte{},
	}
	if err = mm.Unmap(); err != nil {
		return nil, err
	}
	return MmapStreamOpen(baseFilename, serialiser)
}

func MmapStreamOpen(baseFilename string, serialiser serialisation.StreamSerialiser) (s *mmapStream, err error) {
	s = &mmapStream{
		serialiser:   serialiser,
		baseFilename: baseFilename,
	}
	fdfPath := baseFilename + ".frank"
	s.descriptorMmap, err = mmapOpen(fdfPath)
	if err != nil {
		return
	}
	s.descriptor = (*mmapStreamDescriptor)(unsafe.Pointer(&s.descriptorMmap[0]))
	return
}

func (s *mmapStream) CloseFile() error {
	for _, part := range s.subPart {
		if part != nil {
			_ = part.Close()
		}
	}
	if s.writerPart != nil {
		_ = s.writerPart.Close()
	}
	return s.descriptorMmap.Unmap()
}

func (s *mmapStream) Delete() error {
	_ = s.CloseFile()
	files, err := filepath.Glob(s.baseFilename + ".?????")
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return err
		}
	}
	return nil
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
		part, err := openMmapPart(s.baseFilename, s.descriptor.UniqId, partNo, s.descriptor.PartSize, s.serialiser)
		if err != nil {
			if err := createMmapPart(s.baseFilename, s.descriptor.UniqId, partNo, s.descriptor.PartSize); err != nil {
				panic(fmt.Sprintf("failed to create a part file, err: %v", err)) //XXX: maybe less strict
			}
			s.descriptor.PartsCount++
			if err := s.descriptorMmap.Flush(); err != nil {
				panic(fmt.Sprintf("failed to flush descriptor, err: %v", err)) //XXX: maybe less strict
			}
		}
		s.partLClock.Unlock()
		if part != nil && err == nil { // to avoid race
			return part
		}
	}

	// the following does not need synchronisation on the assumption only one part per subscriber will be relevant (no race)
	// and only one part for the writer will be relevant
	part, err := openMmapPart(s.baseFilename, s.descriptor.UniqId, partNo, s.descriptor.PartSize, s.serialiser)
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
	encodedSizePlusHeader := int(encodedSize) + entryHeaderSize
	for i := 0; ; i++ {
		ofsWrite := atomic.LoadUint64(&s.descriptor.Write)
		oldOfsWrite := ofsWrite
		newOfsWrite := ofsWrite + uint64(encodedSizePlusHeader)
		partNo := ofsWrite / s.descriptor.PartSize
		partLeft := s.descriptor.PartSize - (ofsWrite % s.descriptor.PartSize)
		writeEoP := false
		if partLeft < uint64(encodedSizePlusHeader) {
			writeEoP = true
			partNo++
			ofsWrite += partLeft
			newOfsWrite += partLeft
		}
		if atomic.CompareAndSwapUint64(&s.descriptor.Write, oldOfsWrite, newOfsWrite) {
			if writeEoP {
				mp := s.resolvePart(-1, partNo-1)
				mp.WriteEoP(oldOfsWrite)
			}
			mp := s.resolvePart(-1, partNo)
			mp.WriteAt(ofsWrite, elem, encodedSize)
			return
		}
		runtime.Gosched()
		time.Sleep(time.Duration(i) * time.Nanosecond) // notice nanos vs micros
	}
}

func (s *mmapStream) pullBySubId(subId int, waitApproach api.WaitApproach) (elem interface{}, closed bool) {
	var totalNsWait int64
	for i := 0; ; i++ {
		ofsRead := atomic.LoadUint64(&s.descriptor.SubRPos[subId])
		ofsWrite := atomic.LoadUint64(&s.descriptor.Write)
		if ofsRead < ofsWrite {
			var ofsNewRead uint64
			partNo := ofsRead / s.descriptor.PartSize
			part := s.resolvePart(subId, partNo)
			value, length := part.ReadAt(ofsRead)
			if length == math.MaxUint16 {
				partNo++
				endSlack := s.descriptor.PartSize - (ofsRead % s.descriptor.PartSize)
				part = s.resolvePart(subId, partNo)
				value, length = part.ReadAt(ofsRead + endSlack)
				ofsNewRead = ofsRead + endSlack + uint64(entryHeaderSize) + uint64(length)
			} else {
				ofsNewRead = ofsRead + uint64(entryHeaderSize) + uint64(length)
			}
			if atomic.CompareAndSwapUint64(&s.descriptor.SubRPos[subId], ofsRead, ofsNewRead) {
				return value, false
			}
		} else if s.IsClosed() {
			return nil, true
		}
		runtime.Gosched()
		time.Sleep(time.Duration(i) * time.Nanosecond)
		totalNsWait += int64(i)
		if waitApproach == api.UntilClosed {
			// just continue
		} else if totalNsWait > int64(waitApproach) {
			return nil, true
		}
	}
}

func (s *mmapStream) SubscriberIdForName(namedSubscriber string) int {
	s.subIdLock.Lock()
	defer s.subIdLock.Unlock()

	c := sha512.Sum512_256([]byte(namedSubscriber))
	subIdForName := s.descriptor.UniqId ^
		binary.LittleEndian.Uint64(c[0:8]) ^
		binary.LittleEndian.Uint64(c[8:16]) ^
		binary.LittleEndian.Uint64(c[16:24]) ^
		binary.LittleEndian.Uint64(c[24:32])

	// already subscribed? reuse
	for i, subId := range s.descriptor.SubId {
		if subId == subIdForName {
			s.descriptor.SubTime[i] = time.Now().UnixNano()
			serialisation.ToNTString(s.descriptor.SubName[i][:], namedSubscriber)
			return i
		}
	}

	// find a new sloth
	var possibleSubId int
	posibleSubTime := int64(math.MaxInt64)
	for i := 0; i < len(s.descriptor.SubId); i++ {
		if posibleSubTime > s.descriptor.SubTime[i] {
			posibleSubTime = s.descriptor.SubTime[i]
			possibleSubId = i
		}
	}
	// picks the older subscriber slot
	s.descriptor.SubTime[possibleSubId] = time.Now().UnixNano()
	serialisation.ToNTString(s.descriptor.SubName[possibleSubId][:], namedSubscriber)
	s.descriptor.SubId[possibleSubId] = subIdForName
	s.descriptor.SubRPos[possibleSubId] = 0
	return possibleSubId

}

func (s *mmapStream) Close() {
	atomic.StoreUint32(&s.descriptor.Closed, 1)
}

func (s *mmapStream) IsClosed() bool {
	return atomic.LoadUint32(&s.descriptor.Closed) != 0
}

func (s *mmapStream) Reset(subId int) uint64 {
	atomic.StoreUint64(&s.descriptor.SubRPos[subId], 0) //XXX: fix when prune is implemented
	return 0
}
