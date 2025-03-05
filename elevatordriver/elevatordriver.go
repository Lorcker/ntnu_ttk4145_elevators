package elevatordriver

import (
	"log"
	"time"

	"group48.ttk4145.ntnu/elevators/models"
)

// Global variables
var doorTimerDuration = 3
var sendElevatorStateDuration = 1
var NButtons int = 3
var NFloors int = 4

func Starter(pollObstructionSwitch <-chan bool,
	pollFloorSensor <-chan int,
	pollOrders <-chan models.Orders,
	resolvedRequests chan<- models.RequestMessage,
	receiver []chan<- models.ElevatorState,
	id models.Id) {

	// Init elevator, obstruction and timer
	elevator := models.ElevatorState{Id: (uint8(id)), Floor: 0, Behavior: models.Idle, Direction: models.MotorDirection(0)}
	orders := initOrders(NFloors)
	initElevator(orders)

	recieverStartDoorTimer := make(chan bool, 10)
	timerDoor := time.NewTimer((time.Duration(doorTimerDuration)) * time.Second)
	timerDoor.Stop()
	timerSendElevatorState := time.NewTimer(time.Duration(sendElevatorStateDuration) * time.Second)
	isObstructed := false

	for {
		select {
		case orders = <-pollOrders:
			log.Printf("[elevatordriver] Received new orders: %v", orders)
			HandleOrderEvent(&elevator, orders, recieverStartDoorTimer)

		case floor_sensor := <-pollFloorSensor:
			log.Printf("[elevatordriver] Received floor sensor: %v", floor_sensor)
			HandleFloorsensorEvent(&elevator, orders, floor_sensor, recieverStartDoorTimer)

			// Send resolved requests
			o := models.Origin{Source: models.Hall{}, Floor: floor_sensor, ButtonType: models.HallUp}
			r := models.Request{Origin: o, Status: models.Absent}
			resolvedRequests <- models.RequestMessage{Source: id, Request: r}

			o = models.Origin{Source: models.Hall{}, Floor: floor_sensor, ButtonType: models.HallDown}
			r = models.Request{Origin: o, Status: models.Absent}
			resolvedRequests <- models.RequestMessage{Source: id, Request: r}

			o = models.Origin{Source: models.Elevator{Id: id}, Floor: floor_sensor, ButtonType: models.Cab}
			r = models.Request{Origin: o, Status: models.Absent}
			resolvedRequests <- models.RequestMessage{Source: id, Request: r}

		case <-recieverStartDoorTimer:
			log.Printf("[elevatordriver] Received open door message")
			OpenDoor(&elevator)
			timerDoor.Reset(time.Duration(doorTimerDuration) * time.Second)

		case <-pollObstructionSwitch:
			log.Printf("[elevatordriver] Received obstruction message")
			isObstructed = !isObstructed

		case <-timerDoor.C:
			log.Printf("[elevatordriver] Received door closed message")
			if elevator.Behavior == models.DoorOpen && !isObstructed {
				HandleDoorTimerEvent(&elevator, orders, recieverStartDoorTimer)
			} else {
				timerDoor.Reset(time.Duration(doorTimerDuration) * time.Second)
			}
		case <-timerSendElevatorState.C:
			for _, ch := range receiver {
				ch <- elevator
			}
			timerSendElevatorState.Reset(time.Duration(sendElevatorStateDuration) * time.Second)

		}
	}
}
