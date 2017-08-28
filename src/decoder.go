package hessian

import (
  "bytes"
  "github.com/pkg/errors"
)
type Decoder struct {
  buf *bytes.Buffer
}

func NewDecoder(b []byte) *Decoder {
  return &Decoder{bytes.NewBuffer(b)}
}

func (decoder *Decoder) read() (byte, error){
  return decoder.buf.ReadByte()
}

func (decoder *Decoder) readn(n int) []byte {
  return decoder.buf.Next(n)
}

func (decoder *Decoder) readInt() (int, error) {
  code, err := decoder.read()
  if err != nil {
    return 0, err
  }
  if code == 0x49 {
    bits := decoder.readn(4)
    if len(bits) < 4 {
      return 0, errors.New("readInt error: unexpected length of bytes")
    }
    return int((bits[0] << 24) + (bits[1] << 16) + (bits[2] << 8) + bits[3]), nil
  }
  if code >= 0x80 && code <= 0xbf {
    return int(code - 0x80), nil
  }
  if code >= 0xc0 && code <= 0xcf {
    b0, err := decoder.read()
    if err != nil {
      return 0, err
    }
    return int(((code - 0xc8) << 8) + b0), err
  }
  if code >= 0xd0 && code <= 0xd7 {
    bits := decoder.readn(2)
    if len(bits) < 2 {
      return 0, errors.New("readInt error: unexpected length of bytes")
    }
    return int(((code - 0xd4) << 16) + (bits[0] << 8) + bits[1]), nil
  }
  // throw error
  return 0, errors.New("readInt error: " + string(code))
}

func (decoder Decoder) readBoolean() (bool, error) {
  code, err := decoder.read()
  if err != nil {
    return false, err
  }
  switch code {
  case 0x54:
    return true, nil
  case 0x46:
    return false, nil
  default:
    return false, errors.New("readBoolean error: unexpected code, 'T' or 'F' expected")
  }
}
