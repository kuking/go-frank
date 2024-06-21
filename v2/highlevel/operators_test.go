package highlevel

import (
	"github.com/kuking/go-frank/v2/api"
	"github.com/stretchr/testify/assert"
	"testing"
)

func AssertForEach(t *testing.T, stream api.Stream[int64]) {
	var count int64
	stream.ForEach(func(t int64) {
		count += t
	})
	assert.Equal(t, int64((4096-1)*4096/2), count)
}

func TestInMemoryForEach(t *testing.T) {
	AssertForEach(t, givenInMemoryInt64StreamGenerator(4096))
}

func TestPersistentForEach(t *testing.T) {
	// TODO
}
