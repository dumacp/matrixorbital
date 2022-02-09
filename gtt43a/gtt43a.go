/**
Package to send commands and recieve response to and from gtt43a device.
**/
package gtt43a

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	_ "os"
	"sync"
	"time"

	"github.com/tarm/serial"
)

type PortOptions struct {
	Port        string
	Baud        int
	ReadTimeout time.Duration
}

type Display interface {
	Open() error
	Close() bool
	ClrScreen() error
	Text(string) error
	FontSize(int) error
	Send([]byte) error
	recv() ([]byte, error)
	Recv() ([]byte, error)
	SendRecv([]byte) ([]byte, error)
	Echo([]byte) ([]byte, error)
	Version() ([]byte, error)
	SendRecvCmd(int, []byte) ([]byte, error)
	SendCmd(int, []byte) error
	Reset() error
	TextInsertPoint(int, int) error
	GetTextPoint() ([]byte, error)
	TextPoint(int, int) func(data string) error
	TextWindow(int, int, int, int) error
	TextColour(int, int, int) error
	PrintUTF8String(string) error
	PrintUnicode([]byte) error
	UpdateLabel(id, format int, value []byte) error
	UpdateLabelAscii(int, string) error
	UpdateLabelUTF8(int, string) error
	UpdateLabelUnicode(int, []byte) error
	UpdateBargraphValue(int, int) ([]byte, error)
	UpdateTraceValue(int, int) error
	RunScript(string) error
	LoadBitmap(int, string) ([]byte, error)
	BuzzerActive(frec, time int) error
	SetPropertyValueU16(id int, prpType GTT25PropertyType) func(value int) error
	SetPropertyValueS16(id int, prpType GTT25PropertyType) func(value int) error
	SetPropertyValueU8(id int, prpType GTT25PropertyType) func(value int) error
	SetPropertyText(id int, prpType GTT25PropertyType) func(text string) error
	GetPropertyValueU16(id int, prpType GTT25PropertyType) func() ([]byte, error)
	GetPropertyValueS16(id int, prpType GTT25PropertyType) func() ([]byte, error)
	GetPropertyValueU8(id int, prpType GTT25PropertyType) func() (byte, error)
	/**
	ApduSetPropertyValueU16(id int, prpType GTT25PropertyType, value int) []byte
	ApduSetPropertyValueS16(id int, prpType GTT25PropertyType, value int) []byte
	ApduSetPropertyValueU8(id int, prpType GTT25PropertyType, value int) []byte
	ApduSetPropertyText(id int, prpType GTT25PropertyType, text string) []byte
	ApduGetPropertyValueU16(id int, prpType GTT25PropertyType) []byte
	ApduGetPropertyValueS16(id int, prpType GTT25PropertyType) []byte
	ApduGetPropertyValueU8(id int, prpType GTT25PropertyType) []byte
	/**/
	ChangeTouchReporting(style int) error
	GetTouchReporting() ([]byte, error)

	GetToggleState(id int) ([]byte, error)
	GetSliderValue(id int) ([]byte, error)

	WriteScratch(addr int, data []byte) error
	ReadScratch(addr, size int) ([]byte, error)
	Listen() error
	Events() (chan *Event, error)
	StopListen()
	AnimationStartStop(id, action int) error
	AnimationSetFrame(id, state int) error
	AnimationStopAll() error
}

type display struct {
	options *PortOptions
	status  uint32
	port    *serial.Port
	muxSend sync.Mutex
	// muxRecv    sync.Mutex
	bufResp    []byte
	chEvent    chan []byte
	stopListen chan int
}

const (
	OPENED uint32 = iota
	CLOSED
	LISTEN
)

const (
	OFF int = iota
	GREEN
	RED
	YELLOW
)

const (
	timeoutRead   time.Duration = 300 * time.Millisecond
	bufferLen     int           = 1024
	minReadWait   time.Duration = 40 * time.Millisecond
	maxCountError int           = 5
)

//Create a new Display device
func NewDisplay(opt *PortOptions) Display {
	disp := &display{}
	disp.options = opt
	disp.status = CLOSED
	disp.stopListen = make(chan int)
	return disp
}

//Open device comunication channel
func (m *display) Open() error {
	if m.status == OPENED || m.status == LISTEN {
		return nil
	}

	config := &serial.Config{
		Name:        m.options.Port,
		Baud:        m.options.Baud,
		ReadTimeout: m.options.ReadTimeout,
	}

	var err error
	m.port, err = serial.OpenPort(config)
	if err != nil {
		return err
	}

	m.status = OPENED
	return nil
}

