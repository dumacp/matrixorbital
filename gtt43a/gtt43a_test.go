package gtt43a


import (
	_ "fmt"
	"testing"
	_ "strings"
	"time"
)


/**/
func TestAppDemo(t *testing.T) {
	t.Log("Start Logs")
	config := &PortOptions{Port: "/dev/ttyUSB1", Baud: 19200}
        m := NewDisplay(config)

	if ok := m.Open(); !ok {
                t.Error("Not connection")
        }
	defer m.Close()

	/**
	m.RunScript("GTTProject4\\Screen2\\Screen2.bin")
	/**/
	m.UpdateBargraphValue(0, 45)
	m.UpdateLabelUTF8(0, "Cívica: 33\x00")
	m.UpdateLabelUTF8(2, "Alimentador Cívica\x00")
	/**/
	data := []byte{0xFE, 0xFA, 0x01, 0x08, 0x00, 0x09, 0x03, 0x02, 0x00, 0x32}
	sl1 := []byte{0x01, 0x05, 0x0A, 0x10, 0x015, 0x20, 0x27, 0x30, 0x37, 0x40, 0x47, 0x50, 0x57, 0x60}
	for _, v := range sl1 {
		data[len(data) -1] = v
		m.SendRecv(data, true)
		time.Sleep(time.Second * 1)
	}
	/**/
	t.Log("Stop Logs")
}

