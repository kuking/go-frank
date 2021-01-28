package base

import "github.com/kuking/go-frank/api"

func givenInt64StreamGenerator(total int) api.Stream {
	count := int64(0)
	return StreamGenerator(func() api.Optional {
		count++
		if count <= int64(total) {
			return api.OptionalOf(count)
		}
		return api.EmptyOptional()
	})
}

func givenInt64ArrayStream(l int) api.Stream {
	arr := make([]interface{}, l)
	for i := 0; i < l; i++ {
		arr[i] = int64(i)
	}
	return ArrayStream(arr)
}


func givenStringArrayStream() api.Stream {
	return ArrayStream([]interface{}{"Hello", "how", "are", "you", "doing", "?"})
}

