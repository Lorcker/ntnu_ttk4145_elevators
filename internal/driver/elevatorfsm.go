package driver

import (
	"log"

	"group48.ttk4145.ntnu/elevators/internal/elevatorio"
	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
)

const EverybodyGoesOn bool = false

type resolvedRequests func(btn elevator.ButtonType, floor elevator.Floor)

// handleOrderEvent updates the elevator state based on new orders
func handleOrderEvent(state *elevator.State, orders elevator.Order, recieverDoorTimer chan<- bool, rr resolvedRequests) {
	switch state.Behavior {
	case elevator.Idle:
		requestChooseDirection(state, orders, recieverDoorTimer) // Updates the behaviour and direction
		if state.Behavior == elevator.DoorOpen {
			requestClearAtCurrentFloor(*state, &orders, rr) // Clears orders when order comes in at the current floor.
		} else if state.Behavior == elevator.Moving {
			elevatorio.SetMotorDirection(state.Direction)
		}

	case elevator.DoorOpen:
		if requestShouldClearImmediatly(*state, orders) {
			recieverDoorTimer <- true
			requestClearAtCurrentFloor(*state, &orders, rr)
		}
	}
}

// handleFloorsensorEvent updates the elevator state when arriving at a new floor
func handleFloorsensorEvent(state *elevator.State, orders elevator.Order, recieverDoorTimer chan<- bool, rr resolvedRequests, floor elevator.Floor) {
	state.Floor = floor
	elevatorio.SetFloorIndicator(floor)
	if state.Behavior == elevator.Moving && requestShouldStop(*state, orders) {
		elevatorio.SetMotorDirection((0))
		elevatorio.SetDoorOpenLamp(true)
		recieverDoorTimer <- true
		requestClearAtCurrentFloor(*state, &orders, rr)
	}
}

// When timer is done, close the door, and go in desired direction.
func handleDoorTimerEvent(state *elevator.State, orders elevator.Order, recieverDoorTimer chan<- bool, rr resolvedRequests) {
	if state.Behavior == elevator.DoorOpen {
		requestChooseDirection(state, orders, recieverDoorTimer)
		if state.Behavior == elevator.DoorOpen {
			recieverDoorTimer <- true
			requestClearAtCurrentFloor(*state, &orders, rr)
		} else {
			elevatorio.SetDoorOpenLamp(false)
			elevatorio.SetMotorDirection(state.Direction)
		}
	}
}

func openDoor(state *elevator.State) {
	log.Printf("[elevatorfsm] Door open\n")
	elevatorio.SetDoorOpenLamp(true)
	state.Behavior = elevator.DoorOpen
}

// requestChooseDirection updates the elevator direction and behaviour based on the current orders. Inspired by the given C-code.
func requestChooseDirection(e *elevator.State, orders elevator.Order, recieverDoorTimer chan<- bool) {
	switch e.Direction {
	case elevator.Up:
		if requestAbove(*e, orders) {
			e.Direction = elevator.Up
			e.Behavior = elevator.Moving
		} else if requestHere(*e, orders) {
			e.Direction = elevator.Stop
			recieverDoorTimer <- true

		} else if requestBelow(*e, orders) {
			e.Direction = elevator.Down
			e.Behavior = elevator.Moving
		} else {
			e.Direction = elevator.Stop
			e.Behavior = elevator.Idle
		}

	case elevator.Down:
		if requestBelow(*e, orders) {
			e.Direction = elevator.Down
			e.Behavior = elevator.Moving
		} else if requestHere(*e, orders) {
			e.Direction = elevator.Stop
			recieverDoorTimer <- true
		} else if requestAbove(*e, orders) {
			e.Direction = elevator.Up
			e.Behavior = elevator.Moving
		} else {
			e.Direction = elevator.Stop
			e.Behavior = elevator.Idle
		}

	case elevator.Stop:
		if requestHere(*e, orders) {
			e.Direction = elevator.Stop
			recieverDoorTimer <- true
		} else if requestAbove(*e, orders) {
			e.Direction = elevator.Up
			e.Behavior = elevator.Moving
		} else if requestBelow(*e, orders) {
			e.Direction = elevator.Down
			e.Behavior = elevator.Moving
		} else {
			e.Direction = elevator.Stop
			e.Behavior = elevator.Idle
		}
	}
}

