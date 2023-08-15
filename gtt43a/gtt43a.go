/*
*
Package to send commands and recieve response to and from gtt43a device.
*
*/
package gtt43a

import (
	"bufio"
	"bytes"
	"context"
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
	Close() error
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
	SetLabelBackgroundColour(id, r, g, b int) error
	CreateLabelLegacy(id, x, y, width, height, h, v, font, r, g, b int) error
	UpdateBargraphValue(int, int) ([]byte, error)
	UpdateTraceValue(int, int) error
	RunScript(string) error
	LoadBitmapLegcay(id int, filename string) error
	DisplayBitmapLegcay(id int, x, y int) error
	ClearBitmapLegacy(id int) error
	SetBitmapTransparencyLegacy(id int, r, g, b int) error
	SetBacklightLegcay(brightness int) error

	BitmapLoad(int, string) error
	BitmapCapture(id int, left, top, width, height int) error
	BuzzerActive(frec, time int) error
	CreateObject(id int, objectType GTT25ObjectType) error
	DestroyObject(id int) error
	BaseObjectBeginUpdate(id int) error
	BaseObjectEndUpdate(id int) error
	ObjectListGet(id, itemIndex int) error
	SetPropertyValueU16(id int, prpType GTT25PropertyType) func(value int) error
	SetPropertyValueS16(id int, prpType GTT25PropertyType) func(value int) error
	SetPropertyValueU8(id int, prpType GTT25PropertyType) func(value int) error
	SetPropertyText(id int, prpType GTT25PropertyType) func(text string) error
	GetPropertyValueU16(id int, prpType GTT25PropertyType) func() ([]byte, error)
	GetPropertyValueS16(id int, prpType GTT25PropertyType) func() ([]byte, error)
	GetPropertyValueU8(id int, prpType GTT25PropertyType) func() (byte, error)
	GetPropertyText(id int, prpType GTT25PropertyType) func() ([]byte, error)

	ChangeTouchReporting(style int) error
	GetTouchReporting() ([]byte, error)

	GetToggleState(id int) ([]byte, error)
	GetSliderValue(id int) ([]byte, error)

	WriteScratch(addr int, data []byte) error
	ReadScratch(addr, size int) ([]byte, error)
	Listen() error
	ListenWithContext(ctx context.Context) error
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
	mux     sync.Mutex
	wmux    sync.Mutex
	// muxRecv    sync.Mutex
	bufResp chan []byte
	chEvent chan []byte
	cancel  func()
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
	timeoutRead   time.Duration = 600 * time.Millisecond
	bufferLen     int           = 1024
	maxCountError int           = 5
)

// Create a new Display device
func NewDisplay(opt *PortOptions) Display {
	disp := &display{}
	disp.options = opt
	disp.status = CLOSED
	return disp
}

// Open device comunication channel
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

// Clsoe device comunication channel
func (m *display) Close() error {
	fmt.Println("CLOSE gtt43a")
	defer func() {
		m.status = CLOSED
	}()
	if m.cancel != nil {
		m.cancel()
	}
	if m.port == nil {
		return nil
	}
	if err := m.port.Close(); err != nil {
		return err
	}

	return nil
}

// Listen is a go rutine that listening serial port to detect messages
// Return channel with  messages (Event struct)
func (m *display) Listen() error {
	return m.ListenWithContext(context.TODO())
}

