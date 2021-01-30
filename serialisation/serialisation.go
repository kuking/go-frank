package serialisation

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
)

// TODO: this has to handle nil serialisation

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

func (g GobSerialiser) EncodedSize(elem interface{}) (size uint64, err error) {
	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(&elem)
	if err != nil {
		return 0, err
	}
	return uint64(buf.Len()), nil
}

func (g GobSerialiser) Encode(elem interface{}, slice []byte) (err error) {
	return gob.NewEncoder(&WritableSlice{slice: slice[:]}).Encode(&elem)
}

func (g GobSerialiser) Decode(slice []byte) (elem interface{}, err error) {
	buf := bytes.NewBuffer(slice)
	err = gob.NewDecoder(buf).Decode(&elem)
	return
}

type ByteArraySerialiser struct{}

func (s ByteArraySerialiser) EncodedSize(elem interface{}) (size uint64, err error) {
	return uint64(len(asByteArray(elem))), nil
}

func (s ByteArraySerialiser) Encode(elem interface{}, buffer []byte) (err error) {
	arr := asByteArray(elem)
	copy(buffer[0:len(arr)], arr[0:len(arr)])
	return nil
}

func (s ByteArraySerialiser) Decode(slice []byte) (elem interface{}, err error) {
	elem = slice
	return elem, nil
}

type Int64Serialiser struct{}

func (i Int64Serialiser) EncodedSize(elem interface{}) (size uint64, err error) {
	return 8, nil
}

func (i Int64Serialiser) Encode(elem interface{}, buffer []byte) (err error) {
	binary.LittleEndian.PutUint64(buffer[:], uint64(elem.(int64)))
	return nil
}

func (i Int64Serialiser) Decode(slice []byte) (elem interface{}, err error) {
	return int64(binary.LittleEndian.Uint64(slice)), err
}

// --------------------------------------------------------------------------------------------------------------------

func asByteArray(elem interface{}) []byte {
	switch elem.(type) {
	case string:
		return []byte(elem.(string))
	case []byte:
		return elem.([]byte)
	default:
		panic("can not use this serialiser with anything but string or []byte")
	}
}

func AsString(elem interface{}) string {
	switch elem.(type) {
	case string:
		return elem.(string)
	case []byte:
		return string(elem.([]byte))
	default:
		return fmt.Sprint(elem)
	}
}