//Clsoe device comunication channel
func (m *display) Close() bool {
	defer func() {
		m.status = CLOSED
	}()
	if m.port == nil {
		return false
	}
	if err := m.port.Close(); err != nil {
		return false
	}

	return true
}

//Listen is a go rutine that listening serial port to detect messages
//Return channel with  messages (Event struct)
func (m *display) Listen() error {
	if m.status == LISTEN {
		return fmt.Errorf("error: already Listening display")
	}
	if m.status != OPENED {
		return fmt.Errorf("error: port serial is closed")
	}
	countError := 0
	m.bufResp = make([]byte, 0)
	m.chEvent = make(chan []byte, 3)
	log.Println("START listen")
	ch := make(chan []byte)
	go func() {
		defer func() {
			close(m.chEvent)
			close(ch)
			m.status = OPENED
		}()
		funcRead := func(v []byte) {
			//log.Printf("read serial port: [% X]\n", v)
			lenValue := 0
			//log.Printf("vector: [% X]\n", v)
			if len(v) > 2 && bytes.Equal(v[:2], []byte{0xFC, 0xEB}) {
				if len(v) > 8 {
					lenValue = int(binary.BigEndian.Uint16(v[2:4]))
					msg := make([]byte, 0)
					msg = append(msg, v[1])
					msg = append(msg, v[4:lenValue+4]...)
					select {
					case m.chEvent <- msg:
					case <-time.After(timeoutRead):
						// default:
						log.Printf("evento ????X [% X]\n", msg)
					}
				}
			} else if len(v) > 2 && bytes.Equal(v[:2], []byte{0xFC, 0x87}) {
				if len(v) > 5 {
					lenValue = int(binary.BigEndian.Uint16(v[2:4]))
					msg := make([]byte, 0)
					msg = append(msg, v[1])
					msg = append(msg, v[4:lenValue+4]...)
					select {
					case m.chEvent <- msg:
					case <-time.After(timeoutRead):
						// default:
						log.Printf("evento ????X [% X]\n", msg)
					}
				}
			} else if len(v) > 2 && v[0] == byte(0xFC) {
				if len(v) > 4 {
					lenValue = int(binary.BigEndian.Uint16(v[2:4]))
					if len(v) < lenValue+4 {
						return
					}
					m.bufResp = v[:lenValue+4]
					//log.Printf("respuesta low [% X]\n", m.bufResp)
				}
			}
		}

		// buf := make([]byte, 0)
		for {
			/**/
			select {
			case <-m.stopListen:
				return
			default:
			}
			/**/
			//log.Printf("bajo nivel 1\n")
			buf, err := m.recv()
			if err != nil {
				if countError >= maxCountError {
					return
				}
				countError++
			}
			if len(buf) <= 0 {
				continue
			}
			// //log.Printf("bajo nivel 2: len: %d, [% X]\n", len(res), res)
			// buf = append(buf, res...)
			// if n >= bufferLen {
			// 	continue
			// }
			for {
				if len(buf) > 0 && buf[0] == byte(0xFC) {
					if len(buf) > 4 {
						lenValue := int(binary.BigEndian.Uint16(buf[2:4]))
						if len(buf) >= lenValue+4 {
							//log.Printf("Atrapado\n")
							funcRead(buf[0 : lenValue+4])
							//log.Printf("liberado\n")
							buf = buf[4+lenValue:]
							//log.Printf("bajo nivel 3: [% X]\n", buf)
							continue
						}
					}
				}
				// buf = make([]byte, 0)
				break
			}
			// time.Sleep(20 * time.Millisecond)
		}
	}()
	m.status = LISTEN
	return nil
}

func (m *display) StopListen() {
	go func() {
		select {
		case m.stopListen <- 1:
		case <-time.After(10 * time.Second):
		}
	}()
}

//Primitive function to send and recieve bytes to and from display device.
//recv, flag to wait a response form device.
func (m *display) SendRecv(data []byte) ([]byte, error) {
	// m.muxSend.Lock()
	// defer m.muxSend.Unlock()
	// res := make([]byte, 0)
	m.bufResp = make([]byte, 0)

	if err := m.send(data); err != nil {
		return nil, err
	}

	//log.Printf("request End: [% X]\n", data)
	if m.status == LISTEN {
		tAfter1 := time.After(timeoutRead)
		tick1 := time.NewTicker(10 * time.Millisecond)
		defer tick1.Stop()
		for {
			select {
			case <-tick1.C:
				if len(m.bufResp) > 0 {
					//log.Printf("response End 1: [% X]\n", m.bufResp)
					res := m.bufResp
					return res[:], nil
				}
			case <-tAfter1:
				log.Println("timeoutRead")
				return nil, ErrorDevTimeout
			}
		}
	}
	return m.recv()
}

