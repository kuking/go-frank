package transport

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/kuking/go-frank/api"
	"github.com/kuking/go-frank/misc"
	"github.com/kuking/go-frank/persistent"
	"github.com/kuking/go-frank/serialisation"
	"log"
	"net"
	"path"
	"sync/atomic"
	"time"
)

type SyncState int

const (
	DISCONNECTED SyncState = iota
	CONNECTED    SyncState = iota
	PULLING      SyncState = iota
	PUSHING      SyncState = iota
)

type SyncLink struct {
	repl     *Replicator
	errT0    time.Time
	errCount int
	conn     net.Conn
	host     string
	State    SyncState
	basePath string
	Stream   *persistent.MmapStream
	repName  string
	repId    int
	subId    int
	close    uint32
}

func (r *Replicator) NewSyncLinkSend(conn net.Conn, host string, stream *persistent.MmapStream, repName string) *SyncLink {
	repId, subId, _ := stream.ReplicatorIdForNameHost(repName, host)
	sl := &SyncLink{
		repl:     r,
		errT0:    time.Time{},
		errCount: 0,
		conn:     conn,
		host:     host,
		State:    CONNECTED,
		basePath: "",
		Stream:   stream,
		repName:  repName,
		repId:    repId,
		subId:    subId,
		close:    0,
	}
	r.addSyncLink(sl)
	return sl
}

func (r *Replicator) NewSyncLinkRecv(conn net.Conn, host string, basePath string) *SyncLink {
	sl := &SyncLink{
		repl:     r,
		errT0:    time.Time{},
		errCount: 0,
		conn:     conn,
		host:     host,
		State:    CONNECTED,
		basePath: basePath,
		Stream:   nil,
		repName:  "",
		repId:    0,
		subId:    0,
		close:    0,
	}
	r.addSyncLink(sl)
	return sl
}

func (s *SyncLink) Close() {
	atomic.StoreUint32(&s.close, 1)
}

func (s *SyncLink) Closed() bool {
	return atomic.LoadUint32(&s.close) > 0
}

func (s *SyncLink) incError() {
	if s.errT0.IsZero() {
		s.errT0 = time.Now()
		s.errCount = 1
	} else {
		s.errT0 = time.Time{}
		s.errCount = 0
	}
}

func (s *SyncLink) resetError() {
	s.errT0 = time.Time{}
	s.errCount = 0
}

func (s *SyncLink) handleError(err error) bool {
	if err == nil {
		return false
	}
	if s.State != DISCONNECTED { // so it does not logs extra errors I suppose ... ?
		log.Println(err)
	}
	s.incError()
	_ = s.conn.Close()
	s.State = DISCONNECTED
	return true
}

func (s *SyncLink) goFuncSend() {
	var n int
	var err error
	var wireHelloMsg WireHelloMsg
	var wireStatusMsg WireStatusMsg
	var wireDataMsg WireDataMsg

	go s.goFuncSendAncillary()

	for {
		if s.Closed() {
			_ = s.conn.Close()
			s.State = DISCONNECTED
			return
		}
		if s.State == CONNECTED {
			wireHelloMsg = WireHelloMsg{
				Version:      WireVersion,
				Message:      WireHELLO,
				StreamUniqId: s.Stream.GetUniqId(),
				PartSize:     s.Stream.GetPartSize(),
				FirstPart:    s.Stream.GetFirstPart(),
			}
			if s.handleError(binary.Write(s.conn, binary.LittleEndian, &wireHelloMsg)) {
				return
			}

			wireStatusMsg = WireStatusMsg{
				Version:    WireVersion,
				Message:    WireSTATUS,
				FirstPart:  s.Stream.GetFirstPart(),
				PartsCount: s.Stream.GetPartsCount(),
				Closed:     misc.AsUint32Bool(s.Stream.IsClosed()),
			}
			if s.handleError(binary.Write(s.conn, binary.LittleEndian, &wireStatusMsg)) {
				return
			}

			s.Stream.SetSubRPos(s.subId, s.Stream.GetRepHWM(s.repId))

			s.State = PUSHING
		}
		if s.State == PUSHING {
			elem, absPos, closed := s.Stream.PullBySubId(s.subId, api.WaitingUpto10ms)
			if !closed {
				wireDataMsg = WireDataMsg{
					Version: WireVersion,
					Message: WireDATA,
					AbsPos:  absPos,
					Length:  uint16(len(elem.([]byte))),
				}
				if s.handleError(binary.Write(s.conn, binary.LittleEndian, &wireDataMsg)) {
					return
				}
				n, err = s.conn.Write(elem.([]byte))
				if n != len(elem.([]byte)) {
					err = errors.New("short write")
				}
				if s.handleError(err) {
					return
				}
			}
		}
	}
}

