package glk19264a


import (
	"fmt"
	"testing"
	"strings"
	_ "time"
)

/**
func TestTextRow(t *testing.T) {
	t.Log("Start Logs")
	config := &PortOptions{Port: "/dev/ttyUSB2", Baud: 19200}
        m := NewDisplay(config)

	if ok := m.Open(); !ok {
                t.Error("Not connection")
        }
	defer m.Close()

	m.ClrScreen()
	m.Font(2)
	m.SetTextWindow(0)

	row0 := m.TextColRow(0,0)

	if n := row0("PRUEBA DESDE GO Row 1"); n <=0 {
		t.Errorf("error: %s", n)
	}

	rowX := m.TextColRow(0,3)

	m.Font(1)

	if n := rowX("PRUEBA DESDE GO Row 3"); n <=0 {
		t.Errorf("error: %s", n)
	}
	t.Log("Stop Logs")
}

/**
func TestBitmapUpload(t *testing.T) {
	t.Log("Start Logs")
	config := &PortOptions{Port: "/dev/ttyUSB1", Baud: 19200}
        m := NewDisplay(config)

	if ok := m.Open(); !ok {
                t.Error("Not connection")
        }
        defer m.Close()

	//m.ClrScreen()

	//n, err := m.BitmapUpload(2, "/tmp/smile.bmp")
	n, err := m.BitmapUpload(2, "/home/duma/Downloads/check3.bmp")
	if n <= 0 || err != nil {
		t.Error("ERROR: %v, %v", n, err)
	}

	t.Log("Stop Logs")
}
/**
func TestBitmapDrawFile(t *testing.T) {
	t.Log("Start Logs")
	config := &PortOptions{Port: "/dev/ttyUSB1", Baud: 19200}
        m := NewDisplay(config)

	if ok := m.Open(); !ok {
                t.Error("Not connection")
        }
        defer m.Close()

	m.ClrScreen()

	n, err := m.BitmapDrawFile(20,5,"/home/duma/Downloads/check3.bmp")
	if n <= 0 || err != nil {
		t.Error("ERROR: %v, %v", n, err)
	}

	t.Log("Stop Logs")
}
/**

func TestBitmapDraw(t *testing.T) {
	t.Log("Start Logs")
	config := &PortOptions{Port: "/dev/ttyUSB1", Baud: 19200}
        m := NewDisplay(config)

	if ok := m.Open(); !ok {
                t.Error("Not connection")
        }
        defer m.Close()

	//m.ClrScreen()

	n := m.BitmapDraw(20, 10, 2)
	if n <= 0 {
		t.Error("ERROR: %v", n)
	}

	t.Log("Stop Logs")

}
/**

func TestBitmapDrawData(t *testing.T) {
	t.Log("Start Logs")
	config := &PortOptions{Port: "/dev/ttyUSB1", Baud: 19200}
        m := NewDisplay(config)

	if ok := m.Open(); !ok {
                t.Error("Not connection")
        }
        defer m.Close()

	m.ClrScreen()



	data1 := []byte{0x10, 0x0F,0xFF, 0xFF, 0x80, 0x03, 0x80, 0x07, 0x80, 0x0D, 0x80, 0x19, 0x80, 0x31, 0x80, 0x61, 0x80, 0xC1, 0x81, 0x81, 0xE3, 0x01, 0xB6, 0x01, 0x9C, 0x01, 0x8C, 0x01, 0x80, 0x01, 0xFF, 0xFF}
	
//	data1 := append([]byte{ 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1}
//	data1 := append([]byte{ 1 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1}
//	data1 := append([]byte{ 1 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1}
//	data1 := append([]byte{ 1 0 0 0 0 0 0 0 0 0 0 0 1 1 0 1}
//	data1 := append([]byte{ 1 0 0 0 0 0 0 0 0 0 0 1 1 0 0 1}
//	data1 := append([]byte{ 1 0 0 0 0 0 0 0 0 0 1 1 0 0 0 1}
//	data1 := append([]byte{ 1 0 0 0 0 0 0 0 0 1 1 0 0 0 0 1}
//	data1 := append([]byte{ 1 0 0 0 0 0 0 0 1 1 0 0 0 0 0 1}
//	data1 := append([]byte{ 1 0 0 0 0 0 0 1 1 0 0 0 0 0 0 1}
//	data1 := append([]byte{ 1 1 1 0 0 0 1 1 0 0 0 0 0 0 0 1}
//	data1 := append([]byte{ 1 0 1 1 0 1 1 0 0 0 0 0 0 0 0 1}
//	data1 := append([]byte{ 1 0 0 1 1 1 0 0 0 0 0 0 0 0 0 1}
//	data1 := append([]byte{ 1 0 0 0 1 1 0 0 0 0 0 0 0 0 0 1}
//	data1 := append([]byte{ 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1}
//	data1 := append([]byte{ 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1}

	n := m.BitmapDrawData(0, 0, data1)
	if n <= 0 {
		t.Error("ERROR: %v", n)
	}

	t.Log("Stop Logs")

}
/**

func TestBuzzerActive(t *testing.T) {
	t.Log("Start Logs")
	config := &PortOptions{Port: "/dev/ttyUSB1", Baud: 19200}
        m := NewDisplay(config)

	if ok := m.Open(); !ok {
                t.Error("Not connection")
        }
        defer m.Close()


	var n int
	for i:=0 ; i<3 ; i++ {
		n = m.BuzzerActive(700, 150)
		time.Sleep(time.Millisecond * 300)
	}
	if n <= 0 {
		t.Error("ERROR: %v", n)
	}

	t.Log("Stop Logs")
}
/**

func TestBackLigthOffOn(t *testing.T) {
	t.Log("Start Logs")
	config := &PortOptions{Port: "/dev/ttyUSB2", Baud: 19200}
        m := NewDisplay(config)

	if ok := m.Open(); !ok {
                t.Error("Not connection")
        }
        defer m.Close()


	var n int
	for i:=0 ; i<5 ; i++ {
		n = m.KeyPadOff()
		n = m.BuzzerActive(1000, 150)
		time.Sleep(time.Millisecond * 100)
		n = m.KeyPadON(128)
		time.Sleep(time.Millisecond * 100)
	}
	if n <= 0 {
		t.Error("ERROR: %v", n)
	}

	t.Log("Stop Logs")
}
/**

func TestBackLigthOffOn(t *testing.T) {
	t.Log("Start Logs")
	config := &PortOptions{Port: "/dev/ttyUSB1", Baud: 19200}
        m := NewDisplay(config)

	if ok := m.Open(); !ok {
                t.Error("Not connection")
        }
        defer m.Close()


	var n int
	for i:=0 ; i<3 ; i++ {
		for j:=4 ; j>=0 ; j-- {
			n = m.Led(i,j)
			time.Sleep(time.Millisecond * 1000)
		}
	}
	if n <= 0 {
		t.Error("ERROR: %v", n)
	}

	t.Log("Stop Logs")
}
/**/

/**/
func TestTextWindow(t *testing.T) {
	t.Log("Start Logs")
	config := &PortOptions{Port: "/dev/ttyUSB2", Baud: 19200}
        m := NewDisplay(config)

	if ok := m.Open(); !ok {
                t.Error("Not connection")
        }
	defer m.Close()

	m.ClrScreen()
	/**/
	id := 1
	x1 := 1
	y1 := 32
	x2 := 189
	y2 := 62
	font := 1
	charSpace := 1
	lineSpace := 1
	scroll := 62

	m.InitTextWindow(id, x1, y1, x2, y2, font, charSpace, lineSpace, scroll)
	m.SetTextWindow(id)

	slice := []int{0, 1, 2, 3, 4, 5, 6, 7, 8}
	s1 := make([]string,0)
	for _, v := range slice {
		s1 = append(s1,fmt.Sprintf("hola mundo %d!!!", v))
	}
	m.Text(strings.Join(s1, "\n"))
	/**/

	m.Rectangle(1, 0, 30, 190, 63)
	t.Log("Stop Logs")
}

