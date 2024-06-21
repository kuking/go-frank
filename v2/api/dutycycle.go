package api

type WaitTimeOut int64

const (
	UntilClosed     WaitTimeOut = -1
	UntilNoMoreData WaitTimeOut = 0
	WaitingUpto1us  WaitTimeOut = 1_000
	WaitingUpto1ms              = WaitingUpto1us * 1_000
	WaitingUpto1s               = WaitingUpto1ms * 1_000
	WaitingUpTo1m               = WaitingUpto1s * 60
)

// WaitDuty specifies how to wait and process for new data. Different strategies are a trade-off between latency and CPU usage. i.e.
// - busy wait will have the lowest latency, at the expense of putting one CPU at 100% usage.
// - always-wait, will impose a few ms latency due to clock-checking, wait and context-switching.
// A hybrid solution it usually a good trade-off, fast-spin for a few thousand times, and then if no activity is detected the thread is put to sleep and
// the context is switched to another process/goroutine; each time waiting a bit longer until certain upper limit.
type WaitDuty interface {
	// Loop is called after a failed attempt to retrieve a value from a stream due to the fact no data is available.
	// the method will wait a time (yielding the goroutine, sleeping the CPU or by other means.) and return how many nanoseconds it waited.
	// different strategies will wait more or less
	Loop() (waitedNs int64)
	// Reset is called at the beginning of the main loop to notify the WaitDuty that a new pool is being attempted
	Reset()
}

// allocation free reducer
//type Reducer interface {
//	First(interface{})
//	Next(interface{})
//	Result() interface{}
//}