//Send bytes data to device. Don't wait response.
func (m *display) Send(data []byte) error {

	if m.status == CLOSED {
		return fmt.Errorf("device CLOSED")
	}
	return m.send(data)
}

func (m *display) send(data []byte) error {
	m.muxSend.Lock()
	defer m.muxSend.Unlock()
	if data == nil {
		return ErrorDevNull
	}
	if m.status == CLOSED {
		return ErrorDevClosed
	}
	n, err := m.port.Write(data)
	if err != nil || n <= 0 {
		return fmt.Errorf("error Write: %w", err)

	}
	if n <= 0 {
		return ErrorDevEmptyWrite
	}
	/**
	log.Printf("request: [% X]\n", data)
	/**/
	return nil
}

/**/
func (m *display) Recv() ([]byte, error) {
	//m.muxRecv.Lock()
	//defer m.muxRecv.Unlock()
	m.bufResp = make([]byte, 0)
	if m.status == LISTEN {
		tAfter := time.After(timeoutRead)
		tTick := time.NewTicker(10 * time.Millisecond)
		defer tTick.Stop()
		for {
			select {
			case <-tTick.C:
				if len(m.bufResp) > 0 {
					res := make([]byte, len(m.bufResp))
					copy(res, m.bufResp)
					return res, nil
				}
			case <-tAfter:
				return nil, ErrorDevTimeout
			}
		}
	}
	//	log.Printf("NOT LISTEN: %d\n", m.status)
	return m.recv()
}

/**/

//Primitive function to send and recieve bytes to and from display device.
//recv, flag to wait a response form device.
func (m *display) recv() ([]byte, error) {

	if m.status == CLOSED {
		return nil, ErrorDevClosed
	}

	if m.port == nil {
		return nil, ErrorDevNull
	}

	reader := bufio.NewReader(m.port)

	buf := make([]byte, bufferLen)
	tn := time.Now()
	n, err := reader.Read(buf)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
		if m.options.ReadTimeout > minReadWait && time.Since(tn) < minReadWait {
			return nil, err
		}
	}
	if n <= 0 {
		return nil, nil
	}
	//log.Printf("parcial response 1: [% X]\n", res)
	return buf[:n], nil
}

//Send a command to display device
//cmd, id for the command
//wait response
func (m *display) SendRecvCmd(cmd int, data []byte) ([]byte, error) {
	// m.muxSend.Lock()
	// defer m.muxSend.Unlock()
	dat1 := []byte{0xFE, byte(cmd)}
	if data != nil {
		dat1 = append(dat1, data...)
	}
	var res []byte

	//n, res := m.SendRecv(dat1)
	if err := m.send(dat1); err != nil {
		return nil, err
	}

	//log.Printf("request End: [% X]\n", dat1)
	if m.status == LISTEN {
		after := time.After(timeoutRead)
		tick := time.NewTicker(10 * time.Millisecond)
		defer tick.Stop()
		m.bufResp = make([]byte, 0)
	for_src:
		for {
			select {
			case <-tick.C:
				if len(m.bufResp) > 1 {
					//log.Printf("response End 1: [% X]\n", m.bufResp)
					res = make([]byte, len(m.bufResp))
					copy(res, m.bufResp)
					if res[1] != byte(cmd) {
						log.Printf("Other response: [% X]\n", m.bufResp)
						m.bufResp = make([]byte, 0)
						continue
					}
					break for_src
				}
			case <-after:
				log.Println("timeoutRead")
				return res, ErrorDevTimeout
			}
		}
	} else {
		// time.Sleep(timeoutRead * time.Millisecond)
		return m.recv()
	}
	//log.Printf("SendRecvCmd response: [% X]", res)

	switch {
	case len(res) <= 0:
		return nil, ErrorDevEmptyRead
	case len(res) < 4:
		return nil, errors.New("incomplete response")
	case res[0] != byte(0xFC):
		return nil, errors.New("wrong response")
	case res[1] != byte(cmd):
		return nil, fmt.Errorf("incorrect response, CMD: [%X]", res[1])
	}
	return res[4:], nil
}

