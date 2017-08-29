package hessian

import (
	"fmt"
	"strings"
	"testing"
)

func unexpected_error(err error, t *testing.T) {
	if err != nil {
		t.FailNow()
	}
}

func TestReadInt(t *testing.T) {
	code := []byte{0x49, 0x00, 0x00, 0x00, 0x01}
	decoder := NewDecoder(code)
	n, err := decoder.ReadInt()
	unexpected_error(err, t)
	if n != 1 {
		t.Errorf("bytes: 4900000001 should be decoded to 1")
	}
	{
		code := []byte{0xd0, 0x00, 0x00}
		decoder := NewDecoder(code)
		n, err := decoder.ReadInt()
		unexpected_error(err, t)
		if n != -262144 {
			t.Errorf("bytes: 0xd00000 should be decoded to -262144")
		}
	}
}

func TestReadLong(t *testing.T) {
  // 1L
	{
		code := []byte{0xe1}
		decoder := NewDecoder(code)
		n, err := decoder.ReadLong()
		unexpected_error(err, t)
		fmt.Printf("ret is: %d\n", n)
		if n != 1 {
			t.Errorf("bytes: 0xdf should be decoded to -1")
		}
	}
	// -1L
	{
		code := []byte{0xdf}
		decoder := NewDecoder(code)
		n, err := decoder.ReadLong()
		unexpected_error(err, t)
		fmt.Printf("ret is: %d\n", n)
		if n != -1 {
			t.Errorf("bytes: 0xdf should be decoded to -1")
		}
	}
	// -1024L
	{
		code := []byte{0xf4, 0x00}
		decoder := NewDecoder(code)
		n, err := decoder.ReadLong()
		unexpected_error(err, t)
		if n != -1024 {
			t.Errorf("bytes: 0xdf should be decoded to -1024")
		}
	}
}

func TestReadBoolean(t *testing.T) {
	code := []byte{0x54, 0x46, 0x47}
	decoder := NewDecoder(code)
	r, err := decoder.ReadBoolean()
	unexpected_error(err, t)
	if r != true {
		t.Errorf("byte 0x54 should be True")
	}
	r, err = decoder.ReadBoolean()
	unexpected_error(err, t)
	if r != false {
		t.Errorf("byte 0x46 should be False")
	}
	_, err = decoder.ReadBoolean()
	if err == nil {
		t.Errorf("readBoolean should reture error")
	}
}

func TestReadString(t *testing.T) {
	code := []byte{0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f}
	decoder := NewDecoder(code)
	s, err := decoder.ReadString()
	unexpected_error(err, t)
	if strings.Compare(s, "hello") != 0 {
		t.Errorf("readString: decoder error")
	}
	// utf_8
	code = []byte{0x02, 0xe4, 0xbd, 0xa0, 0xe5, 0xa5, 0xbd}
	decoder = NewDecoder(code)
	s, err = decoder.ReadString()
	unexpected_error(err, t)
	if strings.Compare(s, "你好") != 0 {
		t.Errorf("readString: decoder error")
	}
	code = []byte{0x52, 0x00, 0x01, 0xe4, 0xbd, 0xa0, 0x53, 0x00, 0x01, 0xe5, 0xa5, 0xbd}
	decoder = NewDecoder(code)
	s, err = decoder.ReadString()
	unexpected_error(err, t)
	if s != "你好" {
		t.Errorf("readString: decoder error")
	}
}

func TestReadDouble(t *testing.T) {
	{
		code := []byte{0x44, 0x3f, 0xf1, 0xf9, 0xad, 0xbb, 0x8f, 0x8d, 0xa7}
		decoder := NewDecoder(code)
		f, err := decoder.ReadDouble()
		unexpected_error(err, t)
		if f != 1.1234567 {
			t.Errorf("readDouble: decoder error")
		}
	}
	{
		code := []byte{0x5b}
		decoder := NewDecoder(code)
		f, err := decoder.ReadDouble()
		unexpected_error(err, t)
		if f != 0.0 {
			t.Errorf("readDouble: decoder error")
		}
	}
}