func requestAbove(e elevator.State, orders elevator.Order) bool {
	if e.Floor >= elevator.NumFloors-1 {
		return false
	} //Already at top floor

	for i := (e.Floor + 1); i < elevator.NumFloors; i++ {
		for j := range orders[e.Floor] {
			if orders[i][j] {
				return true
			}
		}
	}
	return false
}

func requestHere(e elevator.State, orders elevator.Order) bool {
	for _, order := range orders[e.Floor] {
		if order {
			return true
		}
	}
	return false
}

func requestBelow(e elevator.State, orders elevator.Order) bool {
	if e.Floor == 0 {
		return false
	} // Already at bottom floor
	for i := e.Floor - 1; i >= 0; i-- {
		for j := range len(orders[e.Floor]) {
			if orders[i][j] {
				return true
			}
		}
	}
	return false

}

// requestClearAtCurrentFloor clears the orders that are excecuted
func requestClearAtCurrentFloor(e elevator.State, orders *elevator.Order, rr resolvedRequests) {
	// If EverybodyGoesOn is set to true, then all orders clear if the elevator stops at a floor.
	if EverybodyGoesOn {
		for j := range len((*orders)[e.Floor]) {
			(*orders)[e.Floor][j] = false
			rr(elevator.HallUp, e.Floor)
			rr(elevator.HallDown, e.Floor)
			rr(elevator.Cab, e.Floor)
		}
	} else {
		(*orders)[e.Floor][elevator.Cab] = false //Cab orders are always cleared.
		rr(elevator.Cab, e.Floor)

		switch e.Direction {
		case elevator.Up:
			if !requestAbove(e, (*orders)) && !(*orders)[e.Floor][elevator.HallUp] { // Hall Down request is only cleared if it is not going further up
				(*orders)[e.Floor][elevator.HallDown] = false
				rr(elevator.HallDown, e.Floor)
			}
			(*orders)[e.Floor][elevator.HallUp] = false
			rr(elevator.HallUp, e.Floor)

		case elevator.Down:
			if !requestBelow(e, (*orders)) && !(*orders)[e.Floor][elevator.HallDown] { // Hall Up request is only cleared if it is not going further down
				(*orders)[e.Floor][elevator.HallUp] = false
				rr(elevator.HallUp, e.Floor)
			}
			(*orders)[e.Floor][elevator.HallDown] = false
			rr(elevator.HallDown, e.Floor)

		case elevator.Stop:
			fallthrough
		default:
			(*orders)[e.Floor][elevator.HallDown] = false
			(*orders)[e.Floor][elevator.HallUp] = false
			rr(elevator.HallDown, e.Floor)
			rr(elevator.HallUp, e.Floor)
		}

	}

}

// requestShouldStop returns true if elevator should stop on that floor
func requestShouldStop(e elevator.State, orders elevator.Order) bool {
	switch e.Direction {
	case elevator.Down:
		return orders[e.Floor][elevator.HallDown] || orders[e.Floor][elevator.Cab] || !requestBelow(e, orders)
	case elevator.Up:
		return orders[e.Floor][elevator.HallUp] || orders[e.Floor][elevator.Cab] || !requestAbove(e, orders)
	default:
		return true
	}
}

// requestShouldClearImmediatly returns true if a order that comes in should be handled immediatly
func requestShouldClearImmediatly(e elevator.State, orders elevator.Order) bool {
	if EverybodyGoesOn {
		for i := range len(orders[e.Floor]) {
			if orders[e.Floor][i] {
				return true
			}
		}
		return false
	} else {
		switch e.Direction {
		case elevator.Up:
			return orders[e.Floor][elevator.HallUp]
		case elevator.Down:
			return orders[e.Floor][elevator.HallDown]
		case elevator.Stop:
			return orders[e.Floor][elevator.Cab]
		default:
			return false
		}
	}
}
