package base

import (
	"testing"
	"time"
)

func TestBusyWait(t *testing.T) {

	// estimating timing factor
	w := NewBusyWait()
	t0 := time.Now()
	for i := 0; i < 1_000; i++ {
		w.Loop(false)
	}
	factor := 2 * time.Now().Sub(t0).Nanoseconds() / 1_000

	// test
	w = NewBusyWait()
	t0 = time.Now()
	for i := 0; i < 1_000_000; i++ {
		w.Loop(false)
	}
	tdNs := time.Now().Sub(t0).Nanoseconds()
	if tdNs > factor*1_000_000 {
		t.Fatal("It should be way faster")
	}
}

func TestNewFastSpinThenWait(t *testing.T) {

	// estimates timing factor
	w := NewFastSpinThenWait(1_000_000, 1_000_000, 10_000_000)
	t0 := time.Now()
	for i := 0; i < 1_000; i++ {
		w.Loop(false)
	}
	factor := 2 * time.Now().Sub(t0).Nanoseconds() / 1_000

	// actual test
	w = NewFastSpinThenWait(1_000_000, 1_000_000, 10_000_000)

	// fast spi
	t0 = time.Now()
	for i := 0; i < 1_000_000; i++ {
		w.Loop(false)
	}
	tdNs := time.Now().Sub(t0).Nanoseconds()
	if tdNs > 1_000_000*factor {
		t.Fatal("It should have fast spin", tdNs)
	}

	// slow, wait
	t0 = time.Now()
	for i := 0; i < 10; i++ {
		w.Loop(false)
	}
	tdNs = time.Now().Sub(t0).Nanoseconds()
	if tdNs < 5_000_000*10 {
		t.Fatal("It should have waited")
	}

	// once one is processed, it becomes fast spin again
	t0 = time.Now()
	w.Loop(true)
	for i := 0; i < 1_000_000; i++ {
		w.Loop(false)
	}
	tdNs = time.Now().Sub(t0).Nanoseconds()
	if tdNs > 1_000_000*factor {
		t.Fatal("It should have fast spin again", tdNs)
	}

}
