package base

import (
	"encoding/json"
	"github.com/kuking/go-frank/api"
)

// converts a Stream of strings into either a Map[string]interface{} or a []interface{}, depending on asMap value.
// If firstRowIsHeader is set, it will pick the column name from the first row, otherwise the column names will be
// sequential numbers starting from 1.
func (s *StreamImpl) CSVasMap(firstRowIsHeader bool, asMap bool) api.Stream {
	return nil
}

// converts a Map[string]interface{} or []interface{} into a comma separated (and escaped if string vs number), adding
// an extra element at the beginning with the column names, if provided
func (s *StreamImpl) MapAsCSV(firstRowIsHeader bool) api.Stream {
	return nil
}

// Unmarshalls a string or an []byte into a map[string]interface{} or an []interface{}. If it can not be parsed as a
// valid JSON object, it will map into a nil.
func (s *StreamImpl) JsonToMap() api.Stream {
	return s.Map(func(val interface{}) interface{} {
		var data []byte
		switch val.(type) {
		case string:
			data = []byte(val.(string))
		case []byte:
			data = val.([]byte)
		default:
			return nil
		}
		var result interface{}
		err := json.Unmarshal(data, &result)
		if err != nil {
			return nil
		}
		return result
	})
}

// Maps the object into a json string representation, if it can not be encoded, it will be encoded into a nil value
func (s *StreamImpl) MapToJson() api.Stream {
	return s.Map(func(val interface{}) interface{} {
		res, err := json.Marshal(val)
		if err != nil {
			return nil
		}
		return string(res)
	})
}
