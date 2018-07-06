package gtt43a


import (
	"github.com/tarm/serial"
	"sync"
	"time"
	"os"
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
	ClrWindow(int)		int
	Text(string)		int
	Send([]byte)		int
	SendRecv([]byte, bool)	(int, []byte)
	ColRow(int, int)	int
	TextColRow(int, int)	func(string) int
	BitmapUpload(int, string) (int, error)
	BitmapDraw(int, int, int) int
	BitmapDrawFile(int, int, string) (int, error)
	BitmapDrawData(int, int, []byte) int
	BuzzerActive(int, int) int
	BackLigthOff() int
	BackLigthON(int) int
	Led(int, int) int
	KeyPadOff()	int
	KeyPadON(int)	int
	Font(int)	int
	InitTextWindow(int, int, int, int, int, int, int, int, int) int
	SetTextWindow(int) int
	Rectangle(int, int, int, int, int) int
	WriteScratch(int, []byte) int
	ReadScratch(int, int) (int, []byte)
	AutoTransmKey(bool) int

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

func (m *display) sendRecv(data []byte, recv bool) (int, []byte) {
	m.mux.Lock()
	prefix := data[0]
	cmd := data[1]
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
	n, _ := m.sendRecv(data, false)
	return n
}

func (m *display) SendRecvCmd(cmd int, data []byte, recv bool) (int, []byte) {
	dat1 := []byte{0xFE, byte(cmd)}
	if data != nil {
		dat1 = append(data, data...)
	}

	n, res := sendRecv(dat1, recv)

	switch  {
	case n <= 0:
		return -1, nil
	case res == nil:
		return -2, nil
	case len(res) < 4:
		return -3, nil
	case res[0] != 0xFE:
		return -4, nil
	case res[1] != byte(cmd):
		return -5, nil
	}
	return n, res
}

func (m *display) SendCmd(cmd int, data []byte) int {
	dat1 := []byte{0xFE, byte(cmd)}

	n, _ := SendRecvCmd(dat1, nil, false)
	return n
}

func (m *display) Reset(data []byte) (int) {
	n := m.Send([]byte{0xFE, 0x01})
	return n
}


func (m *display) Text(data string) (int) {
	n := m.Send([]byte(data))
	return n
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

func (m *display) GetTextPoint(x, y int) (int, []byte) {
	return SendRecvCmd(0x7A, nil)
}

func (m *display) ClrScreen() (int) {
	return m.SendCmd(0x58)
}

func (m *display) TextPoint(x, y int) func(data string) int {
	return func (data string) (int) {
		n := m.TextInsertPoint(x, y)
		if n <= 0 {
			return n
		}
		buf = append(buf, []byte(data)...)
		n = m.Send(buf)
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

	return SendCmd(0x2B, data)
}

func (m *display) TextColour(r, g, b) int {
	data := []byte{byte(r), byte(g),byte(b)}

	return SendCmd(0x2E, data)
}



func (m *display) UpdateLabel(id, format int, value string) int  {
	data := []byte{byte(id), byte(formta)}
	data = append(data, []byte(value)...)

	return SendCmd(data)
}

func (m *display) UpdateLabelAscii(id, value string) int  {
	return UpdateLabel(id, 0, value)
}

func (m *display) UpdateLabelUTF8(id, value string) int  {
	return UpdateLabel(id, 2, value)
}

func (m *display) UpdateLabelUnicode(id, value string) int  {
	return UpdateLabel(id, 1, value)
}



/**

func (m *display) BuzzerActive(frec, time int) (int) {

	data := []byte{0xFE, 0xBB}
	frecb := make([]byte, 2)
	timeb := make([]byte, 2)
	binary.LittleEndian.PutUint16(frecb, uint16(frec))
	binary.LittleEndian.PutUint16(timeb, uint16(time))
	data = append(data, frecb...)
	data = append(data, timeb...)
	n3 := m.Send(data)
	return n3
}

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

func (m *display) WriteScratch(addr int, data []byte) (int) {
	dat1 := []byte{0xFE, 0xCC}
	addrb := make([]byte,2)
	sizeb := make([]byte,2)
	binary.LittleEndian.PutUint16(addrb, uint16(addr))
	binary.LittleEndian.PutUint16(sizeb, uint16(len(data)))
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
	binary.LittleEndian.PutUint16(addrb, uint16(addr))
	binary.LittleEndian.PutUint16(sizeb, uint16(size))
	dat1 = append(dat1, addrb...)
	dat1 = append(dat1, sizeb...)
	n, datOut := m.SendRecv(dat1, true)

	return n, datOut
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

