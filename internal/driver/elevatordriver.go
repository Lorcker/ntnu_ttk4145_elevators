package driver

import (
	"log"
	"time"

	"group48.ttk4145.ntnu/elevators/internal/elevatorio"
	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/message"
	"group48.ttk4145.ntnu/elevators/internal/models/request"
)

// Global variables
var doorTimerDuration = 3
var elevatorStatePollRate = time.Millisecond * 1000

func RunDriver(pollObstructionSwitch <-chan message.ObstructionSwitch,
	pollFloorSensor <-chan message.FloorSensor,
	pollOrders <-chan message.Order,
	toRequests chan<- message.RequestStateUpdate,
	toComms chan<- message.ElevatorStateUpdate,
	toOrders chan<- message.ElevatorStateUpdate,
	local elevator.Id) {

	// Init state, obstruction and timer
	state := elevator.State{
		Floor:     0,
		Behavior:  elevator.Idle,
		Direction: elevator.Stop}
	order := elevator.Order{}
	driveToStaringPosition()

	receiverStartDoorTimer := make(chan bool, 10)
	timerDoor := time.NewTimer((time.Duration(doorTimerDuration)) * time.Second)
	timerDoor.Stop()
	tickerSendElevatorState := time.NewTicker(elevatorStatePollRate)
	isObstructed := false

	clearRequestFun := func(btn elevator.ButtonType, floor elevator.Floor) {
		clearRequest(local, btn, floor, toRequests)
	}

	for {
		select {
		case msg := <-pollOrders:
			order = msg.Order
			log.Printf("[elevatordriver] Received new orders:\n\t%v", elevator.OrderToString(order))
			handleOrderEvent(&state, order, receiverStartDoorTimer, clearRequestFun)

		case msg := <-pollFloorSensor:
			log.Printf("[elevatordriver] Received floor sensor: %v", msg)
			handleFloorsensorEvent(&state, order, msg.Floor, receiverStartDoorTimer, clearRequestFun)

		case <-receiverStartDoorTimer:
			log.Printf("[elevatordriver] Received open door message")
			openDoor(&state)
			timerDoor.Reset(time.Duration(doorTimerDuration) * time.Second)

		case <-pollObstructionSwitch:
			log.Printf("[elevatordriver] Received obstruction message")
			isObstructed = !isObstructed
			if state.Behavior == elevator.DoorOpen {
				timerDoor.Reset(time.Duration(doorTimerDuration) * time.Second)
			}

		case <-timerDoor.C:
			log.Printf("[elevatordriver] Received door closed message")
			if state.Behavior == elevator.DoorOpen && !isObstructed {
				handleDoorTimerEvent(&state, order, receiverStartDoorTimer, clearRequestFun)
			} else {
				timerDoor.Reset(time.Duration(doorTimerDuration) * time.Second)
			}
		case <-tickerSendElevatorState.C:
			toComms <- message.ElevatorStateUpdate{Elevator: local, State: state}
			toOrders <- message.ElevatorStateUpdate{Elevator: local, State: state}
		}
	}
}

func driveToStaringPosition() {
	if floor := elevatorio.GetFloor(); floor != 0 {
		elevatorio.SetMotorDirection(-1)
		for elevatorio.GetFloor() != 0 {
		}
		elevatorio.SetMotorDirection(0)
	}
}

func clearRequest(id elevator.Id, btn elevator.ButtonType, floor elevator.Floor, c chan<- message.RequestStateUpdate) {
	log.Printf("[elevatordriver] Cleared request at floor %v, button %v", floor, btn)
	var req request.Request
	switch btn {
	case elevator.Cab:
		req = request.NewCabRequest(floor, id, request.Absent)
	case elevator.HallUp:
		req = request.NewHallRequest(floor, request.Up, request.Absent)
	case elevator.HallDown:
		req = request.NewHallRequest(floor, request.Down, request.Absent)
	}
	msg := message.RequestStateUpdate{
		Source:  id,
		Request: req,
	}
	c <- msg
}
