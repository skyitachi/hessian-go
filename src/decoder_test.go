package hessian

import (
  "testing"
  "fmt"
  "strings"
)

func unexpected_error(err error, t* testing.T) {
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
  fmt.Printf("ans is %s, size is %d\n", s, len(s))
  if strings.Compare(s,"hello") != 0 {
    t.Errorf("readString: decoder error")
  }
  // utf_8
  code = []byte{0x02, 0xe4, 0xbd, 0xa0, 0xe5, 0xa5, 0xbd}
  decoder = NewDecoder(code)
  s, err = decoder.ReadString()
  unexpected_error(err, t)
  fmt.Println(len(s))
  if strings.Compare(s, "你好") != 0 {
    t.Errorf("readString: decoder error")
  }
}
