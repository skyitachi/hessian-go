package hessian

import "testing"

func unexpected_error(err error, t* testing.T) {
  if err != nil {
    t.FailNow()
  }
}

func TestReadInt(t *testing.T) {
  code := []byte{0x49, 0x00, 0x00, 0x00, 0x01}
  decoder := NewDecoder(code)
  n, err := decoder.readInt()
  unexpected_error(err, t)
  if n != 1 {
    t.Errorf("bytes: 4900000001 should be decoded to 1")
  }
}

func TestReadBoolean(t *testing.T) {
  code := []byte{0x54, 0x46, 0x47}
  decoder := NewDecoder(code)
  r, err := decoder.readBoolean()
  unexpected_error(err, t)
  if r != true {
    t.Errorf("byte 0x54 should be True")
  }
  r, err = decoder.readBoolean()
  unexpected_error(err, t)
  if r != false {
    t.Errorf("byte 0x46 should be False")
  }
  _, err = decoder.readBoolean()
  if err == nil {
    t.Errorf("readBoolean should reture error")
  }
}
