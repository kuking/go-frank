package transport

import (
	"errors"
	"fmt"
	"github.com/glycerine/rbuf"
	"github.com/kuking/go-frank/api"
	"log"
	"net"
)

type Replicator struct {
	Close bool
}

type SyncState int

const (
	HANDSHAKING SyncState = iota
	PULLING     SyncState = iota
	PUSHING     SyncState = iota
)

type SyncLink struct {
	conn   net.Conn
	rBuf   *rbuf.FixedSizeRingBuf
	buf    []byte
	State  SyncState
	Stream api.PersistentStream
}

func (l *SyncLink) Handle() {
	l.rBuf = rbuf.NewFixedSizeRingBuf(65535 * 2)
	l.buf = make([]byte, 1024)
	var n int
	var m int
	var err error
	for {
		n, err = l.conn.Read(l.buf)
		if err != nil {
			log.Printf("failed to read from: %v, err: %v\n", l.conn.RemoteAddr(), err)
			break
		}
		m, err = l.rBuf.Write(l.buf[0:n])
		if err != nil || n != m {
			log.Printf("failed to write into ring-buffer from: %v, err: %v\n", l.conn.RemoteAddr(), err)
			break
		}
		err = l.seekAndHandleMessage()
		if err != nil {
			log.Printf("failed to handle message from: %v, err: %v\n", l.conn.RemoteAddr(), err)
			break
		}
	}
	err = l.conn.Close()
	if err != nil {
		log.Printf("failed to close socket, err: %v\n", err)
	}
}

func (l *SyncLink) seekAndHandleMessage() error {
	n, err := l.rBuf.ReadWithoutAdvance(l.buf[0:2])
	if n != 2 || err != nil {
		return nil // assumes buffer not complete, so no error
	}
	if l.buf[0] != WireVersion {
		return errors.New(fmt.Sprintf("unknown wire version %v", l.buf[0]))
	}
	if l.buf[1] == WireHELLO {
		// handle hello
	} else if l.buf[1] == WireDESC {
		// handle wireDESC
	} else if l.buf[1] == WireACK {
		// handle ACK
	} else if l.buf[1] == WireNACK1 {
		// handle NACK1
	} else if l.buf[1] == WireNACKN {
		// handle WireNACKN
	} else if l.buf[1] == WireDATA {
		// handle WireDATA
	} else {
		return errors.New(fmt.Sprintf("unknown message %v", l.buf[1]))
	}
	return nil
}

func (s *Replicator) ListenTCP(bind string) error {
	serverAddr, err := net.ResolveTCPAddr("tcp", bind)
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", serverAddr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err == nil {
			link := SyncLink{
				conn:   conn,
				State:  HANDSHAKING,
				Stream: nil,
			}
			go link.Handle()
		} else {
			log.Printf("error accepting connection, err: %v\n", err)
		}
		if s.Close {
			return listener.Close()
		}
	}
}

func (s *Replicator) ListenUDP(bind string) error {

	ServerAddr, err := net.ResolveUDPAddr("udp", bind)
	if err != nil {
		return err
	}

	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	if err != nil {
		return err
	}

	buf := make([]byte, 1024)
	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		fmt.Println("Received:", string(buf[0:n]), "from", addr, "err:", err)
		if s.Close {
			break
		}
	}
	return ServerConn.Close()
}
