package go_frank

import (
	"testing"
)

func TestSimplest(t *testing.T) {
	s := EmptyStream(256)
	s.Feed("Hello")
	go s.Close()

	result := s.AsArray()
	if len(result) != 1 || result[0] != "Hello" {
		t.Fatal()
	}
}

func TestSimple(t *testing.T) {
	result := givenStringArrayStream().AsArray()
	if len(result) != 6 || result[2] != "are" {
		t.Fatal()
	}
}

func TestReduce(t *testing.T) {
	result := givenStringArrayStream().
		Reduce(func(l, r string) string { return l + " " + r }).
		First()
	if result.isEmpty() || result.Get().(string) != "Hello how are you doing ?" {
		t.Fatal(result.Get())
	}
	// one element reduce
	if givenInt64ArrayStream(1).Reduce(func(l, r int64) int64 { return l + r }).First().Get() != int64(0) {
		t.Fatal()
	}
	// zero elements reducer
	if givenInt64ArrayStream(0).Reduce(func(l, r int64) int64 { return l + r }).First().isPresent() {
		t.Fatal()
	}
	// one element reducerNA
	if givenInt64ArrayStream(1).ReduceNA(&Int64SumReducer{}).First().Get() != int64(0) {
		t.Fatal()
	}
	// zero elements reducerNA
	if givenInt64ArrayStream(0).ReduceNA(&Int64SumReducer{}).First().isPresent() {
		t.Fatal()
	}
}

func TestFirstLast(t *testing.T) {
	result := givenStringArrayStream().First()
	if result.isEmpty() || result.Get() != "Hello" {
		t.Fatal()
	}
	result = givenStringArrayStream().Last()
	if result.isEmpty() || result.Get() != "?" {
		t.Fatal()
	}
}

func TestCount(t *testing.T) {
	count := givenStringArrayStream().Count()
	if count != 6 {
		t.Fatal()
	}
}

func TestMap(t *testing.T) {
	result := givenStringArrayStream().
		Map(func(s string) int { return len(s) }).
		Reduce(func(l, r int) int { return l + r }).
		First()
	if result.isEmpty() || result.Get() != 20 {
		t.Fatal()
	}
}

func TestMapInt64(t *testing.T) {
	result := givenInt64ArrayStream(123).
		MapInt64(func(i int64) int64 { return i + 1 }).
		Reduce(func(l, r int64) int64 { return l + r }).
		First()
	if result.isEmpty() || result.Get() != int64(((123*(123-1))/2)+123) {
		t.Fatal(result.Get())
	}
}

func TestSumInt64(t *testing.T) {
	result := givenStringArrayStream().
		Map(func(s string) int64 { return int64(len(s)) }).
		SumInt64().
		First()
	if !result.isPresent() || result.Get() != int64(20) {
		t.Fatal()
	}
}

func TestFilter(t *testing.T) {
	result := givenStringArrayStream().
		Filter(func(s string) bool { return len(s) == 3 }).
		Reduce(func(l, r string) string { return l + " " + r }).
		First()
	if result.isEmpty() || result.Get() != "Hello doing ?" {
		t.Fatal(result)
	}
	// filterNA
	result = givenStringArrayStream().
		FilterNA(func(s interface{}) bool { return len(s.(string)) == 3 }).
		Reduce(func(l, r string) string { return l + " " + r }).
		First()
	if result.isEmpty() || result.Get() != "Hello doing ?" {
		t.Fatal(result)
	}
}

func TestEmptyFirstLast(t *testing.T) {
	if givenInt64ArrayStream(0).Last().isPresent() {
		t.Fatal()
	}
	if givenInt64ArrayStream(0).First().isPresent() {
		t.Fatal()
	}
}

func givenStringArrayStream() *streamImpl {
	return ArrayStream([]interface{}{"Hello", "how", "are", "you", "doing", "?"})
}

func givenInt64ArrayStream(l int) *streamImpl {
	arr := make([]interface{}, l)
	for i := 0; i < l; i++ {
		arr[i] = int64(i)
	}
	return ArrayStream(arr)
}
