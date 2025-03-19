package orders

import (
	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/request"
)

// hallRequests is a 2D array of booleans, where the first dimension is the floor and the second dimension is the direction.
type hallRequests = [elevator.NumFloors][2]bool

// cabRequests is an array of booleans, where the index is the floor.
type cabRequests = [elevator.NumFloors]bool

// cache stores the latest requests, elevator states and alive information
type cache struct {
	Hr         hallRequests
	Cr         map[elevator.Id]cabRequests
	States     map[elevator.Id]elevator.State
	AlivePeers map[elevator.Id]bool
}

func newCache() *cache {
	return &cache{
		Hr:         hallRequests{},
		Cr:         make(map[elevator.Id]cabRequests),
		States:     make(map[elevator.Id]elevator.State),
		AlivePeers: make(map[elevator.Id]bool),
	}
}

// AddRequest adds a request to the cache and returns true if the cache changed
func (r *cache) AddRequest(req request.Request) (didChange bool) {
	status := req.Status == request.Confirmed
	if request.IsFromHall(req) {
		return r.addHallRequest(req.Origin.GetFloor(), req.Origin.(request.Hall).Direction, status)
	}

	// Must be a cab request
	return r.addCabRequest(req.Origin.(request.Cab).Id, req.Origin.GetFloor(), status)
}

// addHallRequest adds a hall request to the cache and returns true if the cache changed
func (r *cache) addHallRequest(floor elevator.Floor, direction request.Direction, status bool) (didChange bool) {
	hr := r.Hr
	oldStatus := hr[floor][direction]
	hr[floor][direction] = status
	r.Hr = hr
	return oldStatus != status
}

// addCabRequest adds a cab request to the cache and returns true if the cache changed
func (r *cache) addCabRequest(id elevator.Id, floor elevator.Floor, status bool) (didChange bool) {
	cr, ok := r.Cr[id]
	if !ok {
		cr = cabRequests{}
	}
	oldStatus := cr[floor]
	cr[floor] = status
	r.Cr[id] = cr
	return oldStatus != status
}

// AddElevatorState adds an elevator state to the cache and returns true if the cache changed
func (r *cache) AddElevatorState(id elevator.Id, state elevator.State) (didChange bool) {
	oldState, ok := r.States[id]
	if !ok {
		r.States[id] = state
		r.Cr[id] = cabRequests{} // Initialize cab requests for the new elevator
		return true
	}

	if oldState == state {
		return false
	}

	r.States[id] = state
	return true
}

// ProcessAliveUpdate updates the cache with the latest alive information
//
// If a peer is no longer alive, the peer is removed from the cache.
// This includes removing the peer's elevator state, cab requests and alive status.
// This ensures that the peer is not included anymore in the order calculations.
func (r *cache) ProcessAliveUpdate(alive []elevator.Id) {
	newAlive := make(map[elevator.Id]bool)
	for _, id := range alive {
		newAlive[id] = true
		r.AlivePeers[id] = true
	}

	// Check if a peer died
	for id := range r.AlivePeers {
		if _, ok := newAlive[id]; !ok {
			// peer died - remove the peer from the states and alivePeers map to exclude it from the calculations
			delete(r.States, id)
			delete(r.Cr, id)
			delete(r.AlivePeers, id)
		}
	}
}

// IsConsistent checks if the cache is consistent
//
// The cache is consistent if all alive elevators have cab requests and elevator states in the cache
// and all cab requests have a corresponding elevator state and vice versa
func (r *cache) IsConsistent() bool {
	// Check that all alive elevators have cab requests and elevator states in the cache
	for id := range r.AlivePeers {
		if _, ok := r.Cr[id]; !ok {
			return false
		}
		if _, ok := r.States[id]; !ok {
			return false
		}
	}

	// Check that all cab requests have a corresponding elevator state
	for id := range r.Cr {
		if _, ok := r.States[id]; !ok {
			return false
		}
	}

	// Check that all elevator states have a corresponding cab request
	for id := range r.States {
		if _, ok := r.Cr[id]; !ok {
			return false
		}
	}

	return true
}
