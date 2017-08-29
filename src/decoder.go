package hessian

import (
	"bytes"
	"github.com/pkg/errors"
)

type Decoder struct {
	buf       *bytes.Buffer
	lastChunk bool
}

func NewDecoder(b []byte) *Decoder {
	return &Decoder{bytes.NewBuffer(b), false}
}

func (decoder *Decoder) read() (byte, error) {
	return decoder.buf.ReadByte()
}

func (decoder *Decoder) readn(n int) []byte {
	return decoder.buf.Next(n)
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
	return string(ret)
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
		return parseInt32(int8(bits[0]), bits[1:]), nil
	}
	if code >= 0x80 && code <= 0xbf {
		return int32(code - 0x80), nil
	}
	if code >= 0xc0 && code <= 0xcf {
		b0, err := decoder.read()
		if err != nil {
			return 0, err
		}
		return parseInt32(int8(code-0xc8), []byte{b0}), nil
	}
	if code >= 0xd0 && code <= 0xd7 {
		bits := decoder.readn(2)
		if len(bits) < 2 {
			return 0, errors.New("readInt error: unexpected length of bytes")
		}
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
		size := int((code-0x30)<<8 + bits[0])
		ret := decoder.read_n_rune(size)
		if len(ret) < size {
			return "", errors.New("readString error: unexpected length")
		}
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
		return bits, nil
	case code >= 0x20 && code <= 0x2f:
		size := int(code - 0x20)
		ret := decoder.readn(size)
		if len(ret) < size {
			return nil, errors.New("readBinary error: unexpected length")
		}
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
		return parseInt64(int8(bits[0]), bits[1:]), nil
	case code >= 0xd8 && code <= 0xef:
		return parseInt64(int8(code-0xe0), []byte{}), nil
	case code >= 0xf0 && code <= 0xff:
		bits := decoder.readn(1)
		if len(bits) < 1 {
			return -1, errors.New("readLong error: unexpected length")
		}
		return parseInt64(int8(code-0xf8), bits), nil
	case code >= 0x38 && code <= 0x3f:
		bits := decoder.readn(2)
		if len(bits) < 2 {
			return -1, errors.New("readLong error: unexpected length")
		}
		return parseInt64(int8(code-0x3c), bits), nil
	case code == 0x59:
		bits := decoder.readn(4)
		if len(bits) < 4 {
			return -1, errors.New("readLong error: unexpected length")
		}
		return parseInt64(int8(bits[0]), bits[1:]), nil
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
		return parseFloat64FromBytes(bits), nil
	case code == 0x5b:
		return 0.0, nil
	case code == 0x5c:
		return 1.0, nil
	case code == 0x5d:
		bits := decoder.readn(1)
		if len(bits) < 1 {
			return 0.0, errors.New("readDouble: unexpected length")
		}
		return float64(int8(bits[0])), nil
	case code == 0x5e:
		bits := decoder.readn(2)
		if len(bits) < 2 {
			return 0.0, errors.New("readDouble: unexpected length")
		}
		return float64(parseInt16(int8(bits[0]), bits[1])), nil
	case code == 0x5f:
		bits := decoder.readn(4)
		if len(bits) < 4 {
			return 0.0, errors.New("readDouble: unexpected length")
		}
		return float64(parseFloat32FromBytes(bits)), nil
	default:
		return 0.0, errors.New("readDouble: unexpected length")
	}
}
