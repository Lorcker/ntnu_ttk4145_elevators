package comms

import (
	"fmt"
	"log"
	"strconv"

	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/message"
	"group48.ttk4145.ntnu/elevators/internal/models/request"
)

// The requestRegistry holds information about all system requests.
// It is needed as every sending cycle of comms must propagate all system request to other peers.
// The internal system work with sending messaged on change.
// This does not work for comms as packet loss is guaranteed to happen so changes might get lost.
// Thus the registry stores the change and can calculate the diff between two registries to
// enable the conversion back to the internal messaging model.
// Also, in case one elevator dies, the information is backed up here.
type requestRegistry struct {
	// HallUp and HallDown are arrays of request status where the index is the floor
	HallUp   []request.Status
	HallDown []request.Status

	// Map uses the id of the elevator as key
	// Is a string because the json conversion of network module only allows for strings
	// The value is an array of request status where the index is the floor
	Cab map[string][]request.Status
}

func newRequestRegistry() requestRegistry {
	hu := make([]request.Status, elevator.NumFloors)
	hd := make([]request.Status, elevator.NumFloors)
	c := make(map[string][]request.Status)

	for i := elevator.Floor(0); i < elevator.NumFloors; i++ {
		hu[i] = request.Unknown
		hd[i] = request.Unknown
	}

	return requestRegistry{
		HallUp:   hu,
		HallDown: hd,
		Cab:      c,
	}
}

// Adds a new cab to the registry
func (r *requestRegistry) initNewCab(id string) {
	cab := make([]request.Status, elevator.NumFloors)
	for i := elevator.Floor(0); i < elevator.NumFloors; i++ {
		cab[i] = request.Unknown
	}

	r.Cab[id] = cab
}

// update takes in a internal msg from the request module and replaces the stored information
// As the msg were validated by the request module no checks on the status information are needed
func (r *requestRegistry) update(req request.Request) {
	if req.Status == request.Unknown {
		// Ignore unknown requests as they add no value
		return
	}
	floor := req.Origin.GetFloor()

	if request.IsFromHall(req) {
		dir := req.Origin.(request.Hall).Direction
		if dir == request.Up {
			r.HallUp[floor] = req.Status
		} else {
			r.HallDown[floor] = req.Status
		}
	} else {
		id := req.Origin.(request.Cab).Id
		idS := strconv.Itoa(int(id))

		// Check is needed because if comms get info about an elevator it has not seen before
		// it need to adds to the registry and keep it there
		if _, ok := r.Cab[idS]; !ok {
			r.initNewCab(idS)
		}

		// Reassign updated slice to map as no direct update is possible in Go
		cabRequests := r.Cab[idS]
		cabRequests[floor] = req.Status
		r.Cab[idS] = cabRequests
	}
}

// diff calculates the difference between two registries
// and returns a slice of requestMessage where each represents a differing entry
// If both states are Unconfirmed the entry is also included to enable acknoledgement of the request
func (r *requestRegistry) diff(peer elevator.Id, other requestRegistry) []message.RequestStateUpdate {
	var diff []message.RequestStateUpdate

	for floor := elevator.Floor(0); floor < elevator.NumFloors; floor++ {
		if isDifferent(r.HallUp[floor], other.HallUp[floor]) {
			diff = append(diff, message.RequestStateUpdate{
				Source:  peer,
				Request: request.NewHallRequest(floor, request.Up, other.HallUp[floor]),
			})
		}
		if isDifferent(r.HallDown[floor], other.HallDown[floor]) {
			diff = append(diff, message.RequestStateUpdate{
				Source:  peer,
				Request: request.NewHallRequest(floor, request.Down, other.HallDown[floor]),
			})
		}
	}

	for id, otherCab := range other.Cab {
		localCab, ok := r.Cab[id]
		idI, err := strconv.Atoi(id)

		if err != nil {
			log.Fatalf("[comms] failed to convert stored elevator id string to its uint: %e", err)
		}

		if !ok {
			for f := elevator.Floor(0); f < elevator.NumFloors; f++ {
				if isDifferent(request.Unknown, otherCab[f]) {
					diff = append(diff, message.RequestStateUpdate{
						Source:  peer,
						Request: request.NewCabRequest(f, elevator.Id(idI), otherCab[f]),
					})
				}
			}
			continue
		}

		for f := elevator.Floor(0); f < elevator.NumFloors; f++ {
			if isDifferent(localCab[f], otherCab[f]) {
				diff = append(diff, message.RequestStateUpdate{
					Source:  peer,
					Request: request.NewCabRequest(f, elevator.Id(idI), otherCab[f]),
				})
			}
		}
	}

	return diff
}

// isDifferent checks if two request status are different
// If both are Unconfirmed the function returns true to enable acknoledgement of the request
// If the external status is Unkown we ignore it
func isDifferent(a, b request.Status) bool {
	if a == request.Unconfirmed && b == request.Unconfirmed {
		return true
	}
	if b == request.Unknown {
		return false
	}
	return a != b
}

// String returns a string representation of the request registry
func (r *requestRegistry) String() string {
	str := fmt.Sprintf("HallUp: %v, HallDown: %v, Cabs: %v", r.HallUp, r.HallDown, r.Cab)
	return str
}
