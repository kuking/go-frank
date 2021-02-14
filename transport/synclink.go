package transport

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/glycerine/rbuf"
	"github.com/kuking/go-frank/api"
	"github.com/kuking/go-frank/persistent"
	"log"
	"net"
	"time"
)

type SyncState int
type SyncDirection int

const (
	DISCONNECTED SyncState = iota
	HANDSHAKING  SyncState = iota
	PULLING      SyncState = iota
	PUSHING      SyncState = iota

	SEND SyncDirection = iota
	RECV SyncDirection = iota
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

func (s *SyncLink) disconnectAndIncErr() {
	_ = s.conn.Close()
	s.incError()
	s.State = DISCONNECTED
}

func (s *SyncLink) goFuncSend() {
	var err error
	var n int
	var wireHelloMsg WireHelloMsg
	var wireAckMsg WireAcksMsg
	var wireDataMsg WireDataMsg

	for {
		if s.State == DISCONNECTED {
			if s.errCount > 1 {
				time.Sleep(time.Duration(10*s.errCount) * time.Second) // trivial backoff of 10secs per errCount
			}
			s.conn, err = net.Dial("tcp", s.host)
			if err != nil {
				log.Printf("error connecting to host: %v, err: %v\n", s.host, err)
				s.incError()
			} else {
				s.State = HANDSHAKING
			}
		} else if s.State == HANDSHAKING {
			wireHelloMsg = WireHelloMsg{
				Version:      WireVersion,
				Message:      WireHELLO,
				Intention:    1, // push
				StreamUniqId: s.Stream.GetUniqId(),
			}
			err = binary.Write(s.conn, binary.LittleEndian, &wireHelloMsg)
			if err != nil {
				s.disconnectAndIncErr()
			} else {
				err = binary.Read(s.conn, binary.LittleEndian, &wireAckMsg)
				if err != nil {
					s.disconnectAndIncErr()
				} else {
					if wireAckMsg.Version != WireVersion || wireAckMsg.Message != WireNACKN {
						s.disconnectAndIncErr()
					} else {
						s.Stream.SetSubRPos(s.subId, wireAckMsg.AbsPos)
						s.State = PUSHING
					}
				}
			}
		} else if s.State == PUSHING {
			elem, readAbsPos, closed := s.Stream.PullBySubId(s.subId, api.WaitingUpto1s)
			bytes := elem.([]byte)
			if !closed {
				wireDataMsg = WireDataMsg{
					Version: WireVersion,
					Message: WireDATA,
					AbsPos:  readAbsPos,
					Length:  uint16(len(bytes)),
				}
				err = binary.Write(s.conn, binary.LittleEndian, &wireDataMsg)
				if err != nil {
					s.disconnectAndIncErr()
				} else {
					if n, err = s.conn.Write(bytes); n != len(bytes) || err != nil {
						s.disconnectAndIncErr()
					}
				}
			}
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
	} else if s.buf[1] == WireDESC {
		// handle wireDESC
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
