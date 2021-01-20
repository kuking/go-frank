package go_frank

import "testing"

const (
	JSON_EMPLOYEE = `
{
	"employee": {
	"name":       "sonoo",
	"salary":      56000,
	"married":    true
	}
}
`
)

func testJson(t *testing.T) {
	s := ArrayStream([]string{JSON_EMPLOYEE}).
		JsonToMap().
		Map(func(json map[string]interface{}) map[string]interface{} {
			return json
		}).
		MapToJson().
		AsArray()

	if len(s) != 1 || s[0] != "zc" {
		t.Fatal(s)
	}

}
