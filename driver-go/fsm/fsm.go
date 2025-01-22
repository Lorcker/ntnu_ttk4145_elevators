package fsm

import (
	"Driver-go/elevio"
	"Driver-go/requests"
	"fmt"
)

type Elevator struct {
	Floor        int
	Behavior     ElevatorBehavior
	HallRequests [][]bool
	CabRequests  []bool
	Direction    elevio.MotorDirection
}

type ElevatorBehavior int

const (
	EB_Idle ElevatorBehavior = iota
	EB_DoorOpen
	EB_Moving
)

func OnRequestButtonPress(elevator *Elevator, floor int, button elevio.ButtonType) {
	fmt.Printf("Button pressed at floor %d, button type %d\n", floor, button)
	switch elevator.Behavior {
	case EB_Idle:
		if button == 2 {
			elevator.CabRequests[floor] = true
		} else {
			elevator.HallRequests[floor][button] = true
		}
		elevator.Direction, elevator.Behavior = requests.Requests_chooseDirection(elevator)

	}

}
