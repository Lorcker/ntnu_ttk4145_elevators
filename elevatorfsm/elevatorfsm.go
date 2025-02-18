package elevatorfsm

import (
	"group48.ttk4145.ntnu/elevators/elevatorio"
	"group48.ttk4145.ntnu/elevators/orderserver"
)

type Elevator struct {
	Floor     int
	Behavior  ElevatorBehavior
	Requests  [][3]bool
	Direction elevatorio.MotorDirection
}

type ElevatorBehavior int

const (
	EB_Idle ElevatorBehavior = iota
	EB_DoorOpen
	EB_Moving
)

func HandleOrderEvent(elevator Elevator, orders orderserver.Orders) {}

func HandleFloorsensorEvent(elevator Elevator, floor int) {}

func HandleRequestButtonEvent(elevator Elevator, button elevatorio.ButtonEvent) {}

func HandleDoorTimerEvent(elevator Elevator, timer bool) {
	// Remember to check for obstruction
}
