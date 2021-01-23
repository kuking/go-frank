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



