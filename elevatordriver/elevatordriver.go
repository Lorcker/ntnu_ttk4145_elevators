package elevatordriver

import (
	"group48.ttk4145.ntnu/elevators/elevatorfsm"
	"group48.ttk4145.ntnu/elevators/elevatorio"
	"group48.ttk4145.ntnu/elevators/orderserver"
)

func Starter(pollObstructionSwitch chan<- bool,
	pollFloorSensor <-chan bool,
	pollButtons <-chan elevatorio.ButtonEvent,
	doorTimer <-chan bool,
	pollOrders <-chan orderserver.Orders) {

}

func PollElevatorState(receiver chan<- elevatorfsm.Elevator) {}