//Send a Command to display device.
//don't wait response
func (m *display) SendCmd(cmd int, data []byte) error {
	dat1 := []byte{0xFE, byte(cmd)}
	if data != nil {
		dat1 = append(dat1, data...)
	}

	return m.Send(dat1)
}

//Send echo data and to wait for a response.
func (m *display) Echo(data []byte) ([]byte, error) {
	return m.SendRecvCmd(0xFF, data)
}

//Send reset command to display device
func (m *display) Reset() error {
	return m.Send([]byte{0xFE, 0x01})
}

//Request Version and wait for a response.
func (m *display) Version() ([]byte, error) {
	return m.SendRecvCmd(0x00, nil)
}

//Clear actual Screen
func (m *display) ClrScreen() error {
	return m.SendCmd(0x58, nil)
}

//Run script binary. The filename path is a local path in display device
func (m *display) RunScript(filename string) error {
	data := []byte(filename)
	data = append(data, 0x00)
	if err := m.SendCmd(0x5D, data); err != nil {
		return err
	}
	//fmt.Printf("salida Run: %v\n", res)
	//time.Sleep(1 * time.Second)
	if _, err := m.Recv(); err != nil {
		return err
	}
	return nil
}

//Load in display memory a bitmap object from filename. The filename path in a local in display device.
func (m *display) LoadBitmap(id int, filename string) ([]byte, error) {
	data := []byte(filename)
	data = append(data, 0)
	res, err := m.SendRecvCmd(0x5F, data)
	//fmt.Printf("salida Run: %v\n", res)
	return res, err
}

//Active buzzer in device.
//frec, is the frecuency of the signal
//time, is the duration of the signal
func (m *display) BuzzerActive(frec, time int) error {

	data := make([]byte, 0)
	frecb := make([]byte, 2)
	timeb := make([]byte, 2)
	binary.BigEndian.PutUint16(frecb, uint16(frec))
	binary.BigEndian.PutUint16(timeb, uint16(time))
	data = append(data, frecb...)
	data = append(data, timeb...)
	return m.SendCmd(0xBB, data)
}

//TOUCH

//Change Touch Reporting Style
func (m *display) ChangeTouchReporting(style int) error {
	return m.SendCmd(0x87, []byte{byte(style)})
}

//Get Touch Reporting Style
func (m *display) GetTouchReporting() ([]byte, error) {
	return m.SendRecvCmd(0x88, nil)
}

//Create a touch region
//regId, region ID
//x, y, coordinate of the touch region
//width, width of the region
//height, height of the region
/**/

func (m *display) GetToggleState(id int) ([]byte, error) {
	return m.SendRecvCmd(171, []byte{byte(id & 0xFF)})
}

func (m *display) GetSliderValue(id int) ([]byte, error) {
	return m.SendRecvCmd(167, []byte{byte(id & 0xFF)})
}

func (m *display) WriteScratch(addr int, data []byte) error {
	dat1 := make([]byte, 0)
	addrb := make([]byte, 2)
	sizeb := make([]byte, 2)
	binary.BigEndian.PutUint16(addrb, uint16(addr))
	binary.BigEndian.PutUint16(sizeb, uint16(len(data)))
	dat1 = append(dat1, addrb...)
	dat1 = append(dat1, sizeb...)
	dat1 = append(dat1, data...)
	return m.SendCmd(0xCC, dat1)
}

func (m *display) ReadScratch(addr, size int) ([]byte, error) {
	dat1 := make([]byte, 0)
	addrb := make([]byte, 2)
	sizeb := make([]byte, 2)
	binary.BigEndian.PutUint16(addrb, uint16(addr))
	binary.BigEndian.PutUint16(sizeb, uint16(size))
	dat1 = append(dat1, addrb...)
	dat1 = append(dat1, sizeb...)
	datOut, err := m.SendRecvCmd(0xCD, dat1)
	if err != nil {
		return nil, err
	}
	if len(datOut) < 3 {
		return nil, fmt.Errorf("bad response: [% X]", datOut)
	}

	return datOut[2:], nil
}

func (m *display) AnimationStartStop(id, action int) error {
	return m.SendCmd(0xC2, []byte{byte(id), byte(action)})
}

func (m *display) AnimationSetFrame(id, state int) error {
	return m.SendCmd(0xC3, []byte{byte(id), byte(state)})
}

func (m *display) AnimationStopAll() error {
	return m.SendCmd(0xC6, nil)
}
