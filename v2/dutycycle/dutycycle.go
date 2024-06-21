package dutycycle

import (
	"time"
)

type BusyWait struct {
}

func (_ BusyWait) Loop() int64 {
	return 1 // estimated nil-wait, avoid using timers for speed, no need to be precise
}

func (_ BusyWait) Reset() {
}

func NewBusyWait() *BusyWait {
	return &BusyWait{}
}

type FastSpinThenWait struct {
	spinsBeforeSleep int
	minWaitNs        int64
	maxWaitNs        int64
	currentSpin      int
	currentWait      int64
}

func NewFastSpinThenWait(spinsBeforeSleep int, minWaitNs, maxWaitNs int64) *FastSpinThenWait {
	return &FastSpinThenWait{
		spinsBeforeSleep: spinsBeforeSleep,
		minWaitNs:        minWaitNs,
		maxWaitNs:        maxWaitNs,
		currentSpin:      0,
		currentWait:      0,
	}
}

func NewDefaultFastSpinThenWait() *FastSpinThenWait {
	return NewFastSpinThenWait(1_000_000, 1_000_000, 50_000_000)
}

func (w *FastSpinThenWait) Loop() (waitedNs int64) {
	w.currentSpin++
	if w.currentSpin > w.spinsBeforeSleep {
		waitedNs = w.currentWait
		time.Sleep(time.Duration(w.currentWait))
		w.currentWait = (w.currentWait + w.maxWaitNs) / 2
	} else {
		waitedNs = 1 // estimated nil-wait, avoid using timers for speed, no need to be precise
	}
	return
}

func (w *FastSpinThenWait) Reset() {
	w.currentSpin = 0
	w.currentWait = w.minWaitNs
}
