package transport

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/glycerine/rbuf"
	"github.com/kuking/go-frank/api"
	"github.com/kuking/go-frank/misc"
	"github.com/kuking/go-frank/persistent"
	"log"
	"net"
	"sync/atomic"
	"time"
)

type SyncState int

const (
	DISCONNECTED SyncState = iota
	CONNECTED    SyncState = iota
	HANDSHAKING  SyncState = iota
	PULLING      SyncState = iota
	PUSHING      SyncState = iota
)

type SyncLink struct {
	repl     *Replicator
	errT0    time.Time
	errCount int
	conn     net.Conn
	host     string
	rBuf     *rbuf.FixedSizeRingBuf
	buf      []byte
	State    SyncState
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
		rBuf:     nil,
		buf:      nil,
		State:    CONNECTED,
		Stream:   stream,
		repName:  repName,
		repId:    repId,
		subId:    subId,
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
	if s.State != DISCONNECTED { // so it does not fails when we know it will fail
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

}

func (s *SyncLink) Handle() {

	var n int
	var m int
	var err error
	for {
		n, err = s.conn.Read(s.buf)
		if err != nil {
			log.Printf("failed to read from: %v, err: %v\n", s.conn.RemoteAddr(), err)
			break
		}
		m, err = s.rBuf.Write(s.buf[0:n])
		if err != nil || n != m {
			log.Printf("failed to write into ring-buffer from: %v, err: %v\n", s.conn.RemoteAddr(), err)
			break
		}
		err = s.seekAndHandleMessage()
		if err != nil {
			log.Printf("failed to handle message from: %v, err: %v\n", s.conn.RemoteAddr(), err)
			break
		}
	}
	err = s.conn.Close()
	if err != nil {
		log.Printf("failed to close socket, err: %v\n", err)
	}
}

func (s *SyncLink) seekAndHandleMessage() error {
	n, err := s.rBuf.ReadWithoutAdvance(s.buf[0:2])
	if n != 2 || err != nil {
		return nil // assumes buffer not complete, so no error
	}
	if s.buf[0] != WireVersion {
		return errors.New(fmt.Sprintf("unknown wire version %v", s.buf[0]))
	}
	if s.buf[1] == WireHELLO {
		// handle hello
	} else if s.buf[1] == WireSTATUS {
		// handle wireSTATUS
	} else if s.buf[1] == WireACK {
		// handle ACK
	} else if s.buf[1] == WireNACK1 {
		// handle NACK1
	} else if s.buf[1] == WireNACKN {
		// handle WireNACKN
	} else if s.buf[1] == WireDATA {
		// handle WireDATA
	} else {
		return errors.New(fmt.Sprintf("unknown message %v", s.buf[1]))
	}
	return nil
}
