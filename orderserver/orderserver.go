package orderserver

import (
	"log"

	"group48.ttk4145.ntnu/elevators/models"
	m "group48.ttk4145.ntnu/elevators/models"
)

func RunOrderServer(
	validatedRequests <-chan models.Request,
	state <-chan models.ElevatorState,
	alive <-chan []models.Id,
	orders chan<- models.Orders,
	localPeerId models.Id) {

	hallRequests := HallRequests{}
	for i := range hallRequests {
		hallRequests[i] = [2]bool{false, false}
	}

	cabRequests := make(map[models.Id]CabRequests)
	elevators := make(map[models.Id]m.ElevatorState)
	alivePeers := make(map[models.Id]bool)

	//init local vars
	for {
		select {
		case r := <-validatedRequests:
			if r.Status == models.Unconfirmed || r.Status == models.Unknown {
				// These are not relevant for the order sever
				continue
			}

			log.Printf("[orderserver] Received validated request: %v", r)

			status := r.Status == models.Confirmed // convert the status to a boolean

			if _, ok := r.Origin.Source.(models.Hall); ok {
				hallRequests[r.Origin.Floor][r.Origin.ButtonType] = status
			} else {
				cr := cabRequests[r.Origin.Source.(m.Elevator).Id]
				cr[r.Origin.Floor] = status
				cabRequests[r.Origin.Source.(m.Elevator).Id] = cr
			}

			os := calculateOrders(hallRequests, cabRequests, elevators)
			log.Printf("[orderserver] Calculated new orders: %v", os)
			order := os[localPeerId]
			log.Printf("[orderserver] Started sending new Orders to [driver]: %v", order)
			orders <- order
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
					delete(elevators, id)
					delete(cabRequests, id)
					delete(alivePeers, id)
				}
			}

		case newState := <-state:
			//log.Printf("[orderserver] Received new state:\n\tinternal: %v\n\tnew: %v", elevators[newState.Id], newState)
			currentState, ok := elevators[newState.Id]

			if !ok {
				elevators[newState.Id] = newState
				log.Printf("[orderserver] Added a new elevator to internal memory with state: %v", newState)
			} else if !models.IsEStateEqual(currentState, newState) {
				log.Printf("[orderserver] Updated buffered elevator state from %v to %v", currentState, newState)
				elevators[newState.Id] = newState

				os := calculateOrders(hallRequests, cabRequests, elevators)
				log.Printf("[orderserver] Calculated new orders: %v", os)
				order := os[localPeerId]
				log.Printf("[orderserver] Started sending new Orders to [driver]: %v", order)
				orders <- order
			}
		}
	}

}