// handles ACKs readings
func (s *SyncLink) goFuncSendAncillary() {
	var wireAcksMsg WireAcksMsg
	for {
		if s.Closed() {
			return
		}
		// it will be mostly blocked here
		if s.handleError(binary.Read(s.conn, binary.LittleEndian, &wireAcksMsg)) {
			return
		}
		if wireAcksMsg.Version != WireVersion {
			s.handleError(errors.New(fmt.Sprintf("received a message of unknown wire version: %v", wireAcksMsg.Version)))
		}
		if wireAcksMsg.Message != WireNACKN && wireAcksMsg.Message != WireACK {
			s.handleError(errors.New(fmt.Sprintf("received an unrecognised message type: %v", wireAcksMsg.Message)))
		}
		if wireAcksMsg.Message == WireNACKN {
			s.Stream.SetSubRPos(s.subId, wireAcksMsg.AbsPos)
		}
		if wireAcksMsg.Message == WireACK {
			s.Stream.SetRepHWM(s.repId, wireAcksMsg.AbsPos)
		}
	}
}

func (s *SyncLink) goFuncRecv() {
	var bytes []byte
	var err error
	conn := NewBufferedConnSize(s.conn, 64)
	var wireHelloMsg WireHelloMsg
	var wireStatusMsg WireStatusMsg

	for {
		if s.Closed() {
			_ = conn.Close()
			s.State = DISCONNECTED
			return
		}

		bytes, err = conn.Peek(2)
		if s.handleError(err) {
			return
		}
		if bytes[0] != WireVersion {
			s.handleError(errors.New(fmt.Sprintf("invalid wire version: %v", bytes[0])))
			return
		}
		if bytes[1] == WireHELLO {
			if s.State != CONNECTED {
				s.handleError(errors.New(fmt.Sprintf("unexpected WireHELLO message")))
				return
			}
			if s.handleError(binary.Read(conn, binary.LittleEndian, &wireHelloMsg)) {
				return
			}
			if wireHelloMsg.Message != WireHELLO {
				s.handleError(errors.New("invalid WireHELLO message"))
				return
			}
			baseName := path.Join(s.basePath, fmt.Sprintf("%x", wireHelloMsg.StreamUniqId))
			s.Stream, err = persistent.MmapStreamOpen(baseName, serialisation.ByteArraySerialiser{})
			if err != nil {
				s.Stream, err = persistent.MmapStreamCreate(baseName, wireHelloMsg.PartSize, serialisation.ByteArraySerialiser{})
				if s.handleError(err) {
					return
				}
			}
			s.State = PULLING
		}
		if bytes[1] == WireSTATUS {
			if s.handleError(binary.Read(conn, binary.LittleEndian, &wireStatusMsg)) {
				return
			}
			if misc.Uint32Bool(wireStatusMsg.Closed) {
				s.Stream.Close()
			}
		}
		if bytes[1] == WireDATA {
			if s.State != PULLING {
				s.handleError(errors.New("unexpected WireDATA message"))
				return
			}

		}

	}
}
