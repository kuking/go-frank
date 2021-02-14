package persistent

import (
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"github.com/kuking/go-frank/serialisation"
	"math"
	"sync/atomic"
	"time"
)

func (s *MmapStream) SubscriberIdForName(namedSubscriber string) int {
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

func (s *MmapStream) GetReplicatorIds() (reps []int) {
	reps = make([]int, 0)
	for repId := 0; repId < mmapStreamMaxReplicators; repId++ {
		if len(serialisation.FromNTString(s.descriptor.SubName[repId][:])) != 0 {
			reps = append(reps, repId)
		}
	}
	return
}

func (s *MmapStream) ReplicatorIdForNameHost(name, host string) (repId, subId int, created bool) {
	for repId = 0; repId < mmapStreamMaxReplicators; repId++ {
		if serialisation.FromNTString(s.descriptor.RepName[repId][:]) == name {
			serialisation.ToNTString(s.descriptor.RepHost[repId][:], host)
			subId = s.SubscriberIdForName(fmt.Sprintf("REPL:%v", name))
			created = false
			return
		}
	}
	for repId = 0; repId < mmapStreamMaxReplicators; repId++ {
		if len(serialisation.FromNTString(s.descriptor.SubName[repId][:])) == 0 {
			break
		}
	}
	serialisation.ToNTString(s.descriptor.RepName[repId][:], name)
	serialisation.ToNTString(s.descriptor.RepHost[repId][:], host)
	subId = s.SubscriberIdForName(fmt.Sprintf("REPL:%v", name))
	created = true
	return
}

// Gets Subscriber Read Position
func (s *MmapStream) ReadSubRPos(subId int) uint64 {
	return atomic.LoadUint64(&s.descriptor.SubRPos[subId])
}

// Resets Subscriber Read Position to given AbsPos (advance: don't use, for replication purposes.)
func (s *MmapStream) SetSubRPos(subId int, absPos uint64) {
	atomic.StoreUint64(&s.descriptor.SubRPos[subId], absPos)
}

// Gets Writer Position
func (s *MmapStream) WritePos() uint64 {
	return atomic.LoadUint64(&s.descriptor.Write)
}

// Resets Writer Position (advanced: don't use, for replication purposes.)
func (s *MmapStream) SetWritePos(absPos uint64) {
	atomic.StoreUint64(&s.descriptor.Write, absPos)
}

func (s *MmapStream) GetUniqId() uint64 {
	return s.descriptor.UniqId
}