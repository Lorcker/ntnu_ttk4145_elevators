package requests

import (
	"Driver-go/elevio"
	"Driver-go/fsm"
)

func requests_above(elevator *fsm.Elevator) bool {
	for f := elevator.Floor + 1; f < elevio.NUMFLOORS; f++ {
		if elevator.CabRequests[f] {
			return true
		}
		if elevator.HallRequests[f][0] || elevator.HallRequests[f][1] {
			return true
		}
	}
	return false
}

func requests_below(elevator *fsm.Elevator) bool {
	for f := 0; f < elevator.Floor; f++ {
		if elevator.CabRequests[f] {
			return true
		}
		if elevator.HallRequests[f][0] || elevator.HallRequests[f][1] {
			return true
		}
	}
	return false
}

func requests_here(elevator *fsm.Elevator) bool {
	if elevator.CabRequests[elevator.Floor] {
		return true
	}
	if elevator.HallRequests[elevator.Floor][0] || elevator.HallRequests[elevator.Floor][1] {
		return true
	}
	return false
}

func Requests_chooseDirection(elevator *fsm.Elevator) (elevio.MotorDirection, fsm.ElevatorBehaviour) {
	switch elevator.Direction {
	case elevio.MD_Up:
		if requests_above(elevator) {
			return elevio.MD_Up, fsm.EB_Moving
		}
		if requests_here(elevator) {
			return elevio.MD_Down, fsm.EB_DoorOpen
		}
		if requests_below(elevator) {
			return elevio.MD_Down, fsm.EB_Moving
		}
		return elevio.MD_Stop, fsm.EB_Idle
	case elevio.MD_Down:
		if requests_below(elevator) {
			return elevio.MD_Down, fsm.EB_Moving
		}
		if requests_here(elevator) {
			return elevio.MD_Up, fsm.EB_DoorOpen
		}
		if requests_above(elevator) {
			return elevio.MD_Up, fsm.EB_Moving
		}
		return elevio.MD_Stop, fsm.EB_Idle
	case elevio.MD_Stop:
		if requests_here(elevator) {
			return elevio.MD_Stop, fsm.EB_DoorOpen
		}
		if requests_above(elevator) {
			return elevio.MD_Up, fsm.EB_Moving
		}
		if requests_below(elevator) {
			return elevio.MD_Down, fsm.EB_Moving
		}
		return elevio.MD_Stop, fsm.EB_Idle
	default:
		return elevio.MD_Stop, fsm.EB_Idle
	}
}
