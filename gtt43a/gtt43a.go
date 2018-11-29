/**
Package to send commands and recieve response to and from gtt43a device.
**/
package gtt43a


import (
	"github.com/tarm/serial"
	"sync"
	"time"
	_ "os"
	"fmt"
	"encoding/binary"
)


type PortOptions struct {
        Port    string
        Baud    int
}

type Display interface {
	Open()			bool
	Close()			bool
	ClrScreen()		int
	Text(string)		int
	FontSize(int)		int
	Send([]byte)		int
	SendRecv([]byte, bool)	(int, []byte)
	Echo([]byte)		(int, []byte)
	SendRecvCmd(int, []byte, bool)	(int, []byte)
	SendCmd(int, []byte)	int
	Reset()			int
	TextInsertPoint(int, int)	(int)
	GetTextPoint()			(int, []byte)
	TextPoint(int, int) func(data string) int
	TextWindow(int, int, int, int)	int
	TextColour(int, int, int)	int
	PrintUTF8String(string)		int
	PrintUnicode([]byte)	int
	UpdateLabel(id, format int, value []byte) int
	UpdateLabelAscii(int,string)	int
	UpdateLabelUTF8(int, string)	int
	UpdateLabelUnicode(int, []byte) int
	UpdateBargraphValue(int, int)	(int, []byte)
	UpdateTraceValue(int, int)	int
	RunScript(string)		int
	LoadBitmap(int, string)		(int, []byte)
	BuzzerActive(frec, time int) (int)
	WriteScratch(addr int, data []byte) (int)
	ReadScratch(addr, size int) (int, []byte)

}

type display struct {
        options *PortOptions
        status  uint32
        port    *serial.Port
        mux     sync.Mutex
}

const (
        OPENED uint32 = iota
        CLOSED
)

const (
	OFF int = iota
	GREEN
	RED
	YELLOW
)

//Create a new Display device
func NewDisplay(opt *PortOptions) Display {
	disp := &display{}
	disp.options = opt
	disp.status = CLOSED

	return disp
}

//Open device comunication channel
func (m *display) Open() (bool) {
        if m.status == OPENED {
                return true
        }

        config := &serial.Config{
                Name:   m.options.Port,
                Baud:   m.options.Baud,
                ReadTimeout:    60 * time.Millisecond,
        }

        var err error
        m.port, err = serial.OpenPort(config)
        if err != nil {
                //log.Println(err)
                return false
        }

        m.status = OPENED
        return true
}

//Clsoe device comunication channel
func (m *display) Close() bool {
        if m.status == CLOSED {
                return true
        }

        if err := m.port.Close(); err != nil {
                return false
        }
        m.status = CLOSED
        return true
}

//Primitive function to send and recieve bytes to and from display device.
//recv, flag to wait a response form device.
func (m *display) SendRecv(data []byte, recv bool) (int, []byte) {
	m.mux.Lock()
	res := make([]byte,0)
	n := -1
	defer m.mux.Unlock()

	if m.status == CLOSED {
		return -1, nil
	}
	if data != nil {
		n, err := m.port.Write(data)
		if err != nil {
			return -1, nil
		}
		if n <= 0 {
			return n, nil
		}
		if !recv {
			return n, nil
		}
	}

	var err error
	for {
		buf := make([]byte,128)
		n, err = m.port.Read(buf)
		if err != nil && n <=0 {
			//fmt.Println(err)
			break
		}
		res = append(res,buf[:n]...)
	}
	return len(res), res
}

//Send bytes data to device. Don't wait response.
func (m *display) Send(data []byte) (int) {
	n, _ := m.SendRecv(data, false)
	return n
}

//Send a command to display device
//cmd, id for the command
//recv, flag to wait response
func (m *display) SendRecvCmd(cmd int, data []byte, recv bool) (int, []byte) {
	dat1 := []byte{0xFE, byte(cmd)}
	if data != nil {
		dat1 = append(dat1, data...)
	}

	n, res := m.SendRecv(dat1, recv)

	switch  {
	case n <= 0:
		return -1, nil
	case recv && res == nil:
		return -2, nil
	case recv && len(res) < 4 :
		return -3, nil
	case recv && res[0] != 0xFE:
		return -4, nil
	case recv && res[1] != byte(cmd):
		return -5, nil
	}
	return n, res
}

//Send a Command to display device.
func (m *display) SendCmd(cmd int, data []byte) int {
	n, _ := m.SendRecvCmd(cmd, data, false)
	return n
}

//Send echo data and to wait for a response.
func (m *display) Echo(data []byte) (int, []byte) {
	return m.SendRecvCmd(0xFF, data, true)
}

//Send reset command to display device
func (m *display) Reset() (int) {
	n := m.Send([]byte{0xFE, 0x01})
	return n
}

//Print text data in actual (x,y) point in display area
func (m *display) Text(data string) (int) {
	n := m.Send([]byte(data))
	return n
}

//Set font Size
func (m *display) FontSize(size int) (int) {
	return m.SendCmd(0x33, []byte{byte(size)})
}

//Set (x,y) point in display area. The next print and draw command will be set in this point.
func (m *display) TextInsertPoint(x, y int) (int) {
	data := make([]byte,0)
	xb := make([]byte,2)
	yb := make([]byte,2)
	binary.BigEndian.PutUint16(xb, uint16(x))
	binary.BigEndian.PutUint16(yb, uint16(x))
	data = append(data, xb...)
	data = append(data, yb...)
	n := m.SendCmd(0x79, data)
	return n
}

