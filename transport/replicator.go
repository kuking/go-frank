package transport

import (
	"log"
	"net"
	"sync"
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
		if err == nil {
			sl := r.NewSyncLinkRecv(conn, conn.RemoteAddr().String(), "./") //TODO: set correct path
			go sl.goFuncRecv()
		} else {
			log.Printf("error accepting connection, err: %v\n", err)
		}
		if r.Close {
			return listener.Close()
		}
	}
}
