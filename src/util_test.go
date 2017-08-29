package hessian

import (
  "testing"
)

func TestParseFloat64FromBytes(t *testing.T) {
  code := []byte{0x3f, 0xf1, 0xf9, 0xad, 0xbb, 0x8f, 0x8d, 0xa7}
  ret := parseFloat64FromBytes(code)
  if ret != 1.1234567 {
    t.Errorf("parseFloat64FromBytes error")
  }
}
