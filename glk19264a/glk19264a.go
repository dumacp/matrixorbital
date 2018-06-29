package glk19264a


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

	buf := make([]byte,64)
	n, _ = m.port.Read(buf)

	res = append(res,buf[:n]...)

	return n, res
}

func (m *display) Send(data []byte) (int) {
	n, _ := m.SendRecv(data, false)
	return n
}


func (m *display) Text(data string) (int) {
	n := m.Send([]byte(data))
	return n
}

func (m *display) ColRow(col, row int) (int) {
	n := m.Send([]byte{0xFE, 0x47, byte(col), byte(row)})
	return n
}

func (m *display) ClrScreen() (int) {
	n := m.Send([]byte{0xFE, 0x58})
	return n
}

func (m *display) TextColRow(col, row int) func(data string) int {
	return func (data string) (int) {
		buf := []byte{0xFE, 0x47, byte(col), byte(row)}
		buf = append(buf, []byte(data)...)
		n := m.Send(buf)
		return n
	}
}

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

func (m *display) BitmapDraw(x, y, id int) (int) {
	data := []byte{0xFE, 0x62}
	idb := make([]byte, 2)
	binary.LittleEndian.PutUint16(idb, uint16(id))
	data = append(data, idb...)
	data = append(data, byte(x))
	data = append(data, byte(y))
	//fmt.Printf("data: % X, \nlen: %v\n", data, len(data))

	n3 := m.Send(data)
	return n3
}

func (m *display) BitmapDrawFile(x, y int, filename string) (int, error) {
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
	data = append(data, byte(x))
	data = append(data, byte(y))
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

func (m *display) BackLigthOff() (int) {

	data := []byte{0xFE, 0x46}
	n3 := m.Send(data)
	return n3
}

func (m *display) BackLigthON(min int) (int) {

	data := []byte{0xFE, 0x42, byte(min)}
	n3 := m.Send(data)
	return n3
}

func (m *display) Led(num, color int) (int) {

	data := []byte{0xFE, 0x5A, byte(num), byte(color)}
	n3 := m.Send(data)
	return n3
}

func (m *display) KeyPadOff() (int) {

	data := []byte{0xFE, 0x9B}
	n3 := m.Send(data)
	return n3
}

func (m *display) KeyPadON(level int) (int) {

	data := []byte{0xFE, 0x9C, byte(level)}
	n3 := m.Send(data)
	return n3
}

func (m *display) Font(id int) (int) {
	data := []byte{0xFE, 0x31}
	idb := make([]byte, 2)
	binary.LittleEndian.PutUint16(idb, uint16(id))
	data = append(data, idb...)
	n3 := m.Send(data)
	return n3
}


