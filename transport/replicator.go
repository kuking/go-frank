package transport

import (
	"github.com/glycerine/rbuf"
	"github.com/kuking/go-frank/persistent"
	"log"
	"net"
	"sync"
	"time"
)

type Replicator struct {
	mutex sync.Mutex
	Links []*SyncLink
	Close bool
}

func NewReplicator() *Replicator {
	return &Replicator{
		Links: make([]*SyncLink, 0),
		Close: false,
	}
}

func (r *Replicator) addSyncLink(link *SyncLink) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.Links = append(r.Links, link)
}

func (r *Replicator) Send(stream *persistent.MmapStream, replicatorName string, host string) {
	repId, subId, _ := stream.ReplicatorIdForNameHost(replicatorName, host)
	rBuf := rbuf.NewFixedSizeRingBuf(65535 * 2)
	buf := make([]byte, 1024)
	sl := SyncLink{
		repl:     r,
		errT0:    time.Time{},
		errCount: 0,
		conn:     nil,
		host:     host,
		rBuf:     rBuf,
		buf:      buf,
		State:    DISCONNECTED,
		Stream:   stream,
		repName:  replicatorName,
		repId:    repId,
		subId:    subId,
	}
	r.mutex.Lock()
	r.Links = append(r.Links, &sl)
	r.mutex.Unlock()
	go sl.goFuncSend()
}

func (r *Replicator) ListenTCP(bind string) error {
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
		rBuf := rbuf.NewFixedSizeRingBuf(65535 * 2)
		buf := make([]byte, 1024)
		if err == nil {
			sl := SyncLink{
				repl:     r,
				errT0:    time.Time{},
				errCount: 0,
				conn:     conn,
				host:     conn.RemoteAddr().String(),
				rBuf:     rBuf,
				buf:      buf,
				State:    HANDSHAKING,
				Stream:   nil,
				repName:  "",
				repId:    -1,
				subId:    -1,
			}
			r.mutex.Lock()
			r.Links = append(r.Links, &sl)
			r.mutex.Unlock()
			go sl.goFuncRecv()
		} else {
			log.Printf("error accepting connection, err: %v\n", err)
		}
		if r.Close {
			return listener.Close()
		}
	}
}
