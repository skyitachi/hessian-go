package hessian

func parseInt32FromBytes(bits []byte) int32 {
	l := len(bits)
	var ret int32
	for i := 0; i < l; i++ {
		ret += int32(bits[i] << uint(i*8))
	}
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

func parseInt64(sign int8, bits []byte) int64 {
	l := len(bits)
	var ret int64
	ret += int64(sign) << uint(l*8)
	for i := 0; i < l; i++ {
		ret += int64(bits[i] << uint((l-i-1)*8))
	}
	return ret
}
