package hessian
import (
  "encoding/binary"
  "math"
  "bytes"
)

// 不足4位需要补足
func parseInt32FromBytes(bits []byte) (ret int32) {
  l := len(bits)
  if l < 4 {
    complements := []byte{}
    for i := 0; i < 4 - l; i++ {
      complements = append(complements, 0)
    }
    bits = append(complements, bits...)
  }
  buf := bytes.NewBuffer(bits)
  binary.Read(buf, binary.BigEndian, &ret)
  return
}

// 不足8位需补齐
func parseInt64FromBytes(bits []byte) (ret int64) {
  l := len(bits)
  if l < 8 {
    complements := []byte{}
    for i := 0; i < 8 - l; i++ {
      complements = append(complements, 0)
    }
    bits = append(complements, bits...)
  }
  buf := bytes.NewBuffer(bits)
  binary.Read(buf, binary.BigEndian, &ret)
  return
}

func parseFloat64FromBytes(bits []byte) float64 {
  t := binary.BigEndian.Uint64(bits)
  ret := math.Float64frombits(t)
  return ret
}

func parseFloat32FromBytes(bits []byte) float32 {
  t := binary.BigEndian.Uint32(bits)
  ret := math.Float32frombits(t)
  return ret
}

func parseInt32(sign int8, bits []byte) int32 {
	l := len(bits)
	var ret int32
	ret += int32(sign) << uint(l*8)
	for i := 0; i < l; i++ {
		ret += int32(bits[i] << uint((l-i-1)*8))
	}
	return ret
}

func parseInt16(sign int8, last byte) int16 {
  var ret int16
  ret += int16(int16(sign) << 8 + int16(last))
  return ret
}

func parseInt64(sign int8, bits []byte) int64 {
	l := len(bits)
	var ret int64
	ret += int64(sign) << uint(l * 8)
	for i := 0; i < l; i++ {
		ret += int64(bits[i]) << uint((l-i-1)*8)
	}
	return ret
}

