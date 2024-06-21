package highlevel

import (
	"github.com/kuking/go-frank/v2/api"
)

func givenInMemoryInt64StreamGenerator(total int) api.Stream[int64] {
	count := int64(0)
	s := InMemoryStream[int64](256)
	go func() {
		for count < int64(total) {
			s.Feed(count)
			count++
		}
		s.Close()
	}()
	return s
}