// ListenWithContext is a go rutine that listening serial port to detect messages
// Return channel with  messages (Event struct)
func (m *display) ListenWithContext(contxt context.Context) error {
	if m.status == LISTEN {
		return fmt.Errorf("error: already Listening display")
	}
	if m.status != OPENED {
		return fmt.Errorf("error: port serial is closed")
	}
	if m.cancel != nil {
		m.cancel()
	}

	var ctx context.Context
	var cancel func()
	if contxt != nil {
		ctx, cancel = context.WithCancel(contxt)
		m.cancel = cancel
	} else {
		ctx, cancel = context.WithCancel(context.TODO())
		m.cancel = cancel
	}

	countError := 0
	m.bufResp = make(chan []byte)
	m.chEvent = make(chan []byte)
	fmt.Println("START listen")
	ch := make(chan []byte)
	go func() {
		defer func() {
			fmt.Println("STOP listen")
			close(m.chEvent)
			close(ch)
			if cancel != nil {
				cancel()
			}
			if m.status == LISTEN {
				m.status = OPENED
			}
		}()

		funcRead := func(v []byte) {
			lenValue := 0
			if len(v) <= 0 {
				return
			} else if len(v) > 2 && bytes.Equal(v[:2], []byte{0xFC, 0xEB}) {
				log.Printf("respuesta low 0 [% X]\n", v[:])
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
				log.Printf("respuesta low 1 [% X]\n", v[:])
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
			} else if len(v) > 2 && bytes.Equal(v[:2], []byte{0xFC, 0xFB}) {
				log.Printf("respuesta low 2 [% X]\n", v[:])
				select {
				case m.bufResp <- v[:lenValue+4]:
				default:
					log.Printf("msg ????X [% X]\n", v[:lenValue+4])
				}
			} else if len(v) > 4 && v[0] == byte(0xFC) {
				log.Printf("respuesta low 3 [% X]\n", v[:])
				lenValue = int(binary.BigEndian.Uint16(v[2:4]))
				if len(v) < lenValue+4 {
					return
				}
				log.Printf("respuesta low 3 [% X]\n", v[:])
				select {
				case m.bufResp <- v[:lenValue+4]:
				default:
					log.Printf("msg ????X [% X]\n", v[:lenValue+4])
				}
			} else {
				log.Printf("respuesta low 4 [% X]\n", v[:])
				select {
				case m.bufResp <- v[:]:
				default:
					log.Printf("msg ????X [% X]\n", v[:])
				}
			}
		}
		for {
			/**/
			select {
			case <-ctx.Done():
				return
			default:
			}
			buf, err := m.recv()
			if err != nil {
				if countError >= maxCountError {
					return
				}
				countError++
				continue
			}
			countError = 0
			if len(buf) <= 0 {
				continue
			}
			for {
				if len(buf) > 0 && buf[0] == byte(0xFC) {
					if len(buf) > 4 {
						lenValue := int(binary.BigEndian.Uint16(buf[2:4]))
						if len(buf) >= lenValue+4 {
							funcRead(buf[0 : lenValue+4])
							buf = buf[4+lenValue:]
							continue
						}
					}
					funcRead(buf[:])
					break
				} else {
					funcRead(buf[:])
				}
				break
			}
		}
	}()
	m.status = LISTEN
	return nil
}

func (m *display) StopListen() {
	if m.cancel != nil {
		m.cancel()
	}
}

