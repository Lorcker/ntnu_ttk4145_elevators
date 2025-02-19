package orderserver

import (
	"group48.ttk4145.ntnu/elevators/models"
)

func RunOrderServer(
	validatedRequests <-chan models.Request,
	state <-chan models.ElevatorState,
	alive <-chan []uint8,
	orders chan<- models.Orders) {

	//init vars
	numFloors := 4
	hallRequests := make([][2]bool, numFloors)
	cabRequests := make([]bool, numFloors)
	elevatorStates := []models.ElevatorState{}

	select {
	case r := <-validatedRequests:
		if r.Status == models.Confirmed {
			// add the request to the orders channel
			if (r.Origin.Source == models.Hall{}) {
				hallRequests[r.Origin.Floor][r.Origin.ButtonType] = true
			} else {
				cabRequests[r.Origin.Floor] = true
			}
		}
	case a := <-alive:
		for _, elevator := range a {
			found := false
			for _, state := range elevatorStates {
				if state.Id == elevator {
					found = true
					break
				}
			}
			if !found {
				elevatorStates = append(elevatorStates, models.ElevatorState{Id: elevator})
			}
		}
	case s := <-state:
		exists := false
		for index, elevState := range elevatorStates {
			if s.Id == elevState.Id {
				exists = true
				elevatorStates[index] = s
				break
			}
		}
		if !exists {
			panic("State Alive")
		}

		if len(elevatorStates) == 0 {
			panic("No elevator states provided")
		}
		for _, state := range elevatorStates {
			if len(cabRequests) != numFloors {
				panic("Hall and cab requests do not all have the same length")
			}
		}

		isInBounds := func(f int) bool { return f >= 0 && f < numFloors }
		for _, state := range elevatorStates {
			if !isInBounds(state.Floor) {
				panic("Some elevator is at an invalid floor")
			}
			if state.Behavior == models.Moving && !isInBounds(state.Floor+int(state.Direction)) {
				panic("Some elevator is moving away from an end floor")
			}
		}
		// add the logic for the optimal hall requests here
		// and send the outgoingorders to the orders channel
		// orders <- outgoingOrders

	}
}