//Get actual (x,y) point
func (m *display) GetTextPoint() (int, []byte) {
	return m.SendRecvCmd(0x7A, nil, true)
}

//Clear actual Screen
func (m *display) ClrScreen() (int) {
	return m.SendCmd(0x58, nil)
}

//Print data text in this (x,y) point
func (m *display) TextPoint(x, y int) func(data string) int {
	return func (data string) (int) {
		n := m.TextInsertPoint(x, y)
		if n <= 0 {
			return n
		}
		n = m.Send([]byte(data))
		return n
	}
}

//Set (x,y) point for the all future text in the actual windowText
func (m *display) TextWindow(x, y, width, height int) int {
	data := make([]byte,0)
	xb := make([]byte,2)
	yb := make([]byte,2)
	widthb := make([]byte,2)
	heightb := make([]byte,2)
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
func (m *display) TextColour(r, g, b int) int {
	data := []byte{byte(r), byte(g), byte(b)}

	return m.SendCmd(0x2E, data)
}

//Print the text data in UTF-8 codification
func (m *display) PrintUTF8String(text string) int {

	return m.SendCmd(0x25, []byte(text))
}

//Print the data bytes in Unicode (16 bits length) codification
func (m *display) PrintUnicode(data []byte) int {

	return m.SendCmd(0x24, data)
}

//Update the text data (in bytes) in the label ID
func (m *display) UpdateLabel(id, format int, value []byte) int  {
	data := []byte{byte(id), byte(format)}
	data = append(data, []byte(value)...)
	data = append(data, 0x00)

	return m.SendCmd(0x11, data)
}

//Update the text data (in string) in the label ID with Ascii Codification
func (m *display) UpdateLabelAscii(id int, value string) int  {
	return m.UpdateLabel(id, 0, []byte(value))
}

//Update the text data (in string) in the label ID with UTF-8 Codification
func (m *display) UpdateLabelUTF8(id int, value string) int  {
	return m.UpdateLabel(id, 2, []byte(value))
}

//Update the text data (in bytes, 2 bytes for character) in the label ID with Unicode Codification
func (m *display) UpdateLabelUnicode(id int, value []byte) int  {
	return m.UpdateLabel(id, 1, value)
}

//Update value (%0 - %100) in bargraph object
func (m *display) UpdateBargraphValue(id, value int) (int, []byte)  {
	data := []byte{byte(id)}
	valueb := make([]byte, 2)
	binary.BigEndian.PutUint16(valueb, uint16(value))

	data = append(data, valueb...)

	return m.SendRecvCmd(0x69, data, true)
}

//Update value in trace object
func (m *display) UpdateTraceValue(id, value int) int  {
	data := []byte{byte(id)}
	valueb := make([]byte, 2)
	binary.BigEndian.PutUint16(valueb, uint16(value))

	data = append(data, valueb...)

	return m.SendCmd(0x75, data)
}

//Run script binary. The filename path is a local path in display device
func (m *display) RunScript(filename string) int {
	data := []byte(filename)
	data = append(data, 0x00)
	n1, _ := m.SendRecvCmd(0x5D, data, false)
	//fmt.Printf("salida Run: %v\n", res)
	return n1
}

//Load in display memory a bitmap object from filename. The filename path in a local in display device.
func (m *display) LoadBitmap(id int, filename string) (int, []byte) {
	data := []byte(filename)
	data = append(data, 0)
	n1, res := m.SendRecvCmd(0x5F, data, true)
	fmt.Printf("salida Run: %v\n", res)
	return n1, res
}

//Active buzzer in device.
//frec, is the frecuency of the signal
//time, is the duration of the signal
func (m *display) BuzzerActive(frec, time int) (int) {

	data := []byte{0xFE, 0xBB}
	frecb := make([]byte, 2)
	timeb := make([]byte, 2)
	binary.BigEndian.PutUint16(frecb, uint16(frec))
	binary.BigEndian.PutUint16(timeb, uint16(time))
	data = append(data, frecb...)
	data = append(data, timeb...)
	n3 := m.Send(data)
	return n3
}

//TOUCH

//Create a touch region
//regId, region ID
//x, y, coordinate of the touch region
//width, width of the region
//height, height of the region

func (m *display) WriteScratch(addr int, data []byte) (int) {
	dat1 := []byte{0xFE, 0xCC}
	addrb := make([]byte,2)
	sizeb := make([]byte,2)
	binary.BigEndian.PutUint16(addrb, uint16(addr))
	binary.BigEndian.PutUint16(sizeb, uint16(len(data)))
	dat1 = append(dat1, addrb...)
	dat1 = append(dat1, sizeb...)
	dat1 = append(dat1, data...)
	n := m.Send(dat1)

	return n
}

func (m *display) ReadScratch(addr, size int) (int, []byte) {
	dat1 := []byte{0xFE, 0xCD}
	addrb := make([]byte,2)
	sizeb := make([]byte,2)
	binary.BigEndian.PutUint16(addrb, uint16(addr))
	binary.BigEndian.PutUint16(sizeb, uint16(size))
	dat1 = append(dat1, addrb...)
	dat1 = append(dat1, sizeb...)
	n, datOut := m.SendRecv(dat1, true)

	return n, datOut
}

