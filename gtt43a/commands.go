package gtt43a

import (
	"encoding/binary"
	"fmt"
)

type GTT25CommandType []byte

var Bitmap_Load GTT25CommandType = []byte{0x0D, 0x00}
var Bitmap_Capture GTT25CommandType = []byte{0x0D, 0x01}
var Begin_Update GTT25CommandType = []byte{0x1F, 0x00}
var End_Update GTT25CommandType = []byte{0x1F, 0x01}
var Create_Object GTT25CommandType = []byte{0x01, 0x00}
var Destroy_Object GTT25CommandType = []byte{0x01, 0x01}
var Set_Focus GTT25CommandType = []byte{0x02, 0x02}
var ObjectList_Get GTT25CommandType = []byte{0x1A, 0x03}
var SetBacklight GTT25CommandType = []byte{153}

func (typeP GTT25CommandType) Value() []byte {
	return []byte(typeP)
}

func (m *display) BaseObjectBeginUpdate(id int) error {

	data := []byte{0xFE, 0xFA}
	data = append(data, Begin_Update.Value()...)
	idb := make([]byte, 2)
	binary.BigEndian.PutUint16(idb, uint16(id))
	data = append(data, idb...)

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

func (m *display) BaseObjectEndUpdate(id int) error {

	data := []byte{0xFE, 0xFA}
	data = append(data, End_Update.Value()...)
	idb := make([]byte, 2)
	binary.BigEndian.PutUint16(idb, uint16(id))
	data = append(data, idb...)

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

func (m *display) CreateObject(id int, objectType GTT25ObjectType) error {

	data := []byte{0xFE, 0xFA}
	data = append(data, Create_Object.Value()...)
	data = append(data, objectType.Value()...)
	idb := make([]byte, 2)
	binary.BigEndian.PutUint16(idb, uint16(id))
	data = append(data, idb...)

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
	return nil
}

func (m *display) DestroyObject(id int) error {

	data := []byte{0xFE, 0xFA}
	data = append(data, Destroy_Object.Value()...)
	idb := make([]byte, 2)
	binary.BigEndian.PutUint16(idb, uint16(id))
	data = append(data, idb...)

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

func (m *display) SetFocus(id int) error {

	data := []byte{0xFE, 0xFA}
	data = append(data, Set_Focus.Value()...)
	idb := make([]byte, 2)
	binary.BigEndian.PutUint16(idb, uint16(id))
	data = append(data, idb...)

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

func (m *display) ObjectListGet(id, itemIndex int) error {

	data := []byte{0xFE, 0xFA}
	data = append(data, ObjectList_Get.Value()...)
	idb := make([]byte, 2)
	binary.BigEndian.PutUint16(idb, uint16(id))
	data = append(data, idb...)
	indexb := make([]byte, 4)
	binary.BigEndian.PutUint32(indexb, uint32(itemIndex))
	data = append(data, indexb...)

	/**/
	res, err := m.SendRecv(data)
	if err != nil {
		return err
	}
	if len(res) < 3 {
		return fmt.Errorf("error in response: [% X]", res)
	}
	if res[len(res)-1] != byte(0xFE) {
		return fmt.Errorf("error in request SetFocus, status code: [%X]", res[2])
	}
	/**/
	return nil
}

func (m *display) SetBacklightLegcay(brightness int) error {
	data := make([]byte, 0)
	data = append(data, byte(brightness))
	_, err := m.SendRecvCmd(153, data)
	return err
}
