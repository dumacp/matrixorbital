package gtt43a

import "encoding/binary"

//Update the text data (in bytes) in the label ID
func (m *display) UpdateLabel(id, format int, value []byte) error {
	data := []byte{byte(id), byte(format)}
	data = append(data, []byte(value)...)
	data = append(data, 0x00)

	return m.SendCmd(0x11, data)
}

//Update the text data (in string) in the label ID with Ascii Codification
func (m *display) UpdateLabelAscii(id int, value string) error {
	return m.UpdateLabel(id, 0, []byte(value))
}

//Update the text data (in string) in the label ID with UTF-8 Codification
func (m *display) UpdateLabelUTF8(id int, value string) error {
	return m.UpdateLabel(id, 2, []byte(value))
}

//Update the text data (in bytes, 2 bytes for character) in the label ID with Unicode Codification
func (m *display) UpdateLabelUnicode(id int, value []byte) error {
	return m.UpdateLabel(id, 1, value)
}

//Update value (%0 - %100) in bargraph object
func (m *display) UpdateBargraphValue(id, value int) ([]byte, error) {
	data := []byte{byte(id)}
	valueb := make([]byte, 2)
	binary.BigEndian.PutUint16(valueb, uint16(value))

	data = append(data, valueb...)

	return m.SendRecvCmd(0x69, data)
}

//Update value in trace object
func (m *display) UpdateTraceValue(id, value int) error {
	data := []byte{byte(id)}
	valueb := make([]byte, 2)
	binary.BigEndian.PutUint16(valueb, uint16(value))

	data = append(data, valueb...)

	return m.SendCmd(0x75, data)
}
