package go_frank

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

	slice := buf[:0]

	if err = bas.Encode(value, slice[:]); err != nil {
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
	if "lala" != asString(recovered) {
		t.Fatal(fmt.Sprintf("lala != %v", recovered))
	}
}