// Primitive function to send and recieve bytes to and from display device.
// recv, flag to wait a response form device.
func (m *display) SendRecv(data []byte) ([]byte, error) {
	m.wmux.Lock()
	defer m.wmux.Unlock()
	fmt.Println("SendRecv ########")
	defer fmt.Println("end SendRecv ########")

	if err := m.send(data); err != nil {
		return nil, err
	}

	if m.status == LISTEN {
		tAfter1 := time.After(timeoutRead)
		count := 0
		for {
			select {
			case res := <-m.bufResp:
				fmt.Printf("count: %d\n", (count))
				count++
				if len(res) > 0 {
					return res[:], nil
				}
			case <-tAfter1:
				log.Println("timeoutRead ////")
				return nil, ErrorDevTimeout
			}
		}
	}
	/**/
	res, err := m.recv()
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Send bytes data to device. Don't wait response.
func (m *display) Send(data []byte) error {
	if m.status == CLOSED {
		return fmt.Errorf("device CLOSED")
	}
	return m.send(data)
}

func (m *display) send(data []byte) error {

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
	/**/
	log.Printf("request: [% X]\n", data)
	/**/
	return nil
}

/**/
func (m *display) Recv() ([]byte, error) {

	if m.status == LISTEN {
		count := 0
		for {
			select {
			case res := <-m.bufResp:
				fmt.Printf("count: %d\n", (count))
				count++
				if len(res) > 0 {
					return res[:], nil
				}
			case <-time.After(timeoutRead):
				return nil, ErrorDevTimeout
			}
		}
	}

	res, err := m.recv()
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Primitive function to send and recieve bytes to and from display device.
// recv, flag to wait a response form device.
func (m *display) recv() ([]byte, error) {

	m.mux.Lock()
	defer m.mux.Unlock()

	if m.status == CLOSED {
		return nil, ErrorDevClosed
	}

	if m.port == nil {
		return nil, ErrorDevNull
	}

	reader := bufio.NewReader(m.port)
	tn := time.Now()
	buf, err := reader.ReadBytes('\xFE')
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
		if m.options.ReadTimeout > 0 && time.Since(tn) < m.options.ReadTimeout/10 {
			return nil, err
		}
		// fmt.Println("recv timeout")
	}
	n := len(buf)
	if n <= 0 {
		return nil, nil
	}

	response := make([]byte, 0)
	response = append(response, buf[:n]...)
	log.Printf("Response_0: [% X]\n", buf[:n])
	return response, nil
}

// Send a command to display device
// cmd, id for the command
// wait response
func (m *display) SendRecvCmd(cmd int, data []byte) ([]byte, error) {
	m.wmux.Lock()
	defer m.wmux.Unlock()
	fmt.Println("SendRecvCmd ########")
	defer fmt.Println("end SendRecvCmd ########")
	dat1 := []byte{0xFE, byte(cmd)}
	if data != nil {
		dat1 = append(dat1, data...)
	}
	var res []byte

	//n, res := m.SendRecv(dat1)
	if err := m.send(dat1); err != nil {
		return nil, err
	}

	if m.status == LISTEN {
		after := time.After(timeoutRead)
		tick := time.NewTicker(10 * time.Millisecond)
		defer tick.Stop()
		// count := 0
	for_src:
		for {
			select {
			case res = <-m.bufResp:
				// fmt.Printf("count: %d, %X\n", (count), res)
				// count++
				if len(res) > 1 && res[1] == byte(cmd) {
					fmt.Printf("SendRecvCmd response: %X\n", (res))
					break for_src
				}
				continue
			case <-after:
				log.Println("timeoutRead")
				return res, ErrorDevTimeout
			}
		}
	} else {
		/**/
		res, err := m.recv()
		if err != nil {
			return nil, err
		}
		return res, nil
		/**/
	}

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

// Send a Command to display device.
// don't wait response
func (m *display) SendCmd(cmd int, data []byte) error {
	dat1 := []byte{0xFE, byte(cmd)}
	if data != nil {
		dat1 = append(dat1, data...)
	}

	return m.send(dat1)
}

// Send echo data and to wait for a response.
func (m *display) Echo(data []byte) ([]byte, error) {
	return m.SendRecvCmd(0xFF, data)
}

// Send reset command to display device
func (m *display) Reset() error {
	return m.Send([]byte{0xFE, 0x01})
}

// Request Version and wait for a response.
func (m *display) Version() ([]byte, error) {
	return m.SendRecvCmd(0x00, nil)
}

// Clear actual Screen
func (m *display) ClrScreen() error {
	return m.SendCmd(0x58, nil)
}

// Run script binary. The filename path is a local path in display device
func (m *display) RunScript(filename string) error {
	m.wmux.Lock()
	defer m.wmux.Unlock()
	fmt.Println("runScript ########")
	defer fmt.Println("end runScript ########")
	data := []byte(filename)
	data = append(data, 0x00)
	if err := m.SendCmd(0x5D, data); err != nil {
		return err
	}
	var res []byte
	count := 0
	for range make([]int, 8) {
		res, _ = m.Recv()
		// fmt.Printf("////////// 1: %X\n", res)
		if len(res) > 1 {
			if res[1] == 0xFB {
				if count > 0 {
					return nil
				}
				count++
			}
		}
	}
	// fmt.Printf("////////// 2: %X\n", res)
	if len(res) > 1 {
		if res[1] == 0xFA || res[1] == 0xFB {
			log.Println("without err")
			return nil
		}
	}

	return fmt.Errorf("bad response: [% X]", res)
}

// Active buzzer in device.
// frec, is the frecuency of the signal
// time, is the duration of the signal
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

// Change Touch Reporting Style
func (m *display) ChangeTouchReporting(style int) error {
	return m.SendCmd(0x87, []byte{byte(style)})
}

// Get Touch Reporting Style
func (m *display) GetTouchReporting() ([]byte, error) {
	return m.SendRecvCmd(0x88, nil)
}

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
