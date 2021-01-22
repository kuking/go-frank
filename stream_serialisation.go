package go_frank

import (
	"bytes"
	"encoding/gob"
)

type StreamSerialiser interface {
	EncodedSize(elem interface{}) (size uint64, err error)
	Encode(elem interface{}, buffer []byte) (err error)
	Decode(slice []byte) (elem interface{}, err error)
}

type WritableSlice struct {
	slice []byte
}

func (w *WritableSlice) Write(p []byte) (n int, err error) {
	w.slice = append(w.slice, p...)
	return n, nil
}

// this is very inefficient, allocates and encodes twice (for EncodedSize, etc.) -- we will come back to this
type GobSerialiser struct{}

func (g *GobSerialiser) EncodedSize(elem interface{}) (size uint64, err error) {
	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(elem)
	if err != nil {
		return 0, err
	}
	return uint64(buf.Len()), nil
}

func (g *GobSerialiser) Encode(elem interface{}, slice []byte) (err error) {
	return gob.NewEncoder(&WritableSlice{slice: slice[:]}).Encode(&elem)
}

func (g *GobSerialiser) Decode(slice []byte) (elem interface{}, err error) {
	buf := bytes.NewBuffer(slice)
	err = gob.NewDecoder(buf).Decode(&elem)
	return
}
