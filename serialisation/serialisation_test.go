package serialisation

import (
	"encoding/gob"
	"fmt"
	"reflect"
	"testing"
)

func TestGobSerialiser(t *testing.T) {

	buf := [1024]byte{}
	value := map[string]string{"a": "b"}

	gobe := GobSerialiser{}
	gob.Register(map[string]string{})

	size, err := gobe.EncodedSize(value)
	if err != nil {
		t.Fatal(err)
	}
	if size != 45 {
		t.Fatal(size)
	}

	slice := buf[:0]

	if err = gobe.Encode(value, slice); err != nil {
		t.Fatal(err)
	}

	recovered, err := gobe.Decode(buf[0:size])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(value, recovered) {
		t.Fatal(fmt.Sprintf("%v != %v", value, recovered))
	}
}

func TestGobForInt64(t *testing.T) {
	buf := [1024]byte{}
	value := int64(42)

	gobe := GobSerialiser{}

	size, err := gobe.EncodedSize(value)
	if err != nil {
		t.Fatal(err)
	}
	if size != 13 {
		t.Fatal(size)
	}

	if err = gobe.Encode(value, buf[:0]); err != nil {
		t.Fatal(err)
	}

	recovered, err := gobe.Decode(buf[0:size])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(value, recovered) {
		t.Fatal(fmt.Sprintf("%v != %v", value, recovered))
	}
}

func TestByteArraySerialiser(t *testing.T) {

	buf := [1024]byte{}
	value := []byte("hello")

	bas := ByteArraySerialiser{}

	size, err := bas.EncodedSize(value)
	if err != nil {
		t.Fatal(err)
	}
	if size != 5 {
		t.Fatal(size)
	}

	if err = bas.Encode(value, buf[0:size]); err != nil {
		t.Fatal(err)
	}

	recovered, err := bas.Decode(buf[0:size])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(value, recovered) {
		t.Fatal(fmt.Sprintf("%v != %v", value, recovered))
	}
}

func TestByteArraySerialiser_Strings(t *testing.T) {
	bas := ByteArraySerialiser{}
	if size, err := bas.EncodedSize("lala"); size != 4 || err != nil {
		t.Fatal()
	}
	buf := [1024]byte{}
	slice := buf[:0]
	if err := bas.Encode("lala", slice[:]); err != nil {
		t.Fatal(err)
	}
	recovered, err := bas.Decode(buf[0:4])
	if err != nil {
		t.Fatal(err)
	}
	if "lala" != AsString(recovered) {
		t.Fatal(fmt.Sprintf("lala != %v", recovered))
	}
}

func TestInt64Serialiser(t *testing.T) {

	buf := [1024]byte{}
	value := int64(4242)

	bas := Int64Serialiser{}

	size, err := bas.EncodedSize(value)
	if err != nil {
		t.Fatal(err)
	}
	if size != 8 {
		t.Fatal(size)
	}

	if err = bas.Encode(value, buf[0:size]); err != nil {
		t.Fatal(err)
	}

	recovered, err := bas.Decode(buf[0:size])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(value, recovered) {
		t.Fatal(fmt.Sprintf("%v != %v", value, recovered))
	}
}

func TestNTStrings(t *testing.T) {
	buf := [10]byte{}

	ToNTString(buf[:], "hello")
	if FromNTString(buf[:]) != "hello" {
		t.Fatal()
	}
	buf[2] = 0
	if FromNTString(buf[:]) != "he" {
		t.Fatal()
	}
	ToNTString(buf[:], "1234567890ABCDEF")
	if FromNTString(buf[:]) != "1234567890" {
		t.Fatal()
	}

}
