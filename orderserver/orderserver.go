package orderserver

import (
	"log"
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
	hallRequests [][2]bool
	states       map[models.Id]elevatorstate
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
	orders chan<- models.Orders,
	localPeerId models.Id) {

	elevators := newElevators()
	alivePeers := make(map[models.Id]bool)

	//init local vars
	for {
		select {
		case r := <-validatedRequests:
			log.Printf("[orderserver] Received validated request: %v", r)

			if r.Status == models.Unconfirmed || r.Status == models.Unknown {
				// These are not relevant for the order sever
				continue
			}

			status := r.Status == models.Confirmed // convert the status to a boolean

			if len(elevators.states) > 0 {
				// add the request to the orders channel
				if _, ok := r.Origin.Source.(models.Hall); ok {
					elevators.hallRequests[r.Origin.Floor][r.Origin.ButtonType] = status
				} else {
					elevators.states[r.Origin.Source.(models.Elevator).Id].cabRequests[r.Origin.Floor] = status
				}

				// calculates the optimal orders for the elevators
				order := optimalHallRequests(elevators)[localPeerId]
				log.Printf("[orderserver] Turned requests into order: %v", order)
				orders <- order
				log.Printf("[orderserver] Send order to channel: %v", order)
			}

		// handle the alive channel
		case a := <-alive:
			log.Printf("[orderserver] Received alive status: %v", a)

			newAlive := make(map[models.Id]bool)
			for _, id := range a {
				newAlive[id] = true
				alivePeers[id] = true // add the peer to the alivePeers map
			}

			// Check if a peer died
			for id := range alivePeers {
				if _, ok := newAlive[id]; !ok {
					// peer died - remove the peer from the states and alivePeers map to exclude it from the calculations
					delete(elevators.states, id)
					delete(alivePeers, id)
				}
			}

		case newState := <-state:
			currentState, ok := elevators.states[newState.Id]
			if !ok {
				elevators.states[newState.Id] = elevatorstate{ElevatorState: newState, cabRequests: make([]bool, numFloors)}
				log.Printf("[orderserver] Added a new elevator to internal memory with state: %v", newState)
				break
			}
			if currentState.ElevatorState != newState {
				log.Printf("[orderserver] Updated buffered elevator state from %v to %v", currentState.ElevatorState, newState)
				currentState.ElevatorState = newState
				elevators.states[newState.Id] = currentState
			}
		}
	}

}

func newElevators() elevators {
	elevators := elevators{}
	elevators.states = make(map[models.Id]elevatorstate)
	elevators.hallRequests = make([][2]bool, numFloors)
	for i := range elevators.hallRequests {
		elevators.hallRequests[i] = [2]bool{false, false}
	}
	return elevators
}
