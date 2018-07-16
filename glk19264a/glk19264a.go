/**
Package to send commands and recieve response to and from glk19264a device.
**/
package glk19264a


import (
	"github.com/tarm/serial"
	"sync"
	"time"
	"os"
	"fmt"
	"encoding/binary"
)

//Options to open comunication channel with device
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
	PollKey() []int

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

//Create a new display
func NewDisplay(opt *PortOptions) Display {
	disp := &display{}
	disp.options = opt
	disp.status = CLOSED

	return disp
}

//Open communication channel with display device
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

//Close communication channel with display device
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

/**
Primitive function to send and recieve bytes to device.
data, is the bytes to send to device
recv, is a flag to wait response
The function return the count number the bytes in response and the slice with the response.
**/
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

//Function to only send bytes to device. Don't wait response.
//The function return the count number of bytes wrote in the buffer of serial device.
func (m *display) Send(data []byte) (int) {
	n, _ := m.SendRecv(data, false)
	return n
}

//Function to print data string in display. The text is to print in actual (col,row) point.
func (m *display) Text(data string) (int) {
	n := m.Send([]byte(data))
	return n
}

//Function to set (col,row) point in the display area.
func (m *display) ColRow(col, row int) (int) {
	n := m.Send([]byte{0xFE, 0x47, byte(col), byte(row)})
	return n
}

//Function to clear display area and set actual window to 0 (zero)
func (m *display) ClrScreen() (int) {
	n := m.Send([]byte{0xFE, 0x58})
	return n
}

//Function to set (col,row) point and print text in display from that point
func (m *display) TextColRow(col, row int) func(data string) int {
	return func (data string) (int) {
		buf := []byte{0xFE, 0x47, byte(col), byte(row)}
		buf = append(buf, []byte(data)...)
		n := m.Send(buf)
		return n
	}
}

/**
Function to load bitmap in the buffer display
id, is the ID of the bitmap in the memory display
**/
func (m *display) BitmapUpload(id int, filename string) (int, error) {
	file, err1 := os.Open(filename)
	if err1 != nil {
		return 0, err1
	}

	b := make([]byte,1024)
	n2, err2 := file.Read(b)
	if err2 != nil {
		return 0, err2
	}

	idb := make([]byte, 2)
	size := make([]byte, 4)
	binary.LittleEndian.PutUint16(idb, uint16(id))
	//binary.BigEndian.PutUint16(idb, uint16(id))
	binary.LittleEndian.PutUint32(size, uint32(n2))
	//binary.BigEndian.PutUint32(size, uint32(n2))

	data := []byte{0xFE, 0x5E}
	data = append(data, idb...)
	data = append(data, size...)
	data = append(data, b[:n2]...)
	fmt.Printf("data: % X\nlen: %v\n", data, len(data))

	data1 := []byte{0xfe,0x36}
	data2 := []byte{0xfe,0xf5,0x4d,0x4f,0x75,0x6e}

	m.Send(data1)
	m.Send(data2)

	n5 := m.Send(data)
	return n5, nil
}

//Function to print the bitmap identify with id in the memory in display 
func (m *display) BitmapDraw(col, row, id int) (int) {
	data := []byte{0xFE, 0x62}
	idb := make([]byte, 2)
	binary.LittleEndian.PutUint16(idb, uint16(id))
	data = append(data, idb...)
	data = append(data, byte(col))
	data = append(data, byte(row))
	//fmt.Printf("data: % X, \nlen: %v\n", data, len(data))

	n3 := m.Send(data)
	return n3
}

//Function to print bitmap in display from a filename path
func (m *display) BitmapDrawFile(col, row int, filename string) (int, error) {
	file, err1 := os.Open(filename)
	if err1 != nil {
		return 0, err1
	}

	b := make([]byte,1024)
	n2, err2 := file.Read(b)
	if err2 != nil {
		return 0, err2
	}

	data := []byte{0xFE, 0x64}
	data = append(data, byte(col))
	data = append(data, byte(row))
	data = append(data, b[:n2]...)
	fmt.Printf("data: %v, \nlen: %v\n", data, len(data))


	n3 := m.Send(data)
	return n3, nil
}

