package base

import "github.com/kuking/go-frank/api"

type StreamProvider interface {
	Feed(elem interface{})
	Close()
	IsClosed() bool
	Pull() (elem interface{}, closed bool)
	Reset() uint64
	CurrAbsPos() uint64
	PeekLimit() uint64
	Peek(absPos uint64) interface{}
	Wait(approach api.WaitApproach)
}
