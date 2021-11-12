package gtt43a

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

/**/
type GTT25PropertyType []byte

var GaugeValue GTT25PropertyType = []byte{0x03, 0x02}
var LabelText GTT25PropertyType = []byte{0x09, 0x06}
var LabelFontSize GTT25PropertyType = []byte{0x09, 0x0A}
var SliderValue GTT25PropertyType = []byte{0x0A, 0x08}
var ButtonState GTT25PropertyType = []byte{0x15, 0x0C}
var ButtonText GTT25PropertyType = []byte{0x15, 0x03}
var SliderLabelText GTT25PropertyType = []byte{0x0A, 0x09}

func (typeP GTT25PropertyType) Value() []byte {
	return []byte(typeP)
}

func ApduSetPropertyValueU16(id int, prpType GTT25PropertyType, value int) []byte {
	data := []byte{0xFE, 0xFA, 0x01, 0x06}
	idb := make([]byte, 2)
	binary.BigEndian.PutUint16(idb, uint16(id))
	valueb := make([]byte, 2)
	binary.BigEndian.PutUint16(valueb, uint16(value))
	data = append(data, idb...)
	data = append(data, prpType.Value()...)
	data = append(data, valueb...)
	return data
}

//Set Property ValueU16 GTT25Object
func (m *display) SetPropertyValueU16(id int, prpType GTT25PropertyType) func(value int) error {
	return func(value int) error {
		data := ApduSetPropertyValueU16(id, prpType, value)
		/**/
		res, err := m.SendRecv(data)
		if err != nil {
			return err
		}
		if len(res) < 3 {
			return fmt.Errorf("error in response: [% X]", res)
		}
		if res[len(res)-1] != byte(0xFE) {
			return fmt.Errorf("error in request U16, status code: [%X]", res[2])
		}
		/**/
		return nil
	}
}

func ApduSetPropertyValueS16(id int, prpType GTT25PropertyType, value int) []byte {
	data := []byte{0xFE, 0xFA, 0x01, 0x08}
	idb := make([]byte, 2)
	binary.BigEndian.PutUint16(idb, uint16(id))
	valueb := make([]byte, 2)
	binary.BigEndian.PutUint16(valueb, uint16(value))
	data = append(data, idb...)
	data = append(data, prpType.Value()...)
	data = append(data, valueb...)
	return data
}

//Set Property ValueS16 GTT25Object
func (m *display) SetPropertyValueS16(id int, prpType GTT25PropertyType) func(value int) error {
	return func(value int) error {
		data := ApduSetPropertyValueS16(id, prpType, value)
		/**/
		res, err := m.SendRecv(data)
		if err != nil {
			return err
		}
		if len(res) < 3 {
			return fmt.Errorf("error in response: [% X]", res)
		}
		if res[len(res)-1] != byte(0xFE) {
			return fmt.Errorf("error in request S16, status code: [%X]", res[2])
		}
		/**/
		return nil
	}
}

func ApduSetPropertyValueU8(id int, prpType GTT25PropertyType, value int) []byte {
	data := []byte{0xFE, 0xFA, 0x01, 0x04}
	idb := make([]byte, 2)
	binary.BigEndian.PutUint16(idb, uint16(id))
	data = append(data, idb...)
	data = append(data, prpType.Value()...)
	data = append(data, byte(value))
	return data
}

//Set Property ValueU8 GTT25Object
func (m *display) SetPropertyValueU8(id int, prpType GTT25PropertyType) func(value int) error {
	return func(value int) error {
		data := ApduSetPropertyValueU8(id, prpType, value)
		/**/
		res, err := m.SendRecv(data)
		if err != nil {
			return err
		}
		if len(res) < 3 {
			return fmt.Errorf("error in response: [% X]", res)
		}
		if res[len(res)-1] != byte(0xFE) {
			return fmt.Errorf("error in request U8, status code: [%X]", res[2])
		}
		/**/
		return nil
	}
}

