package gtt43a

import (
	"fmt"
	"log"
	_ "strings"
	"testing"
	"time"
)

/**/
func TestAppDemo(t *testing.T) {
	t.Log("Start Logs")
	config := &PortOptions{Port: "/dev/ttyUSB0", Baud: 115200}
	m := NewDisplay(config)

	if ok := m.Open(); !ok {
		t.Error("Not connection")
	}
	defer m.Close()

	/**/
	m.RunScript("GTTProject4\\Screen2\\Screen2.bin")
	time.Sleep(3 * time.Second)
	/**/
	m.UpdateBargraphValue(0, 45)
	m.UpdateLabelUTF8(0, "Cívica: 33\x00")
	m.UpdateLabelUTF8(2, "Alimentador Cívica\x00")
	m.UpdateLabelUTF8(3, "Alimentador Subruta\x00")
	/**/

	resp, err := m.SendRecv([]byte{0xFE, 0x88})
	if err != nil {
		log.Println(err)
	} else {
		if len(resp) > 0 {
			log.Printf("response: [% X]\n", resp)
		}
	}
	resp, err = m.SendRecv([]byte{0xFE, 0x87, 0x03})
	if err != nil {
		log.Println(err)
	} else {
		if len(resp) > 0 {
			log.Printf("response: [% X]\n", resp)
		}
	}
	resp, err = m.SendRecv([]byte{0xFE, 0x88})
	if err != nil {
		log.Println(err)
	} else {
		if len(resp) > 0 {
			log.Printf("response: [% X]\n", resp)
		}
	}

	data := []byte{0xFE, 0xFA, 0x01, 0x08, 0x00, 0x09, 0x03, 0x02, 0x00, 0x32}
	sl1 := []byte{0x01, 0x05, 0x0A, 0x10, 0x015, 0x20, 0x27, 0x30, 0x37, 0x40, 0x47, 0x50, 0x57, 0x60}

	for _, v := range sl1 {
		data[len(data)-1] = v
		m.Send(data)
		time.Sleep(time.Millisecond * 100)
	}

	chRead := make(chan []byte)
	go func() {
		defer close(chRead)
		for {
			buf, err := m.Recv()
			if err == nil && len(buf) > 0 {
				chRead <- buf
			}
		}
	}()

	go func() {
		for i := 0; i < 33; i++ {
			m.UpdateLabelUTF8(2, fmt.Sprintf("Alimentador Cívica: %d", i))
			m.UpdateLabelUTF8(3, fmt.Sprintf("Alimentador Subruta: %d", i))
			time.Sleep(time.Millisecond * 100)
		}
	}()

break_for:
	for {
		select {
		case v := <-chRead:
			log.Printf("read serial port: [% X]\n", v)
		case <-time.After(3 * time.Second):
			break break_for
		}
	}
	/**/
	t.Log("Stop Logs")
}
