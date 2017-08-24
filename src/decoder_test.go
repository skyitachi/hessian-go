package hessian

import "testing"

func TestReadInt(t *testing.T) {
  code := []byte{0x49, 0x00, 0x00, 0x00, 0x01}
  decoder := NewDecoder(code)
  n, err := decoder.readInt()
  if err != nil {
    t.FailNow()
  }
  if n != 1 {
    t.Errorf("bytes: 4900000001 should be decoded to 1")
  }
}
