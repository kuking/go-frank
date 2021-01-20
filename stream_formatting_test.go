package go_frank

import (
	"testing"
)

const (
	jsonEmployee = `{
	"employee": {
		"name":    "robert",
		"salary":  56000,
		"married": false
	}
}`
)

func TestJson(t *testing.T) {
	s := ArrayStream([]interface{}{jsonEmployee, "not a json", "[]", []byte("{}"), 123}).
		JsonToMap().
		ModifyNA(func(jsonI interface{}) {
			switch jsonI.(type) {
			case map[string]interface{}:
				json := jsonI.(map[string]interface{})
				if len(json) != 0 {
					employee := json["employee"].(map[string]interface{})
					employee["salary"] = 1_000_000
					employee["married"] = false
				}
			}
		}).
		MapToJson().
		AsArray()
	if len(s) != 5 ||
		s[0].(string) != "{\"employee\":{\"married\":false,\"name\":\"robert\",\"salary\":1000000}}" ||
		s[1] != nil ||
		s[2].(string) != "[]" ||
		s[3].(string) != "{}" ||
		s[4] != nil {
		t.Fatal(s)
	}
}
