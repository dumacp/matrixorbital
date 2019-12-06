/**
This file contain the functions to manage event messages
/**/
package gtt43a

import (
	"encoding/binary"
	"fmt"
	"log"
	"time"
	_ "bytes"
)


type EventType int
const (
	GTT25BaseObjectOnPropertyChange EventType = iota
	GTT25VisualObjectOnKey
	ButtonClick
	RegionTouch
)

type Event struct {
	Type	EventType
	ObjId	uint16
	Value	[]byte
}


//ListenEvents is a go rutine that listening serial port to detect event messages
//Return channel with event messages (Event struct)
func (m *display) Events() (chan *Event, error) {
	if m.status != LISTEN {
		return nil, fmt.Errorf("Error, display don't be listenning. Execute Listen() before wait events")
	}
	mc := make(chan *Event)
	go func() {
		defer close(mc)
		for v := range m.chEvent {
			log.Printf("read Event: [% X]\n", v)
			objId := uint16(0)
			data := make([]byte,0)
			var evnT EventType
			switch {
			case byte(0xEB) == v[0]:
				evnt := binary.LittleEndian.Uint16(v[1:3])
				switch evnt {
				case 0x01:
					evnT = GTT25BaseObjectOnPropertyChange
				case 0x02:
					evnT = GTT25VisualObjectOnKey
				case 0x15:
					evnT = ButtonClick
				default:
					continue
				}
				objId = binary.BigEndian.Uint16(v[3:5])
				data = v[5:]
			case  byte(0x87) == v[0]:
				switch  {
				case len(v) == 3:
					evnT = RegionTouch
				default:
					continue
				}
				objId = uint16(v[2])
				data = v[1:2]
			default:
				continue
			}

			go func() {
				event := &Event{
					evnT,
					objId,
					data,
				}
				select {
				case mc <- event:
				case <-time.After(timeoutRead * time.Millisecond):
				}
			}()
		}
	}()
	return mc, nil
}



