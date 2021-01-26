package go_frank

import (
	"testing"
)

func TestReduce(t *testing.T) {
	result := givenStringArrayStream().
		Reduce(func(l, r string) string { return l + " " + r }).
		First()
	if result.IsEmpty() || result.Get().(string) != "Hello how are you doing ?" {
		t.Fatal(result.Get())
	}
	// one element reduce
	if givenInt64ArrayStream(1).Reduce(func(l, r int64) int64 { return l + r }).First().Get() != int64(0) {
		t.Fatal()
	}
	// zero elements reducer
	if givenInt64ArrayStream(0).Reduce(func(l, r int64) int64 { return l + r }).First().IsPresent() {
		t.Fatal()
	}
	// one element reducerNA
	if givenInt64ArrayStream(1).ReduceNA(&Int64SumReducer{}).First().Get() != int64(0) {
		t.Fatal()
	}
	// zero elements reducerNA
	if givenInt64ArrayStream(0).ReduceNA(&Int64SumReducer{}).First().IsPresent() {
		t.Fatal()
	}
}

func TestMap(t *testing.T) {
	result := givenStringArrayStream().
		Map(func(s string) int { return len(s) }).
		Reduce(func(l, r int) int { return l + r }).
		First()
	if result.IsEmpty() || result.Get() != 20 {
		t.Fatal()
	}
}

func TestMapInt64(t *testing.T) {
	result := givenInt64ArrayStream(123).
		MapInt64(func(i int64) int64 { return i + 1 }).
		Reduce(func(l, r int64) int64 { return l + r }).
		First()
	if result.IsEmpty() || result.Get() != int64(((123*(123-1))/2)+123) {
		t.Fatal(result.Get())
	}
}

func TestSumInt64(t *testing.T) {
	result := givenStringArrayStream().
		Map(func(s string) int64 { return int64(len(s)) }).
		SumInt64().
		First()
	if !result.IsPresent() || result.Get() != int64(20) {
		t.Fatal()
	}
}

func TestFilter(t *testing.T) {
	result := givenStringArrayStream().
		Filter(func(s string) bool { return len(s) == 3 }).
		Reduce(func(l, r string) string { return l + " " + r }).
		First()
	if result.IsEmpty() || result.Get() != "Hello doing ?" {
		t.Fatal(result)
	}
	// filterNA
	result = givenStringArrayStream().
		FilterNA(func(s interface{}) bool { return len(s.(string)) == 3 }).
		Reduce(func(l, r string) string { return l + " " + r }).
		First()
	if result.IsEmpty() || result.Get() != "Hello doing ?" {
		t.Fatal(result)
	}
}

func TestModifyNA(t *testing.T) {
	result := ArrayStream(map[string]string{"a": "1"}).
		ModifyNA(func(value interface{}) {
			m := value.(map[string]string)
			m["b"] = "2"
		}).
		First()

	if result.IsEmpty() {
		t.Fatal()
	}
	m := result.Get().(map[string]string)
	if m["a"] != "1" || m["b"] != "2" || len(m) != 2 {
		t.Fatal()
	}
}

func TestNilThroughTransformers(t *testing.T) {
	r := ArrayStream([]interface{}{nil, nil, nil}).
		Map(func(i interface{}) interface{} { return nil }).
		ModifyNA(func(i interface{}) {}).
		Filter(func(i interface{}) bool { return i != nil }).
		FilterNA(func(i interface{}) bool { return i != nil }).
		AsArray()

	if len(r) != 3 || r[0] != nil || r[1] != nil || r[2] != nil {
		t.Fatal(r)
	}

}
