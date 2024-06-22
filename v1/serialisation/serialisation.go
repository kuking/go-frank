package serialisation

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
)

type StreamSerialiser interface {
	EncodedSize(elem interface{}) (size uint16, err error)
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

func (g GobSerialiser) EncodedSize(elem interface{}) (size uint16, err error) {
	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(&elem)
	if err != nil {
		return 0, err
	}
	return uint16(buf.Len()), nil
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

func (s ByteArraySerialiser) EncodedSize(elem interface{}) (size uint16, err error) {
	return uint16(len(asByteArray(elem))), nil
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

func (i Int64Serialiser) EncodedSize(elem interface{}) (size uint16, err error) {
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

// normal string to null terminated string
func ToNTString(target []byte, s string) {
	copy(target[:], []byte(s))
	if len(s) < len(target) {
		target[len(s)] = 0
	}
}

// null terminated string to normal string
func FromNTString(buf []byte) string {
	for i := 0; i < len(buf); i++ {
		if buf[i] == 0 {
			return string(buf[0:i])
		}
	}
	return string(buf)
}
