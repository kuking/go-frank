package highlevel

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInMemorySumMinMaxAvg(t *testing.T) {
	stream := AsNumericStream(givenInMemoryInt64StreamGenerator(4096))
	assert.Equal(t, int64((4096-1)*4096/2), stream.Sum())
	stream = AsNumericStream(givenInMemoryInt64StreamGenerator(4096))
	assert.Equal(t, int64(0), stream.Min())
	stream = AsNumericStream(givenInMemoryInt64StreamGenerator(4096))
	assert.Equal(t, int64(4095), stream.Max())
	stream = AsNumericStream(givenInMemoryInt64StreamGenerator(4096))
	assert.InEpsilon(t, 2048.0, stream.Avg(), 0.01)
}

func TestPersistentSum(t *testing.T) {
	// TODO
}

func TestInMemoryAllMatches(t *testing.T) {
	assert.True(t, givenInMemoryInt64StreamGenerator(4096).Skip(1).AllMatch(func(i int64) bool { return i > 0 }))
	assert.False(t, givenInMemoryInt64StreamGenerator(4096).Skip(0).AllMatch(func(i int64) bool { return i > 0 }))
}

func TestPersistentAllMatches(t *testing.T) {
	// TODO
}

func TestInMemoryNoneMatches(t *testing.T) {
	assert.True(t, givenInMemoryInt64StreamGenerator(4096).Skip(1).NoneMatch(func(i int64) bool { return i < 1 }))
	assert.False(t, givenInMemoryInt64StreamGenerator(4096).NoneMatch(func(i int64) bool { return i < 1 }))
}

func TestPersistentNoneMatches(t *testing.T) {
	//TODO
}

func TestInMemoryAtLeastOne(t *testing.T) {
	assert.True(t, givenInMemoryInt64StreamGenerator(4096).AtLeastOne(func(i int64) bool { return i == 2000 }))
	assert.False(t, givenInMemoryInt64StreamGenerator(4096).Skip(2001).AtLeastOne(func(i int64) bool { return i == 2000 }))
}

func TestPersistentAtLeastOne(t *testing.T) {
	//TODO
}

func TestInMemoryFirstLast(t *testing.T) {
	assert.Equal(t, int64(4000), givenInMemoryInt64StreamGenerator(4096).Skip(4000).First().OrPanic()) // even after skipping
	assert.Equal(t, int64(4095), givenInMemoryInt64StreamGenerator(4096).Last().OrPanic())
}
