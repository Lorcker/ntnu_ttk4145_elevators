package driver

import (
	"log"

	"group48.ttk4145.ntnu/elevators/internal/elevatorio"
	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
)

const EverybodyGoesOn bool = false

type ResolvedRequests func(btn elevator.ButtonType, floor elevator.Floor)

type ElevatorFSM struct {
	state             *elevator.State
	orders            elevator.Order
	recieverDoorTimer chan<- bool
	rr                ResolvedRequests
}

func ElevatorFSMInit(state elevator.State, orders elevator.Order, recieverDoorTimer chan<- bool, rr ResolvedRequests) *ElevatorFSM {
	return &ElevatorFSM{&state, orders, recieverDoorTimer, rr}
}

// HandleOrderEvent updates the elevator state based on new orders
func (fsm *ElevatorFSM) HandleOrderEvent() {
	switch fsm.state.Behavior {
	case elevator.Idle:
		fsm.ChooseDirection() // Updates the behaviour and direction
		if fsm.state.Behavior == elevator.DoorOpen {
			fsm.recieverDoorTimer <- true
			fsm.ordersClearAtCurrentFloor() // Clears orders that is handled at the current floor.
		} else if fsm.state.Behavior == elevator.Moving {
			elevatorio.SetMotorDirection(fsm.state.Direction)
		}

	case elevator.DoorOpen:
		if ordersShouldClearImmediatly(*fsm.state, fsm.orders) { //If it is a order at the current floor that should be handled.
			fsm.recieverDoorTimer <- true
			fsm.ordersClearAtCurrentFloor()
		}
	}
}

// HandleFloorsensorEvent updates the elevator state when arriving at a new floor
func (fsm *ElevatorFSM) HandleFloorsensorEvent(floor elevator.Floor) {
	fsm.state.Floor = floor
	elevatorio.SetFloorIndicator(floor)
	if fsm.ordersElevatorShouldStop() {
		elevatorio.SetMotorDirection((0))
		fsm.OpenDoor()
		fsm.recieverDoorTimer <- true
		fsm.ordersClearAtCurrentFloor()
	}
}

// When the door timer is finished, HandleDoorTimerEvent closes the door, and sends the elevator in the desired direction.
func (fsm *ElevatorFSM) HandleDoorTimerEvent() {
	if fsm.state.Behavior != elevator.DoorOpen {
		return
	}

	fsm.ChooseDirection() // updates the behaviour and direction of the elevator
	if fsm.state.Behavior == elevator.DoorOpen {
		fsm.recieverDoorTimer <- true
		fsm.ordersClearAtCurrentFloor()
	} else {
		elevatorio.SetDoorOpenLamp(false)
		elevatorio.SetMotorDirection(fsm.state.Direction)
	}
}

// OpenDoor updates elevator behaviour to doorOpen, and sets the light
func (fsm *ElevatorFSM) OpenDoor() {
	log.Printf("[elevatorfsm] Door open\n")
	elevatorio.SetDoorOpenLamp(true)
	fsm.state.Behavior = elevator.DoorOpen
}

// ChooseDirection calculates and updates the elevator direction and behaviour based on the current orders. Inspired by the given C-code.
func (fsm *ElevatorFSM) ChooseDirection() {
	switch fsm.state.Direction {
	case elevator.Up:
		if fsm.ordersAbove() {
			fsm.state.Direction = elevator.Up
			fsm.state.Behavior = elevator.Moving
		} else if fsm.ordersHere() {
			fsm.state.Direction = elevator.Stop
			fsm.OpenDoor()

		} else if fsm.ordersBelow() {
			fsm.state.Direction = elevator.Down
			fsm.state.Behavior = elevator.Moving
		} else {
			fsm.state.Direction = elevator.Stop
			fsm.state.Behavior = elevator.Idle
		}

	case elevator.Down:
		if fsm.ordersBelow() {
			fsm.state.Direction = elevator.Down
			fsm.state.Behavior = elevator.Moving
		} else if fsm.ordersHere() {
			fsm.state.Direction = elevator.Stop
			fsm.OpenDoor()
		} else if fsm.ordersAbove() {
			fsm.state.Direction = elevator.Up
			fsm.state.Behavior = elevator.Moving
		} else {
			fsm.state.Direction = elevator.Stop
			fsm.state.Behavior = elevator.Idle
		}

	case elevator.Stop:
		if fsm.ordersHere() {
			fsm.state.Direction = elevator.Stop
			fsm.OpenDoor()
		} else if fsm.ordersAbove() {
			fsm.state.Direction = elevator.Up
			fsm.state.Behavior = elevator.Moving
		} else if fsm.ordersBelow() {
			fsm.state.Direction = elevator.Down
			fsm.state.Behavior = elevator.Moving
		} else {
			fsm.state.Direction = elevator.Stop
			fsm.state.Behavior = elevator.Idle
		}
	}
}
