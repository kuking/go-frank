package base

import (
	"time"
)

type BusyWait struct {
}

func (_ BusyWait) Loop(_ bool) {
}

func NewBusyWait() *BusyWait {
	return &BusyWait{}
}

type FastSpinThenWait struct {
	spinsBeforeSleep int
	minWaitNs        uint64
	maxWaitNs        uint64
	currentSpin      int
	currentWait      uint64
}

func NewFastSpinThenWait(spinsBeforeSleep int, minWaitNs, maxWaitNs uint64) *FastSpinThenWait {
	return &FastSpinThenWait{
		spinsBeforeSleep: spinsBeforeSleep,
		minWaitNs:        minWaitNs,
		maxWaitNs:        maxWaitNs,
		currentSpin:      0,
		currentWait:      0,
	}
}

func (w *FastSpinThenWait) Loop(hasProcessed bool) {
	if hasProcessed {
		w.currentSpin = 0
		w.currentWait = w.minWaitNs
		return
	}
	w.currentSpin++
	if w.currentSpin > w.spinsBeforeSleep {
		time.Sleep(time.Duration(w.currentWait))
		w.currentWait = (w.currentWait + w.maxWaitNs) / 2
	}
}
