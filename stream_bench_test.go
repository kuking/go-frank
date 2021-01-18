package go_frank

import (
	"testing"
)

func BenchmarkSum(b *testing.B) {
	count := 1000
	times := 1
	exp := count * (count - 1) / 2
	arr := givenArray(count)
	for i := 0; i < times; i++ {
		if res, ok := ArrayStream(arr).Sum().First(); !ok || res != exp {
			b.Fatal(res)
		}
	}
	b.N = count * times
	// 1000:1000 = 1338ns/op  (ArrayStreamSlow)
	// 10000:10000 = 223ns/op (ArrayStream)
}

func givenArray(count int) []interface{} {
	elems := make([]interface{}, count)
	for i := 0; i < len(elems); i++ {
		elems[i] = i
	}
	return elems
}