func (m *display) BitmapDrawData(x, y int, dat []byte) (int) {

	data := []byte{0xFE, 0x64}
	data = append(data, byte(x))
	data = append(data, byte(y))
	data = append(data, dat...)
	fmt.Printf("data: %v, \nlen: %v\n", dat, len(dat))

	n3 := m.Send(data)
	return n3
}

//Function to active the buzzer device
//frec, is the frecuency of the signal 
//time, is the duration time of the signal
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

//Function to OFF the backlight
func (m *display) BackLigthOff() (int) {

	data := []byte{0xFE, 0x46}
	n3 := m.Send(data)
	return n3
}

//Function to ON the backlight
func (m *display) BackLigthON(min int) (int) {

	data := []byte{0xFE, 0x42, byte(min)}
	n3 := m.Send(data)
	return n3
}

//Function to assign a color to LED, id is the identity number of LED
func (m *display) Led(num, color int) (int) {

	data := []byte{0xFE, 0x5A, byte(num), byte(color)}
	n3 := m.Send(data)
	return n3
}

//Function to OFF keypad light
func (m *display) KeyPadOff() (int) {

	data := []byte{0xFE, 0x9B}
	n3 := m.Send(data)
	return n3
}

//Function to ON keypad light
func (m *display) KeyPadON(level int) (int) {

	data := []byte{0xFE, 0x9C, byte(level)}
	n3 := m.Send(data)
	return n3
}

//Function to set FONT
func (m *display) Font(id int) (int) {
	data := []byte{0xFE, 0x31}
	idb := make([]byte, 2)
	binary.LittleEndian.PutUint16(idb, uint16(id))
	data = append(data, idb...)
	n3 := m.Send(data)
	return n3
}

//Function to create a window text
func (m *display) InitTextWindow(id, x1, y1, x2, y2, font, charSpace, lineSpace, scroll int) int {
	data := []byte{0xFE, 0x2B, byte(id), byte(x1), byte(y1), byte(x2), byte(y2)}
	fontb := make([]byte,2)
	binary.LittleEndian.PutUint16(fontb, uint16(font))
	data = append(data, fontb...)
	data = append(data, byte(charSpace))
	data = append(data, byte(lineSpace))
	data = append(data, byte(scroll))

	n1 := m.Send(data)
	return n1
}

//Function to set the actual window text
func (m *display) SetTextWindow(id int) (int) {
	data := []byte{0xFE, 0x2A, byte(id)}
	n1 := m.Send(data)
        return n1
}

//Function to clear window identify with id number
func (m *display) ClrWindow(id int) (int) {
	n := m.Send([]byte{0xFE, 0x2C, byte(id)})
	return n
}

//Function to print a rectangle in display area
func (m *display) Rectangle(colour, x1, y1, x2, y2 int) int {
	data := []byte{0xFE, 0x72, byte(colour), byte(x1), byte(y1), byte(x2), byte(y2)}

	n1 := m.Send(data)
	return n1
}

//Function to write scratch data in the display memory. This data is inthe volatile memory
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

//Function to read scratch data in the display memory.
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

//Function to enable transmit bytes from display device in event key button.
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

//Function to read key button state
func (m *display) PollKey() []int {
	resp := make([]int,0)
	dat1 := []byte{0xFE, 0x26}
	n, datOut := m.SendRecv(dat1, true)
	//fmt.Printf("data: % X, \nlen: %v\n", datOut, len(datOut))

	if n <= 0 {
		return nil
	}

	if datOut[0] > 0x00 {
		resp = append(resp, int(datOut[0] & 0x7F))
	} else {
		return nil
	}

	if datOut[0] > 0x80 {
		resp1 := make([]int,0)
		for resp1 != nil {
			resp1 = m.PollKey()
			if resp1 != nil && len(resp1) > 0 {
				resp = append(resp, resp1...)
			}
		}
	}

	return resp
}


