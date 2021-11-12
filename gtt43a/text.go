package gtt43a

import "encoding/binary"

//Print text data in actual (x,y) point in display area
func (m *display) Text(data string) error {
	n := m.Send([]byte(data))
	return n
}

//Set font Size
func (m *display) FontSize(size int) error {
	return m.SendCmd(0x33, []byte{byte(size)})
}

//Set (x,y) point in display area. The next print and draw command will be set in this point.
func (m *display) TextInsertPoint(x, y int) error {
	data := make([]byte, 0)
	xb := make([]byte, 2)
	yb := make([]byte, 2)
	binary.BigEndian.PutUint16(xb, uint16(x))
	binary.BigEndian.PutUint16(yb, uint16(x))
	data = append(data, xb...)
	data = append(data, yb...)
	n := m.SendCmd(0x79, data)
	return n
}

//Get actual (x,y) point
func (m *display) GetTextPoint() ([]byte, error) {
	return m.SendRecvCmd(0x7A, nil)
}

//Print data text in this (x,y) point
func (m *display) TextPoint(x, y int) func(data string) error {
	return func(data string) error {
		if err := m.TextInsertPoint(x, y); err != nil {
			return err
		}
		return m.Send([]byte(data))
	}
}

//Set (x,y) point for the all future text in the actual windowText
func (m *display) TextWindow(x, y, width, height int) error {
	data := make([]byte, 0)
	xb := make([]byte, 2)
	yb := make([]byte, 2)
	widthb := make([]byte, 2)
	heightb := make([]byte, 2)
	binary.BigEndian.PutUint16(xb, uint16(x))
	binary.BigEndian.PutUint16(yb, uint16(x))
	binary.BigEndian.PutUint16(widthb, uint16(width))
	binary.BigEndian.PutUint16(heightb, uint16(height))
	data = append(data, xb...)
	data = append(data, yb...)
	data = append(data, widthb...)
	data = append(data, heightb...)

	return m.SendCmd(0x2B, data)
}

//Set the colour for the all future text and label string.
func (m *display) TextColour(r, g, b int) error {
	data := []byte{byte(r), byte(g), byte(b)}

	return m.SendCmd(0x2E, data)
}

//Print the text data in UTF-8 codification
func (m *display) PrintUTF8String(text string) error {

	return m.SendCmd(0x25, []byte(text))
}

//Print the data bytes in Unicode (16 bits length) codification
func (m *display) PrintUnicode(data []byte) error {

	return m.SendCmd(0x24, data)
}
