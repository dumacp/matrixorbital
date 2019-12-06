/**
Package to send commands and recieve response to and from gtt43a device.
**/
package gtt43a

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	_ "os"
	"sync"
	"time"
	"unicode/utf16"

	"github.com/tarm/serial"
)

type PortOptions struct {
	Port string
	Baud int
}

type Display interface {
	Open() bool
	Close() bool
	ClrScreen() int
	Text(string) int
	FontSize(int) int
	Send([]byte) int
	recv() (int, []byte)
	Recv() (int, []byte)
	SendRecv([]byte) (int, []byte)
	Echo([]byte) ([]byte, error)
	Version() ([]byte, error)
	SendRecvCmd(int, []byte) ([]byte, error)
	SendCmd(int, []byte) int
	Reset() int
	TextInsertPoint(int, int) int
	GetTextPoint() ([]byte, error)
	TextPoint(int, int) func(data string) int
	TextWindow(int, int, int, int) int
	TextColour(int, int, int) int
	PrintUTF8String(string) int
	PrintUnicode([]byte) int
	UpdateLabel(id, format int, value []byte) int
	UpdateLabelAscii(int, string) int
	UpdateLabelUTF8(int, string) int
	UpdateLabelUnicode(int, []byte) int
	UpdateBargraphValue(int, int) ([]byte, error)
	UpdateTraceValue(int, int) int
	RunScript(string) int
	LoadBitmap(int, string) ([]byte, error)
	BuzzerActive(frec, time int) int
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
	ChangeTouchReporting(style int) int
	GetTouchReporting() ([]byte, error)
	WriteScratch(addr int, data []byte) int
	ReadScratch(addr, size int) ([]byte, error)
	Listen() error
	Events() (chan *Event, error)
	StopListen()
	AnimationStartStop(id, action int) int
	AnimationSetFrame(id, state int) int
	AnimationStopAll() int
}

type display struct {
	options    *PortOptions
	status     uint32
	port       *serial.Port
	muxSend    sync.Mutex
	muxRecv    sync.Mutex
	chResp     []byte
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
	timeoutRead time.Duration = 180
)

//Create a new Display device
func NewDisplay(opt *PortOptions) Display {
	disp := &display{}
	disp.options = opt
	disp.status = CLOSED
	return disp
}

//Open device comunication channel
func (m *display) Open() bool {
	if m.status == OPENED {
		return true
	}

	config := &serial.Config{
		Name: m.options.Port,
		Baud: m.options.Baud,
		//ReadTimeout:    5 * time.Second,
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

//Listen is a go rutine that listening serial port to detect messages
//Return channel with  messages (Event struct)
func (m *display) Listen() error {
	if m.status == LISTEN {
		return fmt.Errorf("Error: already Listening display!!!")
	}

	if m.status != OPENED {
		return fmt.Errorf("Error: port serial is closed!!!")
	}
	m.chResp = make([]byte, 0)
	m.chEvent = make(chan []byte, 3)
	m.status = LISTEN
	log.Println("START listen")
	ch := make(chan []byte, 0)
	go func() {
		defer func() {
			close(m.chEvent)
			close(ch)
		}()

		buf := make([]byte, 0)
		for {
			/**/
			select {
			case <-m.stopListen:
				m.status = OPENED
				return
			default:
			}
			/**/
			//log.Printf("bajo nivel 1\n")
			n, res := m.recv()
			//log.Printf("bajo nivel 2: len: %d, [% X]\n", len(res), res)
			buf = append(buf, res...)
			if n >= 1024 {
				continue
			}
			if n >= 0 {
				for {
					if len(buf) > 0 && buf[0] == byte(0xFC) {
						if len(buf) > 4 {
							lenValue := int(binary.BigEndian.Uint16(buf[2:4]))
							if len(buf) >= lenValue+4 {
								//log.Printf("Atrapado\n")
								ch <- buf[0 : lenValue+4]
								//log.Printf("liberado\n")
								buf = buf[4+lenValue:]
								//log.Printf("bajo nivel 3: [% X]\n", buf)
								continue
							}
						}
					}
					buf = make([]byte, 0)
					break
				}
			}
			time.Sleep(20 * time.Millisecond)
		}
	}()
	go func() {
		for v := range ch {
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
					//case <-time.After(timeoutRead * time.Millisecond):
					default:
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
					//case <-time.After(timeoutRead * time.Millisecond):
					default:
						log.Printf("evento ????X [% X]\n", msg)
					}
				}
			} else if len(v) > 2 && v[0] == byte(0xFC) {
				if len(v) > 4 {
					lenValue = int(binary.BigEndian.Uint16(v[2:4]))
					if len(v) < lenValue+4 {
						continue
					}
					m.chResp = v[:lenValue+4]
					//log.Printf("respuesta low [% X]\n", m.chResp)
				}
			}
		}
		//log.Printf("EXITTTT FOR\n")
		log.Println("STOP listen")
	}()
	return nil
}

