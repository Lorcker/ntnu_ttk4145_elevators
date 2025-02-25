package orderserver

import (
	"group48.ttk4145.ntnu/elevators/models"
)

type elevator struct {
	cabRequests []bool
	state       models.ElevatorState
}

func RunOrderServer(
	validatedRequests <-chan models.Request,
	state <-chan models.ElevatorState,
	alive <-chan []uint8,
	orders chan<- models.Orders) {

	//init vars
	numFloors := 4
	hallRequests := make([][2]bool, numFloors)
	elevators := []elevator{}

	select {
	case r := <-validatedRequests:
		if r.Status == models.Confirmed {
			// add the request to the orders channel
			for _, elevator := range elevators {
				if source, ok := r.Origin.Source.(models.Elevator); ok && uint8(source.Id) == elevator.state.Id {
					elevator.cabRequests[r.Origin.Floor] = true
				} else if _, ok := r.Origin.Source.(models.Hall); ok {
					hallRequests[r.Origin.Floor][r.Origin.ButtonType] = true
				} else {
					panic("Invalid request source")
				}
			}
		}
	case a := <-alive:
		for _, id := range a {
			found := false
			for _, elevator := range elevators {
				if elevator.state.Id == id {
					found = true
					break
				}
			}
			if !found {
				elevators = append(elevators, elevator{
					cabRequests: make([]bool, numFloors),
					state: models.ElevatorState{
						Id: id,
					},
				})
			}
		}
	case s := <-state:
		exists := false
		for index, elevState := range elevators {
			if s.Id == elevState.state.Id {
				exists = true
				elevators[index].state = s
				break
			}
		}
		if !exists {
			panic("State Alive")
		}

		if len(elevators) == 0 {
			panic("No elevator states provided")
		}
		for _, state := range elevators {
			if len(state.cabRequests) != numFloors {
				panic("Hall and cab requests do not all have the same length")
			}
		}

		isInBounds := func(f int) bool { return f >= 0 && f < numFloors }
		for _, state := range elevators {
			if !isInBounds(state.state.Floor) {
				panic("Some elevator is at an invalid floor")
			}
			if state.state.Behavior == models.Moving && !isInBounds(state.state.Floor+int(state.state.Direction)) {
				panic("Some elevator is moving away from an end floor")
			}
		}
		// add the logic for the optimal hall requests here
		// and send the outgoingorders to the orders channel
		// orders <- outgoingOrders

	}
}
