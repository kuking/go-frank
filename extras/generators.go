package extras

import (
	frank "github.com/kuking/go-frank"
	"github.com/kuking/go-frank/api"
)

func Int64Generator(start, stop int64) api.Stream {
	count := start
	return frank.StreamGenerator(func() api.Optional {
		count++
		if count <= stop {
			return api.OptionalOf(count - 1)
		}
		return api.EmptyOptional()
	})
}
