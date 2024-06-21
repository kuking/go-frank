package highlevel

import (
	"github.com/kuking/go-frank/v2/api"
	"github.com/stretchr/testify/assert"
	"testing"
)

func AssertSkip512Count512(t *testing.T, stream api.Stream[int64]) {
	assert.Equal(t, 512, stream.Skip(512).Count())
}

func AssertCountIs1024(t *testing.T, stream api.Stream[int64]) {
	assert.Equal(t, 1024, stream.Count())
}

func AssertSkip1024Following1024Equals(t *testing.T, stream api.Stream[int64]) {
	stream = stream.Skip(1024)
	assert.False(t, stream.IsClosed())
	for i := 1024; i < 2048; i++ {
		assert.Equal(t, int64(i), stream.Pull().OrPanic())
	}
	assert.True(t, stream.IsClosed())
	assert.True(t, stream.Pull().IsEmpty())
	assert.True(t, stream.IsClosed())
}

func TestInMemorySkipCountClose(t *testing.T) {
	AssertSkip512Count512(t, givenInMemoryInt64StreamGenerator(1024))
	AssertCountIs1024(t, givenInMemoryInt64StreamGenerator(1024))
	AssertSkip1024Following1024Equals(t, givenInMemoryInt64StreamGenerator(2048))
}

func TestPersistentSkipCountClose(t *testing.T) {
	//XXX TODO
}

func AssertPrevOperations(t *testing.T, stream api.Stream[int64]) {
	assert.False(t, stream.Prev()) // can not go backward on a stream at the beginning of it
	assert.Equal(t, int64(0), stream.Pull().OrPanic())
	assert.Equal(t, int64(1), stream.Pull().OrPanic())
	assert.Equal(t, int64(2), stream.Pull().OrPanic())
	assert.True(t, stream.Prev())
	assert.Equal(t, int64(2), stream.Pull().OrPanic())
	assert.True(t, stream.Prev())
	assert.True(t, stream.Prev())
	assert.True(t, stream.Prev())
	assert.False(t, stream.Prev())
	assert.Equal(t, int64(0), stream.Pull().OrPanic())
	assert.Equal(t, int64(1), stream.Pull().OrPanic())
	assert.Equal(t, int64(2), stream.Pull().OrPanic())
}

func TestInMemoryPrev(t *testing.T) {
	AssertPrevOperations(t, givenInMemoryInt64StreamGenerator(10))
}

func TestPersistentPrev(t *testing.T) {
	// TODO
}

func AssertSkipWhileMatching(t *testing.T, stream api.Stream[int64]) {
	assert.Equal(t, 512, stream.SkipWhile(func(i int64) bool { return i < 512 }).Count())
}

func TestInMemorySkipWhileMatching(t *testing.T) {
	AssertSkipWhileMatching(t, givenInMemoryInt64StreamGenerator(1024))
}

func TestPersistentSkipWhileMatching(t *testing.T) {
	// TODO
}

func AssertFindFirst(t *testing.T, stream api.Stream[int64]) {
	assert.Equal(t, 1024-78, stream.FindFirst(func(i int64) bool { return i == 78 }).Count())
}

func TestInMemoryFindFirst(t *testing.T) {
	AssertFindFirst(t, givenInMemoryInt64StreamGenerator(1024))
}

func TestPersistentFindFirst(t *testing.T) {
	// TODO
}

func AssertFilter(t *testing.T, stream api.Stream[int64]) {
	assert.Equal(t, 1024/3+1, stream.Filter(func(i int64) bool { return i%3 == 0 }).Count())
}

func TestInMemoryFilter(t *testing.T) {
	AssertFilter(t, givenInMemoryInt64StreamGenerator(1024))
}

func TestPersistentFilter(t *testing.T) {
	// TODO
}
