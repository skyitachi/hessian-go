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

func TestParseInt64FromBytes(t *testing.T) {
  {
    // 1000000001
    code := []byte{0x3b, 0x9a, 0xca, 0x01}
    ret := parseInt64FromBytes(code)
    if ret != 1000000001 {
      t.Errorf("parseInt64FromBytes error: expect 1000000001 found %d", ret)
    }
  }

}
