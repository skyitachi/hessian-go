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

func (decoder *Decoder) read_n_rune(n int) string {
  var ret []rune
  for {
    r, _, err := decoder.buf.ReadRune()
    if err != nil {
      decoder.buf.UnreadRune()
      break
    }
    ret = append(ret, r)
    n--
    if n <= 0 {
      break
    }
  }
  return string(ret)
}

func (decoder *Decoder) ReadInt() (int, error) {
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

func (decoder Decoder) ReadBoolean() (bool, error) {
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

/**
 * string ::= x52 b1 b0 <utf8-data> string
 *        ::= S b1 b0 <utf8-data>
 *        ::= [x00-x1f] <utf8-data>
 *        ::= [x30-x33] b0 <utf8-data>
 */
func (decoder Decoder) ReadString() (string, error) {
  code, err := decoder.read()
  if err != nil {
    return "", err
  }
  switch {
  case code == 0x53 || code == 0x52:
    bits := decoder.readn(2)
    size := int(bits[0]<<8 + bits[1])
    ret := decoder.read_n_rune(size)
    if len(ret) < size {
      return "", errors.New("readString error: unexpected length")
    }
    return ret, nil
  case code >= 0x00 && code <= 0x1f:
    size := int(code)
    ret := decoder.read_n_rune(size)
    if len(ret) < size {
      return "", errors.New("readString error: unexpected length")
    }
    return ret, nil
  case code >= 0x30 && code <= 0x33:
    bits := decoder.readn(1)
    if len(bits) < 1 {
      return "", errors.New("readString error: unexpected length")
    }
    size := int(code + bits[0])
    ret := decoder.read_n_rune(size)
    if len(ret) < size {
      return "", errors.New("readString error: unexpected length")
    }
    return ret, nil
  default:
    return "", errors.New("readString error: unexpected code")
  }
}

