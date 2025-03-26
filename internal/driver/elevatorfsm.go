package driver

import (
	"fmt"
	"log"

	"group48.ttk4145.ntnu/elevators/internal/elevatorio"
	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
)

var EverybodyGoesOn bool = false

type resolvedRequests func(btn elevator.ButtonType, floor elevator.Floor)

func handleOrderEvent(state *elevator.State, orders elevator.Order, recieverDoorTimer chan<- bool, rr resolvedRequests) {
	switch state.Behavior {
	case elevator.Idle:
		requestChooseDirection(state, orders, recieverDoorTimer)
		if state.Behavior == elevator.DoorOpen {
			requestClearAtCurrentFloor(*state, &orders, rr)
		} else if state.Behavior == elevator.Moving {
			elevatorio.SetMotorDirection(state.Direction)
		}

	case elevator.DoorOpen:
		if RequestShouldClearImmediatly(*state, orders) {
			recieverDoorTimer <- true
			requestClearAtCurrentFloor(*state, &orders, rr)
		}
	}
}

func handleFloorsensorEvent(state *elevator.State, orders elevator.Order, floor elevator.Floor, recieverDoorTimer chan<- bool, rr resolvedRequests) {
	state.Floor = floor
	elevatorio.SetFloorIndicator(floor)

	if state.Behavior == elevator.Moving && RequestShouldStop(*state, orders) {
		elevatorio.SetMotorDirection((0))
		elevatorio.SetDoorOpenLamp(true)
		requestClearAtCurrentFloor(*state, &orders, rr)
		recieverDoorTimer <- true
	}
}

// When timer is done, close the door, and go in desired direction.
func handleDoorTimerEvent(state *elevator.State, orders elevator.Order, recieverDoorTimer chan<- bool, rr resolvedRequests) {
	if (*state).Behavior == elevator.DoorOpen {
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

func emergencyStop(state *elevator.State) {
	log.Printf("[elevatorfsm] Stop button not implemented :(\n")
}

// Little bit inspired by the given C-code :)
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

func requestClearAtCurrentFloor(e elevator.State, orders *elevator.Order, rr resolvedRequests) {
	//Definisjon. True: Alle ordre skal fjernes fra etasjen (alle går på). False: Bare de i samme retning.
	if EverybodyGoesOn {
		for j := range len((*orders)[e.Floor]) {
			(*orders)[e.Floor][j] = false
			rr(elevator.HallUp, e.Floor)
			rr(elevator.HallDown, e.Floor)
			rr(elevator.Cab, e.Floor)
		}
	} else {
		(*orders)[e.Floor][elevator.Cab] = false
		rr(elevator.Cab, e.Floor)

		switch e.Direction {
		case elevator.Up:
			if !requestAbove(e, (*orders)) && !(*orders)[e.Floor][elevator.HallUp] {
				(*orders)[e.Floor][elevator.HallDown] = false
				rr(elevator.HallDown, e.Floor)
			}
			(*orders)[e.Floor][elevator.HallUp] = false
			rr(elevator.HallUp, e.Floor)

		case elevator.Down:
			if !requestBelow(e, (*orders)) && !(*orders)[e.Floor][elevator.HallDown] {
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

func RequestShouldStop(e elevator.State, orders elevator.Order) bool {
	switch e.Direction {
	case elevator.Down:
		return orders[e.Floor][elevator.HallDown] || orders[e.Floor][elevator.Cab] || !requestBelow(e, orders)
	case elevator.Up:
		return orders[e.Floor][elevator.HallUp] || orders[e.Floor][elevator.Cab] || !requestAbove(e, orders)
	default:
		return true
	}
}

// Decision: Have to decide if everyone will get in the state, even tho they might be going in the opposite direction.

func RequestShouldClearImmediatly(e elevator.State, orders elevator.Order) bool {
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

// Debug functions.
func printOrders(orders elevator.Order) {
	// Iterate through the outer slice (rows)
	log.Printf("Floor\t Up\t Down\t Cab\n")
	for i := range len(orders) {
		// Iterate through the inner slice (columns) at each row
		fmt.Printf("%d\t", i)
		for j := range len(orders[i]) {
			// Print the Order information
			fmt.Printf("%t\t ", orders[i][j])
		}
		fmt.Printf("\n\n")
	}
}

func printElevatorState(state elevator.State) {
	log.Printf("[elevatorfsm]\n\nFloor: %d\n", state.Floor)
	log.Printf("Behavior: %d\n", state.Behavior)
	log.Printf("Direction: %d\n\n", state.Direction)
}
