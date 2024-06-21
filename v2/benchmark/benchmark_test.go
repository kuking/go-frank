package benchmark

import (
	v2 "github.com/kuking/go-frank/v2"
	"testing"
)

func BenchmarkSumInt64(b *testing.B) {
	capacity := 1_000_000
	stream := v2.AsNumericStream(v2.InMemoryStream[int64](capacity + 1))
	for i := 0; i < capacity; i++ {
		stream.Feed(int64(i))
	}
	stream.Close()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stream.Reset()
		if stream.Sum() == int64(capacity*(capacity-1)) {
			b.Fatal()
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(capacity*b.N), "ns/elem")
	b.ReportMetric(float64(capacity), "elems/op")
}
