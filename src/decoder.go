package hessian

import (
	"bytes"
	"errors"
	"time"
  "fmt"
)
const UNTYPED = "untyped"
var TIME_DEFAULT_VALUE = time.Unix(0, 0)

type List struct {
  ValueType string
  Value []interface{}
}

type Decoder struct {
	buf       *bytes.Buffer
  lastChunk bool
  types []string
  refMap map[int32]interface{}
  refId int32
  byteCount int32 // how many bytes read after last successful read, for recovery
  runeCount int32 // ...
}

func NewDecoder(b []byte) *Decoder {
	return &Decoder{
    buf:bytes.NewBuffer(b),
    lastChunk: false,
    types: []string{},
    refMap: make(map[int32]interface{}),
    refId: 0,
    byteCount: 0,
    runeCount: 0,
  }
}

func (decoder *Decoder) success() {
  decoder.byteCount = 0
  decoder.runeCount = 0
}

func (decoder *Decoder) Recover() {
  // FIXME: have problems 顺序问题
  decoder.unread_n_byte(decoder.byteCount)
  decoder.unread_n_rune(decoder.runeCount)
  decoder.success()
}

func (decoder *Decoder) read() (byte, error) {
  decoder.byteCount += 1
	return decoder.buf.ReadByte()
}

func (decoder *Decoder) readn(n int) []byte {
  bs := decoder.buf.Next(n)
  l := len(bs)
  decoder.byteCount += int32(l)
  return bs
}

func (decoder *Decoder) unread_n_byte(n int32) {
  var i int32
  for i = 0; i < n; i++ {
    decoder.buf.UnreadByte()
  }
}

func (decoder *Decoder) unread_n_rune(n int32) {
  var i int32
  for i = 0; i < n; i++ {
    decoder.buf.UnreadRune()
  }
}

func (decoder *Decoder) read_n_rune(n int) string {
	if n == 0 {
		return ""
	}
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
  decoder.runeCount += int32(len(ret))
	return string(ret)
}

func (decoder *Decoder) readFixedLengthTypedValue(length int, typeName string) ([]interface{}, error) {
  ret := []interface{}{}
  for i := 0; i < length; i++ {
    v, err := dynamic_call(decoder, typeName)
    if err != nil {
      fmt.Println(err)
      return []interface{}{}, errors.New("readFixedLengthTypedValue: unexpected type value")
    }
    ret = append(ret, v)
  }
  return ret, nil
}

func (decoder *Decoder) readFixedLengthUnTypedValue(length int) ([]interface{}, error) {
  ret := []interface{}{}
  for i := 0; i < length; i++ {
    code, err := decoder.read()
    if err != nil {
      return []interface{}{}, err
    }
    v, err := dynamic_call(decoder, CODE_TO_TYPE[code])
    if err != nil {
      return []interface{}{}, err
    }
    ret = append(ret, v)
  }
  return ret, nil
}

func (decoder *Decoder) addRef(ref interface{}) {
  decoder.refMap[decoder.refId] = ref
  decoder.refId++
}