func (m *display) StopListen() {
	select {
	case m.stopListen <- 1:
	case <-time.After(1 * time.Second):
	}
}

//Primitive function to send and recieve bytes to and from display device.
//recv, flag to wait a response form device.
func (m *display) SendRecv(data []byte) (int, []byte) {
	m.muxSend.Lock()
	defer m.muxSend.Unlock()
	res := make([]byte, 0)
	m.chResp = make([]byte, 0)

	if n := m.send(data); n <= 0 {
		return -1, nil
	}

	//log.Printf("request End: [% X]\n", data)
	if m.status == LISTEN {
		tAfter1 := time.After(timeoutRead * time.Millisecond)
		tick1 := time.NewTicker(10 * time.Millisecond)
		defer tick1.Stop()
		for {
			select {
			case <-tick1.C:
				if len(m.chResp) > 0 {
					//log.Printf("response End 1: [% X]\n", m.chResp)
					res = m.chResp
					return len(res), res
				}
			case <-tAfter1:
				log.Println("timeoutRead")
				return -2, res
			}
		}
	} else {
		return m.recv()
	}

	return len(res), res
}

//Send bytes data to device. Don't wait response.
func (m *display) Send(data []byte) int {
	m.muxSend.Lock()
	defer m.muxSend.Unlock()
	if m.status == CLOSED {
		return -1
	}
	return m.send(data)
}

func (m *display) send(data []byte) int {
	if data == nil {
		return -1
	}
	n, err := m.port.Write(data)
	if err != nil || n <= 0 {
		log.Printf("Error Write: %s\n", err)
		return -1
	}
	/**
	log.Printf("request: [% X]\n", data)
	/**/
	return n
}

/**/
func (m *display) Recv() (int, []byte) {
	//m.muxRecv.Lock()
	//defer m.muxRecv.Unlock()
	m.chResp = make([]byte, 0)
	if m.status == LISTEN {
		tAfter := time.After(timeoutRead * time.Millisecond)
		tTick := time.NewTicker(10 * time.Millisecond)
		defer tTick.Stop()
		for {
			select {
			case <-tTick.C:
				if len(m.chResp) > 0 {
					res := m.chResp
					return len(res), res
				}
			case <-tAfter:
				return -2, nil
			}
		}
	}
	//	log.Printf("NOT LISTEN: %d\n", m.status)
	return m.recv()
}

/**/

//Primitive function to send and recieve bytes to and from display device.
//recv, flag to wait a response form device.
func (m *display) recv() (int, []byte) {

	if m.status == CLOSED {
		return -1, nil
	}

	res := make([]byte, 0)

	buf := make([]byte, 1024)
	n, err := m.port.Read(buf)
	if err != nil || n <= 0 {
		return -1, res
	}
	res = append(res, buf[:n]...)
	//log.Printf("parcial response 1: [% X]\n", res)
	return n, res
}

