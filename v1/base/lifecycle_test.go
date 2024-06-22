package base

import (
	"github.com/kuking/go-frank/v1/api"
	"testing"
)

// missing tests
// i.e. feeding data at the head of the stream, or the tail, or something in between, it should work.

func TestCloseFeedWorksInChain(t *testing.T) {

	for c := 0; c < 4; c++ {
		s := make([]api.Stream, 4)

		s[0] = EmptyStream(1024)
		s[1] = s[0].Skip(16)
		s[2] = s[1].Map(func(i int) int { return i + 1 })
		s[3] = s[2].Reduce(func(a, b int) int { return a + b })

		for i := 0; i < len(s); i++ {
			if s[i].IsClosed() {
				t.Fatal()
			}
		}

		for i := 0; i < 100; i++ {
			s[c].Feed(i)
		}
		go s[c].Close() // close the c'th in the chain so it can be fully consumed

		s[3].Count() // consumes so it should be closed fully

		for i := 0; i < len(s); i++ {
			if !s[i].IsClosed() {
				t.Fatal()
			}
		}
	}

}
