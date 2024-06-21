package api

type Provider[T any] interface {
	Feed(elem T)
	Close()
	IsClosed() bool
	Pull() (elem T, found bool)
	Prev() (moved bool)
	Reset() (position uint64)
	ReadPos() uint64
	WritePos() uint64
	WithWaitTimeOut(waitTimeOut WaitTimeOut)
	WithWaitDuty(waitDuty WaitDuty)
}
