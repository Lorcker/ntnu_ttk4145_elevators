package driver

import (
	"log"

	"group48.ttk4145.ntnu/elevators/internal/elevatorio"
	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
)

const EverybodyGoesOn bool = false

type resolvedRequests func(btn elevator.ButtonType, floor elevator.Floor)

// fsmHandleOrderEvent updates the elevator state based on new orders
func fsmHandleOrderEvent(state *elevator.State, orders elevator.Order, recieverDoorTimer chan<- bool, rr resolvedRequests) {
	switch state.Behavior {
	case elevator.Idle:
		fsmChooseDirection(state, orders, recieverDoorTimer) // Updates the behaviour and direction
		if state.Behavior == elevator.DoorOpen {
			orderClearAtCurrentFloor(*state, &orders, rr) // Clears orders when order comes in at the current floor.
		} else if state.Behavior == elevator.Moving {
			elevatorio.SetMotorDirection(state.Direction)
		}

	case elevator.DoorOpen:
		if orderShouldClearImmediatly(*state, orders) {
			recieverDoorTimer <- true
			orderClearAtCurrentFloor(*state, &orders, rr)
		}
	}
}

// fsmHandleFloorsensorEvent updates the elevator state when arriving at a new floor
func fsmHandleFloorsensorEvent(state *elevator.State, orders elevator.Order, recieverDoorTimer chan<- bool, rr resolvedRequests, floor elevator.Floor) {
	state.Floor = floor
	elevatorio.SetFloorIndicator(floor)
	if state.Behavior == elevator.Moving && orderElevatorShouldStop(*state, orders) {
		elevatorio.SetMotorDirection((0))
		elevatorio.SetDoorOpenLamp(true)
		recieverDoorTimer <- true
		orderClearAtCurrentFloor(*state, &orders, rr)
	}
}

// When timer is done, close the door, and go in desired direction.
func fsmHandleDoorTimerEvent(state *elevator.State, orders elevator.Order, recieverDoorTimer chan<- bool, rr resolvedRequests) {
	if state.Behavior == elevator.DoorOpen {
		fsmChooseDirection(state, orders, recieverDoorTimer)
		if state.Behavior == elevator.DoorOpen {
			recieverDoorTimer <- true
			orderClearAtCurrentFloor(*state, &orders, rr)
		} else {
			elevatorio.SetDoorOpenLamp(false)
			elevatorio.SetMotorDirection(state.Direction)
		}
	}
}

func fsmOpenDoor(state *elevator.State) {
	log.Printf("[elevatorfsm] Door open\n")
	elevatorio.SetDoorOpenLamp(true)
	state.Behavior = elevator.DoorOpen
}

// fsmChooseDirection updates the elevator direction and behaviour based on the current orders. Inspired by the given C-code.
func fsmChooseDirection(e *elevator.State, orders elevator.Order, recieverDoorTimer chan<- bool) {
	switch e.Direction {
	case elevator.Up:
		if orderAbove(*e, orders) {
			e.Direction = elevator.Up
			e.Behavior = elevator.Moving
		} else if orderHere(*e, orders) {
			e.Direction = elevator.Stop
			recieverDoorTimer <- true

		} else if orderBelow(*e, orders) {
			e.Direction = elevator.Down
			e.Behavior = elevator.Moving
		} else {
			e.Direction = elevator.Stop
			e.Behavior = elevator.Idle
		}

	case elevator.Down:
		if orderBelow(*e, orders) {
			e.Direction = elevator.Down
			e.Behavior = elevator.Moving
		} else if orderHere(*e, orders) {
			e.Direction = elevator.Stop
			recieverDoorTimer <- true
		} else if orderAbove(*e, orders) {
			e.Direction = elevator.Up
			e.Behavior = elevator.Moving
		} else {
			e.Direction = elevator.Stop
			e.Behavior = elevator.Idle
		}

	case elevator.Stop:
		if orderHere(*e, orders) {
			e.Direction = elevator.Stop
			recieverDoorTimer <- true
		} else if orderAbove(*e, orders) {
			e.Direction = elevator.Up
			e.Behavior = elevator.Moving
		} else if orderBelow(*e, orders) {
			e.Direction = elevator.Down
			e.Behavior = elevator.Moving
		} else {
			e.Direction = elevator.Stop
			e.Behavior = elevator.Idle
		}
	}
}
