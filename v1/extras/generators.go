package extras

import (
	frank "github.com/kuking/go-frank/v1"
	api2 "github.com/kuking/go-frank/v1/api"
)

func Int64Generator(start, stop int64) api2.Stream {
	count := start
	return frank.StreamGenerator(func() api2.Optional {
		count++
		if count <= stop {
			return api2.OptionalOf(count - 1)
		}
		return api2.EmptyOptional()
	})
}
