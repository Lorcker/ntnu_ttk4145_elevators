package elevatorio

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/message"
	"group48.ttk4145.ntnu/elevators/internal/models/request"
)

const _pollRate = 20 * time.Millisecond

var _initialized bool = false
var _numFloors int = int(elevator.NumFloors)
var _mtx sync.Mutex
var _conn net.Conn
var _local elevator.Id

func Init(addr string, local elevator.Id) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_local = local
	_initialized = true
}

func SetMotorDirection(dir elevator.MotorDirection) {
	write([4]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(btn elevator.ButtonType, floor elevator.Floor, value bool) {
	write([4]byte{2, byte(btn), byte(floor), toByte(value)})
}

func SetFloorIndicator(floor elevator.Floor) {
	write([4]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	write([4]byte{4, toByte(value), 0, 0})
}

func SetStopLamp(value bool) {
	write([4]byte{5, toByte(value), 0, 0})
}

func PollNewRequests(receiver chan<- message.RequestStateUpdate) {
	prev := make([][3]bool, _numFloors)
	for {
		time.Sleep(_pollRate)
		for f := 0; f < _numFloors; f++ {
			for b := elevator.ButtonType(0); b < 3; b++ {
				wasPressed := GetButton(b, f)
				if wasPressed != prev[f][b] && wasPressed {
					log.Printf("[elevatorio] Button %v at floor %v pressed", b, f)

					var req request.Request
					switch b {
					case elevator.HallUp:
						req = request.NewHallRequest(elevator.Floor(f), request.Up, request.Unconfirmed)
					case elevator.HallDown:
						req = request.NewHallRequest(elevator.Floor(f), request.Down, request.Unconfirmed)
					case elevator.Cab:
						req = request.NewCabRequest(elevator.Floor(f), _local, request.Unconfirmed)
					}
					receiver <- message.RequestStateUpdate{Source: _local, Request: req}
				}
				prev[f][b] = wasPressed
			}
		}
	}
}

func PollFloorSensor(receiver chan<- message.FloorSensor) {
	prev := -1
	for {
		time.Sleep(_pollRate)
		v := GetFloor()
		if v != prev && v != -1 {
			receiver <- message.FloorSensor{Floor: elevator.Floor(v)}
		}
		prev = v
	}
}

func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(receiver chan<- message.ObstructionSwitch) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetObstruction()
		if v != prev {
			receiver <- message.ObstructionSwitch{}
		}
		prev = v
	}
}

func GetButton(button elevator.ButtonType, floor int) (isPressed bool) {
	a := read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(a[1])
}

func GetFloor() int {
	a := read([4]byte{7, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

func GetStop() bool {
	a := read([4]byte{8, 0, 0, 0})
	return toBool(a[1])
}

func GetObstruction() bool {
	a := read([4]byte{9, 0, 0, 0})
	return toBool(a[1])
}

func read(in [4]byte) [4]byte {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var out [4]byte
	_, err = _conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return out
}

func write(in [4]byte) {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
