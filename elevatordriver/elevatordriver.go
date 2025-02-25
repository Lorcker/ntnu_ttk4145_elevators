package elevatordriver

import (
	"group48.ttk4145.ntnu/elevators/models"
)

func Starter(pollObstructionSwitch chan<- bool,
	pollFloorSensor <-chan int,
	pollOrders <-chan models.Orders,
	resolvedRequests chan<- models.Request,
	receiver []chan<- models.ElevatorState) {
	for {
		switch {
		}
	}

}
