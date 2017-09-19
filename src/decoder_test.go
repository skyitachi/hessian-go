package hessian

import (
	"fmt"
	"strings"
	"testing"
  "time"
  "log"
  "reflect"
)

func unexpected_error(err error, t *testing.T) {
	if err != nil {
    log.Println(err)
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
  // -1
  {
    code := []byte{0x8f}
    decoder := NewDecoder(code)
    n, err := decoder.ReadInt()
    unexpected_error(err, t)
    if n != -1 {
      t.Errorf("bytes: 0x8f should be decoded to -1")
    }
  }
  // 65536
  {
    code := []byte{0xd5, 0x00, 0x00}
    decoder := NewDecoder(code)
    n, err := decoder.ReadInt()
    unexpected_error(err, t)
    if n != 65536 {
      t.Errorf("bytes: 0x8f should be decoded to -1")
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
			t.Errorf("bytes: 0xdf should be decoded to 1")
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

func TestReadDate(t *testing.T) {
	{
		code := []byte{0x4a, 0x00, 0x00, 0x01, 0x5e, 0x31, 0x6b, 0xe5, 0xce}
		decoder := NewDecoder(code)
    ti, err := decoder.ReadDate()
		unexpected_error(err, t)
    y, m, d := ti.Date()
    if y != 2017 || m != time.August || d != 30 {
      t.Errorf("readDate: decoder error")
    }
	}
  {
    code := []byte{0x4b, 0x00, 0xb2, 0xe6, 0xc0}
    decoder := NewDecoder(code)
    ti, err := decoder.ReadDate()
    unexpected_error(err, t)
    y, m, d := ti.Date()
    fmt.Println(y, m, d)
    if y != 1992 || m != time.April || d != 17 {
      t.Errorf("readDate: decoder error")
    }
  }
}

func TestReadType(t *testing.T) {
  {
    code := []byte{0x03, 0x43, 0x61, 0x72, 0x90}
    decoder := NewDecoder(code)
    t1, err := decoder.ReadType()
    unexpected_error(err, t)
    if t1 != "Car" {
      t.Errorf("readType: decoder error")
    }
    t2, err := decoder.ReadType()
    unexpected_error(err, t)
    if t2 != "Car" {
      t.Errorf("readType: decoder error")
    }
  }
}

func TestReadNull(t *testing.T) {
  {
    code := []byte{0x4e}
    decoder := NewDecoder(code)
    r, err := decoder.ReadNull()
    unexpected_error(err, t)
    if r != nil {
      t.Errorf("readNull:decoder error")
    }
  }
}

func TestReadList(t *testing.T) {
  {
    originValue := []int32{1}
    code := []byte{0x71,0x12,0x5b,0x6a,0x61,0x76,0x61,0x2e,0x6c,0x61,0x6e,0x67,0x2e,0x49,0x6e,0x74,0x65,0x67,0x65,0x72,0x91}
    decoder := NewDecoder(code)
    ret, err := decoder.ReadList()
    unexpected_error(err, t)
    if len(ret.Value) != len(originValue) {
      t.Errorf("readList: decoder error")
    }

    for idx, v := range ret.Value {
      if v != originValue[idx] {
        t.Errorf("readList: decoder error")
      }
    }
  }
  {
    originValue := []int32{1, -1, 65536}
    code := []byte{0x73, 0x04, 0x5b, 0x69, 0x6e, 0x74, 0x91,0x8f, 0xd5, 0x00, 0x00}
    decoder := NewDecoder(code)
    ret, err := decoder.ReadList()
    unexpected_error(err, t)
    if len(ret.Value) != len(originValue) {
      t.Errorf("readList: decoder error")
    }
    for idx, v := range ret.Value {
      fmt.Println("v is ", v)
      if v != originValue[idx] {
        t.Errorf("readList: decoder error")
      }
    }
  }
}

func TestReadMap(t *testing.T) {
  {
    /**
       map[int32]string{
        1: "hello"
        2: "hello"
       }
     */
    code := []byte{0x48,0x91,0x05,0x68,0x65,0x6C,0x6C,0x6F,0x92,0x05,0x68,0x65,0x6C,0x6C,0x6F,0x5A}
    decoder := NewDecoder(code)
    ret, err := decoder.ReadMap()
    unexpected_error(err, t)
    for k, v:= range ret {
      if reflect.TypeOf(k).Name() != "int32" {
        t.Errorf("readMap: decoder error")
      }
      if reflect.TypeOf(v).Name() != "string" {
        t.Errorf("readMap: decoder error")
      }
      kTyped := reflect.ValueOf(k).Interface().(int32)
      vTyped := reflect.ValueOf(v).Interface().(string)
      if kTyped == 1 && vTyped != "hello" {
        t.Errorf("readMap: decoder error")
      }
      if kTyped == 2 && vTyped != "hello" {
        t.Errorf("readMap: decoder error")
      }
    }
  }
}
