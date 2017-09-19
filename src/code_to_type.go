package hessian

var CODE_TO_TYPE = map[byte]string{}

func addCode(code byte, typeName string) {
  CODE_TO_TYPE[code] = typeName
}

func addCodeRange(start byte, end byte, typeName string) {
  for i := start; i <= end; i++ {
    CODE_TO_TYPE[i] = typeName
  }
}

func init() {
  addCode(0x91, "int")
  addCode(0x49, "int")
  addCodeRange(0x80, 0xbf, "int")
  addCodeRange(0xc0, 0xcf, "int")
  addCodeRange(0xd0, 0xd7, "int")
  addCode(0x54, "bool")
  addCode(0x46, "bool")
  addCode(0x53, "string")
  addCode(0x52, "string")
  addCodeRange(0x00, 0x1f, "string")
  addCodeRange(0x30, 0x33, "string")
  addCode(0x4d, "typedmap")
}
