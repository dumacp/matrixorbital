package gtt43a

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

//Load in display memory a bitmap object from filename. The filename path in a local in display device.
func (m *display) LoadBitmapLegcay(id int, filename string) error {
	data := make([]byte, 0)
	data = append(data, byte(id))
	data = append(data, []byte(filename)...)
	data = append(data, 0)
	_, err := m.SendRecvCmd(0x5F, data)
	return err
}

func (m *display) DisplayBitmapLegcay(id int, x, y int) error {
	data := make([]byte, 0)
	data = append(data, byte(id))
	xb := make([]byte, 2)
	binary.BigEndian.PutUint16(xb, uint16(x))
	yb := make([]byte, 2)
	binary.BigEndian.PutUint16(yb, uint16(y))
	data = append(data, xb...)
	data = append(data, yb...)
	_, err := m.SendRecvCmd(0x61, data)
	return err
}

//Load in display memory a bitmap object from filename. The filename path in a local in display device.
func (m *display) ClearBitmapLegacy(id int) error {
	data := make([]byte, 0)
	data = append(data, 0x01)
	data = append(data, byte(id))
	_, err := m.SendRecvCmd(0xD0, data)
	return err
}

func (m *display) SetBitmapTransparencyLegacy(id int, r, g, b int) error {
	data := make([]byte, 0)
	data = append(data, byte(id))
	data = append(data, byte(r))
	data = append(data, byte(g))
	data = append(data, byte(b))
	_, err := m.SendRecvCmd(0x62, data)
	return err
}

//Load in display memory a bitmap object from filename. The filename path in a local in display device.
func (m *display) BitmapLoad(id int, filename string) error {
	prefix := []byte{0xFE, 0xFA}
	prefix = append(prefix, Bitmap_Load.Value()...)
	idb := make([]byte, 2)
	binary.BigEndian.PutUint16(idb, uint16(id))
	prefix = append(prefix, idb...)

	value16 := utf16.Encode([]rune(filename))
	value := make([]byte, 0)
	for _, v := range value16 {
		tempB := make([]byte, 2)
		binary.LittleEndian.PutUint16(tempB, uint16(v))
		value = append(value, tempB...)
	}
	value = append(value, 0x00)

	filepacket := make([]byte, 0)
	filepacket = append(filepacket, byte(0))
	lenb := make([]byte, 2)
	binary.BigEndian.PutUint16(lenb, uint16(len(value)))
	filepacket = append(filepacket, lenb...)
	filepacket = append(filepacket, []byte(value)...)

	data := make([]byte, 0)
	data = append(data, prefix...)
	data = append(data, filepacket...)

	res, err := m.SendRecv(data)
	if err != nil {
		return err
	}
	if len(res) < 3 {
		return fmt.Errorf("error in response: [% X]", res)
	}
	if res[len(res)-1] != byte(0xFE) {
		return fmt.Errorf("error in request U8, status code: [%X], [% X]", res[2], res)
	}
	return nil
}

func (m *display) BitmapCapture(id int, left, top, width, height int) error {
	prefix := []byte{0xFE, 0xFA}
	prefix = append(prefix, Bitmap_Capture.Value()...)
	idb := make([]byte, 2)
	binary.BigEndian.PutUint16(idb, uint16(id))
	prefix = append(prefix, idb...)

	leftb := make([]byte, 2)
	binary.BigEndian.PutUint16(leftb, uint16(left))
	topb := make([]byte, 2)
	binary.BigEndian.PutUint16(topb, uint16(top))
	widthb := make([]byte, 2)
	binary.BigEndian.PutUint16(widthb, uint16(width))
	heightb := make([]byte, 2)
	binary.BigEndian.PutUint16(heightb, uint16(height))

	data := make([]byte, 0)
	data = append(data, prefix...)
	data = append(data, leftb...)
	data = append(data, topb...)
	data = append(data, widthb...)
	data = append(data, heightb...)

	res, err := m.SendRecv(data)
	if err != nil {
		return err
	}
	if len(res) < 3 {
		return fmt.Errorf("error in response: [% X]", res)
	}
	if res[len(res)-1] != byte(0xFE) {
		return fmt.Errorf("error in request U8, status code: [%X], [% X]", res[2], res)
	}
	return nil
}
