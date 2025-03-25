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
		fsmChooseDirection(state, orders) // Updates the behaviour and direction
		if state.Behavior == elevator.DoorOpen {
			recieverDoorTimer <- true
			ordersClearAtCurrentFloor(*state, &orders, rr) // Clears orders that is handled at the current floor.
		} else if state.Behavior == elevator.Moving {
			elevatorio.SetMotorDirection(state.Direction)
		}

	case elevator.DoorOpen:
		if ordersShouldClearImmediatly(*state, orders) { //If it is a order at the current floor that should be handled.
			recieverDoorTimer <- true
			ordersClearAtCurrentFloor(*state, &orders, rr)
		}
	}
}

// fsmHandleFloorsensorEvent updates the elevator state when arriving at a new floor
func fsmHandleFloorsensorEvent(state *elevator.State, orders elevator.Order, recieverDoorTimer chan<- bool, rr resolvedRequests, floor elevator.Floor) {
	state.Floor = floor
	elevatorio.SetFloorIndicator(floor)
	if state.Behavior == elevator.Moving && ordersElevatorShouldStop(*state, orders) {
		elevatorio.SetMotorDirection((0))
		fsmOpenDoor(state)
		recieverDoorTimer <- true
		ordersClearAtCurrentFloor(*state, &orders, rr)
	}
}

// When the door timer is finished, fsmHandleDoorTimerEvent closes the door, and sends the elevator in the desired direction.
func fsmHandleDoorTimerEvent(state *elevator.State, orders elevator.Order, recieverDoorTimer chan<- bool, rr resolvedRequests) {
	if state.Behavior == elevator.DoorOpen {
		fsmChooseDirection(state, orders) // updates the behaviour and direction of the elevator
		if state.Behavior == elevator.DoorOpen {
			recieverDoorTimer <- true
			ordersClearAtCurrentFloor(*state, &orders, rr)
		} else {
			elevatorio.SetDoorOpenLamp(false)
			elevatorio.SetMotorDirection(state.Direction)
		}
	}
}

// fsmOpenDoor updates elevator behaviour to doorOpen, and sets the light
func fsmOpenDoor(state *elevator.State) {
	log.Printf("[elevatorfsm] Door open\n")
	elevatorio.SetDoorOpenLamp(true)
	state.Behavior = elevator.DoorOpen
}

// fsmChooseDirection calculates and updates the elevator direction and behaviour based on the current orders. Inspired by the given C-code.
func fsmChooseDirection(e *elevator.State, orders elevator.Order) {
	switch e.Direction {
	case elevator.Up:
		if ordersAbove(*e, orders) {
			e.Direction = elevator.Up
			e.Behavior = elevator.Moving
		} else if ordersHere(*e, orders) {
			e.Direction = elevator.Stop
			fsmOpenDoor(e)

		} else if ordersBelow(*e, orders) {
			e.Direction = elevator.Down
			e.Behavior = elevator.Moving
		} else {
			e.Direction = elevator.Stop
			e.Behavior = elevator.Idle
		}

	case elevator.Down:
		if ordersBelow(*e, orders) {
			e.Direction = elevator.Down
			e.Behavior = elevator.Moving
		} else if ordersHere(*e, orders) {
			e.Direction = elevator.Stop
			fsmOpenDoor(e)
		} else if ordersAbove(*e, orders) {
			e.Direction = elevator.Up
			e.Behavior = elevator.Moving
		} else {
			e.Direction = elevator.Stop
			e.Behavior = elevator.Idle
		}

	case elevator.Stop:
		if ordersHere(*e, orders) {
			e.Direction = elevator.Stop
			fsmOpenDoor(e)
		} else if ordersAbove(*e, orders) {
			e.Direction = elevator.Up
			e.Behavior = elevator.Moving
		} else if ordersBelow(*e, orders) {
			e.Direction = elevator.Down
			e.Behavior = elevator.Moving
		} else {
			e.Direction = elevator.Stop
			e.Behavior = elevator.Idle
		}
	}
}