func (decoder *Decoder) ReadInt() (int32, error) {
	code, err := decoder.read()
	if err != nil {
		return 0, err
	}
	if code == 0x49 {
		bits := decoder.readn(4)
		if len(bits) < 4 {
			return 0, errors.New("readInt error: unexpected length of bytes")
		}
    decoder.success()
		return parseInt32FromBytes(bits), nil
	}
	if code >= 0x80 && code <= 0xbf {
    decoder.success()
		return int32(int8(code - 0x90)), nil
	}
	if code >= 0xc0 && code <= 0xcf {
		b0, err := decoder.read()
		if err != nil {
			return 0, err
		}
    decoder.success()
		return parseInt32(int8(code-0xc8), []byte{b0}), nil
	}
	if code >= 0xd0 && code <= 0xd7 {
		bits := decoder.readn(2)
		if len(bits) < 2 {
			return 0, errors.New("readInt error: unexpected length of bytes")
		}
    decoder.success()
		return parseInt32(int8(code-0xd4), bits), nil
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
    decoder.success()
		return true, nil
	case 0x46:
    decoder.success()
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
func (decoder *Decoder) ReadString() (string, error) {
	code, err := decoder.read()
	if err != nil {
		return "", err
	}
	switch {
	case code == 0x53:
		// final chunk
		decoder.lastChunk = true
		bits := decoder.readn(2)
		size := int(bits[0]<<8 + bits[1])
		ret := decoder.read_n_rune(size)
		if len(ret) < size {
			return "", errors.New("readString error: unexpected length")
		}
    decoder.success()
		return ret, nil
	case code == 0x52:
		bits := decoder.readn(2)
		size := int(bits[0]<<8 + bits[1])
		ret := decoder.read_n_rune(size)
		if len(ret) < size {
			return "", errors.New("readString error: unexpected length")
		}
		for !decoder.lastChunk {
			chunk, err := decoder.ReadString()
			if err != nil {
				return "", err
			}
			ret += chunk
		}
		decoder.lastChunk = false
    decoder.success()
		return ret, nil
	case code >= 0x00 && code <= 0x1f:
		size := int(code)
		ret := decoder.read_n_rune(size)
		if len(ret) < size {
			return "", errors.New("readString error: unexpected length")
		}
    decoder.success()
		return ret, nil
	case code >= 0x30 && code <= 0x33:
		bits := decoder.readn(1)
		if len(bits) < 1 {
			return "", errors.New("readString error: unexpected length")
		}
		size := int((code-0x30)<<8 + bits[0])
		ret := decoder.read_n_rune(size)
		if len(ret) < size {
			return "", errors.New("readString error: unexpected length")
		}
    decoder.success()
		return ret, nil
	default:
		return "", errors.New("readString error: unexpected code")
	}
}

/**
 * binary ::= b b1 b0 <binary-data> binary
 *        ::= B(final_chunk) b1 b0 <binary-data>
 *        ::= [x20-x2f] <binary-data>
 */
func (decoder *Decoder) ReadBinary() ([]byte, error) {
	code, err := decoder.read()
	if err != nil {
		return nil, errors.New("readBinary error: unexpected code")
	}
	switch {
	case code == 0x62:
		bits := decoder.readn(2)
		if len(bits) < 2 {
			return nil, errors.New("readBinary error: unexpected length")
		}
		size := int(bits[0]<<8 + bits[1])
		ret := decoder.readn(size)
		if len(ret) < size {
			return nil, errors.New("readBinary error: unexpected length")
		}
		for !decoder.lastChunk {
			chunk, err := decoder.ReadBinary()
			if err != nil {
				return nil, err
			}
			ret = append(ret, chunk...)
		}
		// reset
		decoder.lastChunk = false
    decoder.success()
		return ret, nil
	case code == 0x42 /*B*/ :
		decoder.lastChunk = true
		bits := decoder.readn(2)
		if len(bits) < 2 {
			return nil, errors.New("readBinary error: unexpected length")
		}
		size := int(bits[0]<<8 + bits[1])
		bits = decoder.readn(size)
		if len(bits) < size {
			return nil, errors.New("readBinary error: unexpected length")
		}
    decoder.success()
		return bits, nil
	case code >= 0x20 && code <= 0x2f:
		size := int(code - 0x20)
		ret := decoder.readn(size)
		if len(ret) < size {
			return nil, errors.New("readBinary error: unexpected length")
		}
    decoder.success()
		return ret, nil
	default:
		return nil, errors.New("readBinary error: unexpected code")
	}
}

/**
 * long ::= L b7 b6 b5 b4 b3 b2 b1 b0
 *      ::= [xd8-xef]
 *      ::= [xf0-xff] b0
 *      ::= [x38-x3f] b1 b0
 *      ::= x4c b3 b2 b1 b0 // hessian 2 java use 0x59, hessian 3 use 0x77
 * here we use 0x59
 */
func (decoder Decoder) ReadLong() (int64, error) {
	code, err := decoder.read()
	if err != nil {
		return -1, err
	}
	switch {
	case code == 0x4c /*L*/ :
		bits := decoder.readn(8)
		if len(bits) < 8 {
			return -1, errors.New("readLong error: unexpected length")
		}
    decoder.success()
		return parseInt64FromBytes(bits), nil
	case code >= 0xd8 && code <= 0xef:
    return parseInt64(int8(code-0xe0), []byte{}), nil
	case code >= 0xf0 && code <= 0xff:
		bits := decoder.readn(1)
		if len(bits) < 1 {
			return -1, errors.New("readLong error: unexpected length")
		}
    decoder.success()
		return parseInt64(int8(code-0xf8), bits), nil
	case code >= 0x38 && code <= 0x3f:
		bits := decoder.readn(2)
		if len(bits) < 2 {
			return -1, errors.New("readLong error: unexpected length")
		}
    decoder.success()
		return parseInt64(int8(code-0x3c), bits), nil
	case code == 0x59:
		bits := decoder.readn(4)
		if len(bits) < 4 {
			return -1, errors.New("readLong error: unexpected length")
		}
    decoder.success()
		return parseInt64FromBytes(bits), nil
	}
	return -1, errors.New("readLong error: unexpected code")
}

/**
 *   double ::= D b7 b6 b5 b4 b3 b2 b1 b0
 *          ::= x5b
 *          ::= x5c
 *          ::= x5d b0
 *          ::= x5e b1 b0
 *          ::= x5f b3 b2 b1 b0
 */
func (decoder Decoder) ReadDouble() (float64, error) {
	code, err := decoder.read()
	if err != nil {
		return -1, err
	}
	switch {
	case code == 0x44:
		bits := decoder.readn(8)
		if len(bits) < 8 {
			return 0.0, errors.New("readDouble: unexpected length")
		}
    decoder.success()
		return parseFloat64FromBytes(bits), nil
	case code == 0x5b:
    decoder.success()
		return 0.0, nil
	case code == 0x5c:
    decoder.success()
		return 1.0, nil
	case code == 0x5d:
		bits := decoder.readn(1)
		if len(bits) < 1 {
			return 0.0, errors.New("readDouble: unexpected length")
		}
    decoder.success()
		return float64(int8(bits[0])), nil
	case code == 0x5e:
		bits := decoder.readn(2)
		if len(bits) < 2 {
			return 0.0, errors.New("readDouble: unexpected length")
		}
    decoder.success()
		return float64(parseInt16(int8(bits[0]), bits[1])), nil
	case code == 0x5f:
		bits := decoder.readn(4)
		if len(bits) < 4 {
			return 0.0, errors.New("readDouble: unexpected length")
		}
    decoder.success()
		return float64(parseFloat32FromBytes(bits)), nil
	default:
		return 0.0, errors.New("readDouble: unexpected length")
	}
}

func (decoder Decoder) ReadDate() (time.Time, error) {
	code, err := decoder.read()
  if err != nil {
    return TIME_DEFAULT_VALUE, err
  }
  switch code {
  case 0x4a:
    bits := decoder.readn(8)
    if len(bits) < 8 {
      return TIME_DEFAULT_VALUE, errors.New("readDate: unexpected length")
    }
    decoder.success()
    ms := parseInt64FromBytes(bits)
    return time.Unix(ms / 1000, (ms % 1000) * 1e6), nil
  case 0x4b:
    bits := decoder.readn(4)
    if len(bits) < 4 {
      return TIME_DEFAULT_VALUE, errors.New("readDate: unexpected length")
    }
    decoder.success()
    minutes := parseInt64FromBytes(bits)
    return time.Unix(minutes * 60, 0), nil
  default:
    return TIME_DEFAULT_VALUE, errors.New("readDate: unexpected code")
  }
}

func (decoder *Decoder) ReadNull() (interface{}, error) {
  code, err := decoder.read()
  if err != nil {
    return nil, err
  }
  if code == 0x4e {
    decoder.success()
    return nil, nil
  }
  return nil, errors.New("readNull: unexpected code")
}

/**
 * ref :: = x51 int
 */
func (decoder *Decoder) ReadRef() (interface{}, error){
  code, err := decoder.read()
  if err != nil {
    return nil, err
  }
  if code != 0x51 {
    return nil, errors.New("readRef: unexpected code")
  }
  refId, err := decoder.ReadInt()
  if err != nil {
    return nil, err
  }
  ret, ok := decoder.refMap[refId]
  if !ok {
    return nil, errors.New("readRef: unexpected ref")
  }
  decoder.success()
  return ret, nil
}

/**
 * type ::= string
 *      ::= int(type-ref)
 */
func (decoder *Decoder) ReadType() (string, error) {
  s, err := decoder.ReadString()
  if err != nil {
    decoder.Recover()
    refId, err := decoder.ReadInt()
    if err != nil {
      return "", errors.New("readType: unexpected code")
    }
    ref, ok := decoder.refMap[refId]
    if !ok {
      return "", errors.New("readType: unknown type")
    }
    stringType, ok := ref.(string) // assertion
    if !ok {
      return "", errors.New("readType: unknown type")
    }
    decoder.success()
    return stringType, nil
  }
  decoder.addRef(s)
  decoder.success()
  return s, nil
}

/**
list ::= x55 type value* 'Z'   # variable-length list
     ::= 'V' type int value*   # fixed-length list
     ::= x57 value* 'Z'        # variable-length untyped list
     ::= x58 int value*        # fixed-length untyped list
     ::= [x70-77] type value*  # fixed-length typed list
     ::= [x78-7f] value*       # fixed-length untyped list
*/
func (decoder *Decoder) ReadList() (List, error) {
  code, err := decoder.read()
  if err != nil {
    return List{}, err
  }
  switch {
  case code == 0x55:
    parsedType, err := decoder.ReadType()
    if err != nil {
      return List{}, err
    }
    listValue := []interface{}{}
    for {
      v, err := dynamic_call(decoder, parsedType)
      listValue := append(listValue, v)
      if err != nil {
        decoder.Recover()
        end, err := decoder.read()
        if err != nil || end != 0x5a /*Z*/ {
          return List{}, errors.New("readList: unexpected code")
        } else {
          return List{
            ValueType: parsedType,
            Value: listValue,
          }, nil
        }
      }
    }
  case code == 0x56:
    parsedType, err := decoder.ReadType()
    if err != nil {
      return List{}, err
    }
    size, err := decoder.ReadInt()
    if err != nil {
      return List{}, err
    }
    listValue, err := decoder.readFixedLengthTypedValue(int(size), parsedType)
    if err != nil {
      return List{}, err
    }
    return List{parsedType, listValue}, nil
  case code == 0x58:
    size, err := decoder.ReadInt()
    if err != nil {
      return List{}, err
    }
    listValue, err := decoder.readFixedLengthUnTypedValue(int(size))
    if err != nil {
      return List{}, err
    }
    return List{
      ValueType: UNTYPED,
      Value: listValue,
    }, nil
  case code >= 0x70 && code <= 0x77:
    size := int(code - 0x70)
    parsedType, err := decoder.ReadType()
    // parsedType startWith '[', eg: "[java.lang.Integer"
    if err != nil {
      fmt.Println("parse type error")
      return List{}, err
    }
    ret := List{
      ValueType: parsedType[1:],
      Value: []interface{}{},
    }
    listValue, err := decoder.readFixedLengthTypedValue(size, parsedType[1:])
    if err != nil {
      return List{}, err
    }
    ret.Value = listValue
    return ret, nil
  }
  return List{}, nil
}

func (decoder *Decoder) ReadMap() (map[interface{}]interface{}, error) {
  code, err := decoder.read()
  if err != nil {
    return nil, err
  }
  if code != 0x48 {
    return nil, errors.New("decoder ReadMap: unexpected error")
  }
  ret := map[interface{}]interface{}{}
  for {
    code, err = decoder.read()
    if err != nil {
      return nil, err
    }
    if code == 0x5a {
      break
    }
    keyType := CODE_TO_TYPE[code]
    decoder.unread_n_byte(1)
    key, err := dynamic_call(decoder, keyType)
    if err != nil {
      return nil, err
    }
    code, err = decoder.read()
    if err != nil {
      return nil, err
    }
    valueType := CODE_TO_TYPE[code]
    decoder.unread_n_byte(1)
    value, err := dynamic_call(decoder, valueType)
    if err != nil {
      return nil, err
    }
    ret[key] = value
  }
  return ret, nil
}

func dynamic_call(decoder *Decoder, typeName string) (interface{}, error) {
  switch typeName {
  case "java.lang.Integer":
    fallthrough
  case "int":
    return decoder.ReadInt()
  case "double":
    return decoder.ReadDouble()
  case "string":
    return decoder.ReadString()
  }
  return nil, errors.New("no such method")
}

