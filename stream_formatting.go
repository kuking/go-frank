package go_frank

import "encoding/json"

// converts a Stream of strings into either a Map[string]interface{} or a []interface{}, depending on asMap value.
// If firstRowIsHeader is set, it will pick the column name from the first row, otherwise the column names will be
// sequential numbers starting from 1.
func (s *streamImpl) CSVasMap(firstRowIsHeader bool, asMap bool) Stream {
	return nil
}

// converts a Map[string]interface{} or []interface{} into a comma separated (and escaped if string vs number), adding
// an extra element at the beginning with the column names, if provided
func (s *streamImpl) MapAsCSV(firstRowIsHeader bool) Stream {
	return nil
}

func (s *streamImpl) JsonToMap() Stream {
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

func (s *streamImpl) MapToJson() Stream {
	ns := streamImpl{
		closed: 0,
		prev:   s,
	}
	ns.pull = func(n *streamImpl) (read interface{}, closed bool) {
		read, closed = ns.prev.pull(ns.prev)
		for ; !closed; {
			if read != nil {
				res, err := json.Marshal(read)
				if err != nil {
					return nil, closed
				}
				return string(res), closed
			}
			read, closed = ns.prev.pull(ns.prev)
		}
		return
	}
	return &ns
}
