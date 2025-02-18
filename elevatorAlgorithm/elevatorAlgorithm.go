package elevatorAlgorithm

import (
	"group48.ttk4145.ntnu/elevators/elevatorfsm"
	"group48.ttk4145.ntnu/elevators/elevatorio"
)

func requestsAbove(e elevatorfsm.Elevator) bool {
	for i := e.Floor + 1; i < len(e.Requests); i++ {
		for _, req := range e.Requests[i] {
			if req {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e elevatorfsm.Elevator) bool {
	for i := 0; i < e.Floor; i++ {
		for _, req := range e.Requests[i] {
			if req {
				return true
			}
		}
	}
	return false
}

func anyRequests(e elevatorfsm.Elevator) bool {
	for _, floorRequests := range e.Requests {
		for _, req := range floorRequests {
			if req {
				return true
			}
		}
	}
	return false
}

func anyRequestsAtFloor(e elevatorfsm.Elevator) bool {
	for _, req := range e.Requests[e.Floor] {
		if req {
			return true
		}
	}

	return false
}

func shouldStop(e elevatorfsm.Elevator) bool {
	switch e.Direction {
	case elevatorio.MD_Up:
		return e.Requests[e.Floor][elevatorio.BT_HallUp] ||
			e.Requests[e.Floor][elevatorio.BT_Cab] ||
			!requestsAbove(e) ||
			e.Floor == 0 ||
			e.Floor == len(e.Requests)-1
	case elevatorio.MD_Down:
		return e.Requests[e.Floor][elevatorio.BT_HallDown] ||
			e.Requests[e.Floor][elevatorio.BT_Cab] ||
			!requestsBelow(e) ||
			e.Floor == 0 ||
			e.Floor == len(e.Requests)-1
	case elevatorio.MD_Stop:
		return true
	}
	return false
}

func chooseDirection(e elevatorfsm.Elevator) elevatorio.MotorDirection {
	switch e.Direction {
	case elevatorio.MD_Up:
		if requestsAbove(e) {
			return elevatorio.MD_Up
		} else if anyRequestsAtFloor(e) {
			return elevatorio.MD_Stop
		} else if requestsBelow(e) {
			return elevatorio.MD_Down
		} else {
			return elevatorio.MD_Stop
		}
	case elevatorio.MD_Down, elevatorio.MD_Stop:
		if requestsBelow(e) {
			return elevatorio.MD_Down
		} else if anyRequestsAtFloor(e) {
			return elevatorio.MD_Stop
		} else if requestsAbove(e) {
			return elevatorio.MD_Up
		} else {
			return elevatorio.MD_Stop
		}
	}
	return elevatorio.MD_Stop
}

func clearReqsAtFloor(e elevatorfsm.Elevator, onClearedRequest func(elevatorio.ButtonType)) elevatorfsm.Elevator {
	e2 := e

	clear := func(c elevatorio.ButtonType) {
		if e2.Requests[e2.Floor][c] {
			if onClearedRequest != nil {
				onClearedRequest(c)
			}
			e2.Requests[e2.Floor][c] = false
		}
	}

	clearRequestType := "all" // or "inDirn", depending on your logic

	switch clearRequestType {
	case "all":
		for c := range e2.Requests[0] {
			clear(elevatorio.ButtonType(c))
		}
	case "inDirn":
		clear(elevatorio.BT_Cab)

		switch e.Direction {
		case elevatorio.MD_Up:
			if e2.Requests[e2.Floor][elevatorio.BT_HallUp] {
				clear(elevatorio.BT_HallUp)
			} else if !requestsAbove(e2) {
				clear(elevatorio.BT_HallDown)
			}
		case elevatorio.MD_Down:
			if e2.Requests[e2.Floor][elevatorio.BT_HallDown] {
				clear(elevatorio.BT_HallDown)
			} else if !requestsBelow(e2) {
				clear(elevatorio.BT_HallUp)
			}
		case elevatorio.MD_Stop:
			clear(elevatorio.BT_HallUp)
			clear(elevatorio.BT_HallDown)
		}
	}

	return e2
}
