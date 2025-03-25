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
var engineTimerDuration = 10
var elevatorStatePollRate = time.Millisecond * 1000

func RunDriver(pollObstructionSwitch <-chan message.Obstruction,
	pollFloorSensor <-chan message.FloorArrival,
	pollOrders <-chan message.ServiceOrder,
	toRequests chan<- message.RequestState,
	toComms chan<- message.ElevatorState,
	toOrders chan<- message.ElevatorState,
	isSelfAlive chan<- bool,
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
	timerEngine := time.NewTimer((time.Duration(engineTimerDuration)) * time.Second)
	timerEngine.Stop()
	tickerSendElevatorState := time.NewTicker(elevatorStatePollRate)
	isObstructed := false
	localAlive := true

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
			timerEngine.Reset(time.Duration(engineTimerDuration) * time.Second)
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
		case <-timerEngine.C:
			log.Printf("[elevatordriver] Received engine timer message %v", isObstructed)
			if state.Behavior == elevator.Moving || isObstructed {
				localAlive = false
				log.Printf("[elevatordriver] Elevator failure\n")
				handleEngineTimerEvent(&state, order, isSelfAlive)
				timerEngine.Reset(time.Duration(1) * time.Second)
			} else {
				localAlive = true
				timerEngine.Reset(time.Duration(engineTimerDuration) * time.Second)

			}
		case <-tickerSendElevatorState.C:
			toComms <- message.ElevatorState{Elevator: local, State: state, Alive: localAlive}
			toOrders <- message.ElevatorState{Elevator: local, State: state, Alive: localAlive}
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

func clearRequest(id elevator.Id, btn elevator.ButtonType, floor elevator.Floor, c chan<- message.RequestState) {

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
	msg := message.RequestState{
		Source:  id,
		Request: req,
	}
	c <- msg
}