//Send a command to display device
//cmd, id for the command
//wait response
func (m *display) SendRecvCmd(cmd int, data []byte) ([]byte, error) {
	m.muxSend.Lock()
	defer m.muxSend.Unlock()
	dat1 := []byte{0xFE, byte(cmd)}
	if data != nil {
		dat1 = append(dat1, data...)
	}
	res := make([]byte, 0)
	n := -1

	//n, res := m.SendRecv(dat1)
	if n = m.send(dat1); n <= 0 {
		return nil, errors.New("empty response")
	}

	//log.Printf("request End: [% X]\n", dat1)
	if m.status == LISTEN {
		after := time.After(timeoutRead * time.Millisecond)
		tick := time.NewTicker(10 * time.Millisecond)
		defer tick.Stop()
		m.chResp = make([]byte, 0)
	for_src:
		for {
			select {
			case <-tick.C:
				if len(m.chResp) > 2 {
					//log.Printf("response End 1: [% X]\n", m.chResp)
					res = m.chResp
					if res[1] != byte(cmd) {
						log.Printf("Other response: [% X]\n", m.chResp)
						m.chResp = make([]byte, 0)
						continue
					}
					break for_src
				}
			case <-after:
				log.Println("timeoutRead")
				return res, errors.New("timeoutRead")
			}
		}
	} else {
		time.Sleep(timeoutRead * time.Millisecond)
		n, res = m.recv()
	}
	//log.Printf("SendRecvCmd response: [% X]", res)

	switch {
	case n <= 0:
		return nil, errors.New("empty response")
	case res == nil:
		return nil, errors.New("empty response")
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
func (m *display) SendCmd(cmd int, data []byte) int {
	dat1 := []byte{0xFE, byte(cmd)}
	if data != nil {
		dat1 = append(dat1, data...)
	}

	n := m.Send(dat1)

	if n <= 0 {
		return -1
	}
	return n
}

//Send echo data and to wait for a response.
func (m *display) Echo(data []byte) ([]byte, error) {
	return m.SendRecvCmd(0xFF, data)
}

//Send reset command to display device
func (m *display) Reset() int {
	n := m.Send([]byte{0xFE, 0x01})
	return n
}

//Request Version and wait for a response.
func (m *display) Version() ([]byte, error) {
	return m.SendRecvCmd(0x00, nil)
}

//Print text data in actual (x,y) point in display area
func (m *display) Text(data string) int {
	n := m.Send([]byte(data))
	return n
}

//Set font Size
func (m *display) FontSize(size int) int {
	return m.SendCmd(0x33, []byte{byte(size)})
}

//Set (x,y) point in display area. The next print and draw command will be set in this point.
func (m *display) TextInsertPoint(x, y int) int {
	data := make([]byte, 0)
	xb := make([]byte, 2)
	yb := make([]byte, 2)
	binary.BigEndian.PutUint16(xb, uint16(x))
	binary.BigEndian.PutUint16(yb, uint16(x))
	data = append(data, xb...)
	data = append(data, yb...)
	n := m.SendCmd(0x79, data)
	return n
}

//Get actual (x,y) point
func (m *display) GetTextPoint() ([]byte, error) {
	return m.SendRecvCmd(0x7A, nil)
}

//Clear actual Screen
func (m *display) ClrScreen() int {
	return m.SendCmd(0x58, nil)
}

//Print data text in this (x,y) point
func (m *display) TextPoint(x, y int) func(data string) int {
	return func(data string) int {
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
	data := make([]byte, 0)
	xb := make([]byte, 2)
	yb := make([]byte, 2)
	widthb := make([]byte, 2)
	heightb := make([]byte, 2)
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
func (m *display) UpdateLabel(id, format int, value []byte) int {
	data := []byte{byte(id), byte(format)}
	data = append(data, []byte(value)...)
	data = append(data, 0x00)

	return m.SendCmd(0x11, data)
}

//Update the text data (in string) in the label ID with Ascii Codification
func (m *display) UpdateLabelAscii(id int, value string) int {
	return m.UpdateLabel(id, 0, []byte(value))
}

//Update the text data (in string) in the label ID with UTF-8 Codification
func (m *display) UpdateLabelUTF8(id int, value string) int {
	return m.UpdateLabel(id, 2, []byte(value))
}

//Update the text data (in bytes, 2 bytes for character) in the label ID with Unicode Codification
func (m *display) UpdateLabelUnicode(id int, value []byte) int {
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
func (m *display) UpdateTraceValue(id, value int) int {
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
	n1 := m.SendCmd(0x5D, data)
	//fmt.Printf("salida Run: %v\n", res)
	//time.Sleep(1 * time.Second)
	m.Recv()
	return n1
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
func (m *display) BuzzerActive(frec, time int) int {

	data := make([]byte, 0)
	frecb := make([]byte, 2)
	timeb := make([]byte, 2)
	binary.BigEndian.PutUint16(frecb, uint16(frec))
	binary.BigEndian.PutUint16(timeb, uint16(time))
	data = append(data, frecb...)
	data = append(data, timeb...)
	n3 := m.SendCmd(0xBB, data)
	return n3
}

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
		_, res := m.SendRecv(data)
		if len(res) < 3 {
			return fmt.Errorf("Error in response: [% X]", res)
		}
		if res[len(res)-1] != byte(0xFE) {
			return fmt.Errorf("Error in request U16, status code: [%X]", res[2])
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
		_, res := m.SendRecv(data)
		if len(res) < 3 {
			return fmt.Errorf("Error in response: [% X]", res)
		}
		if res[len(res)-1] != byte(0xFE) {
			return fmt.Errorf("Error in request S16, status code: [%X]", res[2])
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
		_, res := m.SendRecv(data)
		if len(res) < 3 {
			return fmt.Errorf("Error in response: [% X]", res)
		}
		if res[len(res)-1] != byte(0xFE) {
			return fmt.Errorf("Error in request U8, status code: [%X]", res[2])
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
		_, res := m.SendRecv(data)
		if len(res) < 3 {
			return fmt.Errorf("Error in response: [% X]", res)
		}
		if res[len(res)-1] != byte(0xFE) {
			return fmt.Errorf("Error in request, status code: [%X]", res[2])
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
		_, res := m.SendRecv(data)
		if len(res) < 3 {
			return nil, fmt.Errorf("Error in response: [% X]", res)
		}
		if res[len(res)-3] != byte(0xFE) {
			return nil, fmt.Errorf("Error in request U16, status code: [%X]", res[2])
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
		_, res := m.SendRecv(data)
		if len(res) < 3 {
			return nil, fmt.Errorf("Error in response: [% X]", res)
		}
		if res[len(res)-3] != byte(0xFE) {
			return nil, fmt.Errorf("Error in request S16, status code: [%X]", res[2])
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
		_, res := m.SendRecv(data)
		if len(res) < 3 {
			return 0x00, fmt.Errorf("Error in response: [% X]", res)
		}
		if res[len(res)-2] != byte(0xFE) {
			return byte(0x00), fmt.Errorf("Error in request U8, status code: [%X]", res[2])
		}
		return res[len(res)-1], nil
	}
}

//TOUCH

//Change Touch Reporting Style
func (m *display) ChangeTouchReporting(style int) int {
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

func (m *display) WriteScratch(addr int, data []byte) int {
	dat1 := make([]byte, 0)
	addrb := make([]byte, 2)
	sizeb := make([]byte, 2)
	binary.BigEndian.PutUint16(addrb, uint16(addr))
	binary.BigEndian.PutUint16(sizeb, uint16(len(data)))
	dat1 = append(dat1, addrb...)
	dat1 = append(dat1, sizeb...)
	dat1 = append(dat1, data...)
	n := m.SendCmd(0xCC, dat1)

	return n
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

func (m *display) AnimationStartStop(id, action int) int {
	n := m.SendCmd(0xC2, []byte{byte(id), byte(action)})
	return n
}

func (m *display) AnimationSetFrame(id, state int) int {
	n := m.SendCmd(0xC3, []byte{byte(id), byte(state)})
	return n
}

func (m *display) AnimationStopAll() int {
	n := m.SendCmd(0xC6, nil)
	return n
}
