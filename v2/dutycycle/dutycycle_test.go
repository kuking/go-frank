package dutycycle

import (
	"testing"
	"time"
)

func TestBusyWait(t *testing.T) {
	// estimating timing factor
	w := NewBusyWait()
	t0 := time.Now()
	for i := 0; i < 1_000; i++ {
		w.Loop()
	}
	factor := (5 * time.Now().Sub(t0).Nanoseconds() / 1_000) + 1

	// test
	w = NewBusyWait()
	t0 = time.Now()
	for i := 0; i < 1_000_000; i++ {
		w.Loop()
	}
	tdNs := time.Now().Sub(t0).Nanoseconds()
	if tdNs > factor*1_000_000.0 {
		t.Fatalf("It should be way faster: %d > %d\n", tdNs, factor*1_000_000)
	}
}

func TestNewFastSpinThenWait(t *testing.T) {

	// estimates timing factor
	w := NewFastSpinThenWait(1_000_000, 1_000_000, 10_000_000)
	t0 := time.Now()
	for i := 0; i < 1_000; i++ {
		w.Loop()
	}
	factor := (5 * time.Now().Sub(t0).Nanoseconds() / 1_000.0) + 1

	// actual test
	w = NewFastSpinThenWait(1_000_000, 1_000_000, 10_000_000)

	// fast spi
	t0 = time.Now()
	for i := 0; i < 1_000_000; i++ {
		w.Loop()
	}
	tdNs := time.Now().Sub(t0).Nanoseconds()
	if tdNs > 1_000_000*factor {
		t.Fatalf("It should be way faster: %d > %d\n", tdNs, factor*1_000_000)
	}

	// slow, wait
	t0 = time.Now()
	for i := 0; i < 10; i++ {
		w.Loop()
	}
	tdNs = time.Now().Sub(t0).Nanoseconds()
	if tdNs < 5_000_000*10 {
		t.Fatal("It should have waited")
	}

	// once one is processed, it becomes fast spin again
	t0 = time.Now()
	w.Reset()
	for i := 0; i < 1_000_000; i++ {
		w.Loop()
	}
	tdNs = time.Now().Sub(t0).Nanoseconds()
	if tdNs > 1_000_000*factor {
		t.Fatalf("It should be way faster: %d > %d\n", tdNs, factor*1_000_000)
	}

}
