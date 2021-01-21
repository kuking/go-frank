package go_frank

type StreamSerialiser interface {
	EncodedSize(interface{})
	Encode(interface{}, []byte) error
	Decode([]byte) (interface{}, error)
}
