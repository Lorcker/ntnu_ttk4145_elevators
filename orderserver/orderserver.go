package orderserver

import (
	"fmt"
	"time"

	"group48.ttk4145.ntnu/elevators/models"
)

type localElevator struct {
	requests [][3]bool
	models.ElevatorState
}

type elevatorstate struct {
	models.ElevatorState
	cabRequests []bool
}

type elevators struct {
	requests [][2]bool
	states   []elevatorstate
}

// Constants for the duration of the door opening and closing and the time it takes to travel between floors
// SHOULD BE MOVED TO A CONFIG FILE
const (
	doorOpenDuration = time.Second * 3
	travelDuration   = time.Second * 2
	numFloors        = 4
)

func RunOrderServer(
	validatedRequests <-chan models.Request,
	state <-chan models.ElevatorState,
	alive <-chan []models.Id,
	orders chan<- models.Orders) {

	//init local vars
	elevators := elevators{}
	elevators.requests = make([][2]bool, numFloors)
	for i := range elevators.requests {
		elevators.requests[i] = [2]bool{false, false}
	}
	for {
		select {
		case r := <-validatedRequests:
			// if the request is confirmed, add it to the orders channel
			if r.Status == models.Confirmed && elevators.states != nil {
				// add the request to the orders channel
				if r.Origin.ButtonType == models.HallUp || r.Origin.ButtonType == models.HallDown {
					elevators.requests[r.Origin.Floor][r.Origin.ButtonType] = true
				} else {
					for _, elevator := range elevators.states {
						if models.Id(elevator.Id) == r.Origin.Source.(models.Elevator).Id {
							elevator.cabRequests[r.Origin.Floor] = true
						}
					}
				}
				// calculates the optimal orders for the elevators

				orders <- optimalHallRequests(elevators)[1]
			}
		// handle the alive channel
		case a := <-alive:
			// check if the elevator is already in the list of elevators
			for _, id := range a {
				found := false
				for _, elevator := range elevators.states {
					if elevator.Id == id {
						found = true
						break
					}
				}
				// if the elevator is not found in the list of elevators, add it
				if !found {
					elevators.states = append(elevators.states, elevatorstate{
						ElevatorState: models.ElevatorState{
							Id: id,
						},
					})
				}
			}
		case s := <-state:
			fmt.Println("State received", s)
			// update the state of the elevator
			print("ElevatorSates", elevators.states)
			for index, elevState := range elevators.states {
				if s.Id == elevState.Id {
					elevators.states[index] = elevatorstate{
						ElevatorState: s,
					}
					fmt.Println("Elevator state updated", s)
					break
				} else {
					fmt.Println("Elevator not found")
				}

			}
			isInBounds := func(f int) bool { return f >= 0 && f < numFloors }
			for _, state := range elevators.states {
				if !isInBounds(state.Floor) {
					panic("Some elevator is at an invalid floor")
				}
				if state.Behavior == models.Moving && !isInBounds(state.Floor+int(state.Direction)) {
					panic("Some elevator is moving away from an end floor")
				}
			}
		}
	}

}
