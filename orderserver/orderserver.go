package orderserver

import (
	"group48.ttk4145.ntnu/elevators/statedataserver"
)

type Orders [][]bool // first index is elevator, second index is floor

func PollOrders(receiver chan<- Orders) {

}

func optimalHallRequests(
	hallRequests [][2]bool,
	elevatorStates []statedataserver.GlobalElevator) Orders {

	numFloors := len(hallRequests[0])
	if len(elevatorStates) == 0 {
		panic("No elevator states provided")
	}
	for _, state := range elevatorStates {
		if len(state.CabRequests) != numFloors {
			panic("Hall and cab requests do not all have the same length")
		}
	}

	isInBounds := func(f int) bool { return f >= 0 && f < numFloors }
	for _, state := range elevatorStates {
		if !isInBounds(state.Floor) {
			panic("Some elevator is at an invalid floor")
		}
		if state.Behavior == statedataserver.EB_Moving && !isInBounds(state.Floor+int(state.Direction)) {
			panic("Some elevator is moving away from an end floor")
		}
	}
	// add the logic for the optimal hall requests here
	return Orders{}
}
