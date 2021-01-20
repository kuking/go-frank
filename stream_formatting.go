package go_frank

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
	return nil
}
func (s *streamImpl) MapToJson() Stream {
	return nil
}
