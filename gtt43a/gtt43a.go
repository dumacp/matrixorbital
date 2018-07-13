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
	PrintUnicodeString(string)	int
	UpdateLabel(id, format int, value string) int
	UpdateLabelAscii(int,string)	int
	UpdateLabelUTF8(int, string)	int
	UpdateLabelUnicode(int, string) int
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

func NewDisplay(opt *PortOptions) Display {
	disp := &display{}
	disp.options = opt
	disp.status = CLOSED

	return disp
}

func (m *display) Open() (bool) {
        if m.status == OPENED {
                return true
        }

        config := &serial.Config{
                Name:   m.options.Port,
                Baud:   m.options.Baud,
                ReadTimeout:    1 * time.Second,
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

func (m *display) SendRecv(data []byte, recv bool) (int, []byte) {
	m.mux.Lock()
	res := make([]byte,0)
	defer m.mux.Unlock()

	if m.status == CLOSED {
		return -1, nil
	}
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

	buf := make([]byte,128)
	n, _ = m.port.Read(buf)

	res = append(res,buf[:n]...)


	return n, res
}

func (m *display) Send(data []byte) (int) {
	n, _ := m.SendRecv(data, false)
	return n
}

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

func (m *display) SendCmd(cmd int, data []byte) int {
	n, _ := m.SendRecvCmd(cmd, data, false)
	return n
}

func (m *display) Echo(data []byte) (int, []byte) {
	return m.SendRecvCmd(0xFF, data, true)
}

func (m *display) Reset() (int) {
	n := m.Send([]byte{0xFE, 0x01})
	return n
}


func (m *display) Text(data string) (int) {
	n := m.Send([]byte(data))
	return n
}

func (m *display) FontSize(size int) (int) {
	return m.SendCmd(0x33, []byte{byte(size)})
}

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

func (m *display) GetTextPoint() (int, []byte) {
	return m.SendRecvCmd(0x7A, nil, true)
}

func (m *display) ClrScreen() (int) {
	return m.SendCmd(0x58, nil)
}

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

func (m *display) TextColour(r, g, b int) int {
	data := []byte{byte(r), byte(g), byte(b)}

	return m.SendCmd(0x2E, data)
}

func (m *display) PrintUTF8String(text string) int {

	return m.SendCmd(0x25, []byte(text))
}

func (m *display) PrintUnicodeString(text string) int {

	return m.SendCmd(0x24, []byte(text))
}

func (m *display) UpdateLabel(id, format int, value string) int  {
	data := []byte{byte(id), byte(format)}
	data = append(data, []byte(value)...)
	data = append(data, 0x00)

	return m.SendCmd(0x11, data)
}

func (m *display) UpdateLabelAscii(id int, value string) int  {
	return m.UpdateLabel(id, 0, value)
}

func (m *display) UpdateLabelUTF8(id int, value string) int  {
	return m.UpdateLabel(id, 2, value)
}

func (m *display) UpdateLabelUnicode(id int, value string) int  {
	return m.UpdateLabel(id, 1, value)
}

func (m *display) UpdateBargraphValue(id, value int) (int, []byte)  {
	data := []byte{byte(id)}
	valueb := make([]byte, 2)
	binary.BigEndian.PutUint16(valueb, uint16(value))

	data = append(data, valueb...)

	return m.SendRecvCmd(0x69, data, true)
}

func (m *display) UpdateTraceValue(id, value int) int  {
	data := []byte{byte(id)}
	valueb := make([]byte, 2)
	binary.BigEndian.PutUint16(valueb, uint16(value))

	data = append(data, valueb...)

	return m.SendCmd(0x75, data)
}


func (m *display) RunScript(filename string) int {
	data := []byte(filename)
	data = append(data, 0x00)
	n1, _ := m.SendRecvCmd(0x5D, data, false)
	//fmt.Printf("salida Run: %v\n", res)
	return n1
}

func (m *display) LoadBitmap(id int, filename string) (int, []byte) {
	data := []byte(filename)
	data = append(data, 0)
	n1, res := m.SendRecvCmd(0x5F, data, true)
	fmt.Printf("salida Run: %v\n", res)
	return n1, res
}



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

/**
func (m *display) BackLigthBrightness(int brightness) (int) {

	data := []byte{0xFE, 0x99, byte(brightness)}
	n3 := m.Send(data)
	return n3
}

func (m *display) Font(id int) (int) {
	data := []byte{0xFE, 0x31, byte(id)}
	n3, _ := m.SendRecv(data)
	return n3
}

func (m *display) TextWindow(x, y, width, height int) int {
	data := []byte{0xFE, 0x2B}
	xb := make([]byte,2)
	yb := make([]byte,2)
	widthb := make([]byte,2)
	heightb := make([]byte,2)
	binary.LittleEndian.Putint16(xb, uint16(x))
	binary.LittleEndian.Putint16(yb, uint16(y))
	binary.LittleEndian.Putint16(widthb, uint16(width))
	binary.LittleEndian.Putint16(heightb, uint16(height))
	data = append(data, xb...)
	data = append(data, yb...)
	data = append(data, widthb...)
	data = append(data, heightb...)

	n1 := m.Send(data)
	return n1
}

func (m *display) SetTextWindow(id int) (int) {
	data := []byte{0xFE, 0x2A, byte(id)}
	n1 := m.Send(data)
        return n1
}

func (m *display) ClrWindow(id int) (int) {
	n := m.Send([]byte{0xFE, 0x2C, byte(id)})
	return n
}

func (m *display) Rectangle(colour, x1, y1, x2, y2 int) int {
	data := []byte{0xFE, 0x72, byte(colour), byte(x1), byte(y1), byte(x2), byte(y2)}

	n1 := m.Send(data)
	return n1
}

func (m *display) AutoTransmKey(on bool) (int) {
	var dat1 []byte
	if on {
		dat1 = []byte{0xFE, 0x41}
	} else {
		dat1 = []byte{0xFE, 0x4F}
	}
	n := m.Send(dat1)

	return n
}
/**/

