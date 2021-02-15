package transport

import (
	"fmt"
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

func (r *Replicator) houseKeeping() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for i, link := range r.Links {
		if link.State == DISCONNECTED {
			r.Links[len(r.Links)-1], r.Links[i] = r.Links[i], r.Links[len(r.Links)-1]
			r.Links = r.Links[:len(r.Links)-1]
		}
	}
}

func (r *Replicator) WaitAll(exitOnZero bool) {
	prevReadPos := map[uint64]uint64{}
	prevWritePos := map[uint64]uint64{}

	for {
		r.houseKeeping()
		time.Sleep(1 * time.Second)
		for i, l := range r.Links {
			uniqId := l.Stream.GetUniqId()
			readPos := l.Stream.ReadSubRPos(l.subId)
			writePos := l.Stream.WritePos()
			if prevReadPos[uniqId] != 0 || prevWritePos[uniqId] != 0 {
				readPct := float32(readPos) * 100.0 / float32(writePos)
				readMiB := float32((readPos - prevReadPos[uniqId]) / 1024.0 / 1024.0)
				writeMiB := float32((writePos - prevWritePos[uniqId]) / 1024.0 / 1024.0)
				fmt.Printf("[%v: R: %v (%2.2fMiB) %2.2f%% W: %v (%2.2fMiB)]\n", i, readPos, readMiB, readPct, writePos, writeMiB)
			}
			prevReadPos[uniqId] = readPos
			prevWritePos[uniqId] = writePos
		}
		if exitOnZero {
			r.mutex.Lock()
			l := len(r.Links)
			r.mutex.Unlock()
			if l == 0 {
				return
			}
		}
	}
}

func (r *Replicator) ConnectTCP(stream *persistent.MmapStream, replicatorName string, bind string) error {
	conn, err := net.Dial("tcp", bind)
	if err != nil {
		return err
	}
	sl := r.NewSyncLinkSend(conn, bind, stream, replicatorName)
	go sl.goFuncSend()
	return nil
}

func (r *Replicator) ListenTCP(bind string, basePath string) error {
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
			sl := r.NewSyncLinkRecv(conn, conn.RemoteAddr().String(), basePath)
			go sl.goFuncRecv()
		} else {
			log.Printf("error accepting connection, err: %v\n", err)
		}
		if r.Close {
			return listener.Close()
		}
	}
}