func ApduSetPropertyText(id int, prpType GTT25PropertyType, text string) []byte {
	data := []byte{0xFE, 0xFA, 0x01, 0x0A}
	idb := make([]byte, 2)
	binary.BigEndian.PutUint16(idb, uint16(id))
	data = append(data, idb...)
	data = append(data, prpType.Value()...)
	data = append(data, 0x00)
	value16 := utf16.Encode([]rune(text))
	value := make([]byte, 0)
	for _, v := range value16 {
		tempB := make([]byte, 2)
		binary.LittleEndian.PutUint16(tempB, uint16(v))
		value = append(value, tempB...)
	}
	lenb := make([]byte, 2)
	binary.BigEndian.PutUint16(lenb, uint16(len(value)))
	data = append(data, lenb...)
	data = append(data, value...)
	return data
}

//Set Property Text GTT25Object
func (m *display) SetPropertyText(id int, prpType GTT25PropertyType) func(text string) error {
	return func(text string) error {
		data := ApduSetPropertyText(id, prpType, text)
		/**/
		res, err := m.SendRecv(data)
		if err != nil {
			return err
		}
		if len(res) < 3 {
			return fmt.Errorf("error in response: [% X]", res)
		}
		if res[len(res)-1] != byte(0xFE) {
			return fmt.Errorf("error in request, status code: [%X]", res[2])
		}
		/**/
		return nil
	}
}

func ApduGetPropertyValueU16(id int, prpType GTT25PropertyType) []byte {
	data := []byte{0xFE, 0xFA, 0x01, 0x07}
	idb := make([]byte, 2)
	binary.BigEndian.PutUint16(idb, uint16(id))
	data = append(data, idb...)
	data = append(data, prpType.Value()...)
	return data
}

//Get Property ValueU16 GTT25Object
func (m *display) GetPropertyValueU16(id int, prpType GTT25PropertyType) func() ([]byte, error) {
	return func() ([]byte, error) {
		data := ApduGetPropertyValueU16(id, prpType)
		res, err := m.SendRecv(data)
		if err != nil {
			return nil, err
		}
		if len(res) < 3 {
			return nil, fmt.Errorf("error in response: [% X]", res)
		}
		if res[len(res)-3] != byte(0xFE) {
			return nil, fmt.Errorf("error in request U16, status code: [%X]", res[2])
		}

		return res[len(res)-2:], nil
	}
}

func ApduGetPropertyValueS16(id int, prpType GTT25PropertyType) []byte {
	data := []byte{0xFE, 0xFA, 0x01, 0x09}
	idb := make([]byte, 2)
	binary.BigEndian.PutUint16(idb, uint16(id))
	data = append(data, idb...)
	data = append(data, prpType.Value()...)
	return data
}

//Get Property ValueS16 GTT25Object
func (m *display) GetPropertyValueS16(id int, prpType GTT25PropertyType) func() ([]byte, error) {
	return func() ([]byte, error) {
		data := ApduGetPropertyValueS16(id, prpType)
		res, err := m.SendRecv(data)
		if err != nil {
			return nil, err
		}
		if len(res) < 3 {
			return nil, fmt.Errorf("error in response: [% X]", res)
		}
		if res[len(res)-3] != byte(0xFE) {
			return nil, fmt.Errorf("error in request S16, status code: [%X]", res[2])
		}

		return res[len(res)-2:], nil
	}
}

func ApduGetPropertyValueU8(id int, prpType GTT25PropertyType) []byte {
	data := []byte{0xFE, 0xFA, 0x01, 0x05}
	idb := make([]byte, 2)
	binary.BigEndian.PutUint16(idb, uint16(id))
	data = append(data, idb...)
	data = append(data, prpType.Value()...)
	return data
}

//Get Property ValueU8 GTT25Object
func (m *display) GetPropertyValueU8(id int, prpType GTT25PropertyType) func() (byte, error) {
	return func() (byte, error) {
		data := ApduGetPropertyValueU8(id, prpType)
		res, err := m.SendRecv(data)
		if err != nil {
			return 0, err
		}
		if len(res) < 3 {
			return 0x00, fmt.Errorf("error in response: [% X]", res)
		}
		if res[len(res)-2] != byte(0xFE) {
			return byte(0x00), fmt.Errorf("error in request U8, status code: [%X]", res[2])
		}
		return res[len(res)-1], nil
	}
}
