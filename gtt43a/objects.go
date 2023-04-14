package gtt43a

type GTT25ObjectType []byte

func (typeP GTT25ObjectType) Value() []byte {
	return []byte(typeP)
}

var ObjectType_Bitmap GTT25ObjectType = []byte{0x00, 0x0d}
var ObjectType_VisualBitmap GTT25ObjectType = []byte{0x00, 0x1f}

const textEncoding_ASCII = 1
const textEncoding_UTF8 = 2
