package orderserver

import (
	"log"
	"time"

	"group48.ttk4145.ntnu/elevators/internal/models"
)

const orderRefreshRate = time.Millisecond * 2000

func RunOrderServer(
	validatedRequests <-chan models.Request,
	state <-chan models.ElevatorState,
	alive <-chan []models.Id,
	orders chan<- models.Orders,
	localPeerId models.Id) {

	cache := newCache()
	orderRefresh := time.NewTicker(orderRefreshRate)

	//init local vars
	for {
		select {
		case r := <-validatedRequests:
			if r.Status == models.Unconfirmed || r.Status == models.Unknown {
				// These are not relevant for the order sever
				continue
			}

			if cache.AddRequest(r) {
				log.Printf("[orderserver] Request cache changed to:\n\tHallRequests: %v\n\tCabRequests: %v", cache.Hr, cache.Cr)
			}
		case a := <-alive:
			log.Printf("[orderserver] Received alive status: %v", a)
			cache.ProcessAliveUpdate(a)

		case newState := <-state:
			if cache.AddElevatorState(newState) {
				log.Printf("[orderserver] Elevator state cache changed to: %v", cache.Elevators)
			}

		case <-orderRefresh.C:
			if !cache.IsConsistent() {
				continue
			}
			os := calculateOrders(cache.Hr, cache.Cr, cache.Elevators)
			log.Printf("[orderserver] Calculated new orders: %v", os)

			order := os[localPeerId]
			orders <- order
		}

	}
}

type cache struct {
	Hr         HallRequests
	Cr         map[models.Id]CabRequests
	Elevators  map[models.Id]models.ElevatorState
	AlivePeers map[models.Id]bool
}

func newCache() *cache {
	return &cache{
		Hr:         make(HallRequests, models.NumFloors),
		Cr:         make(map[models.Id]CabRequests),
		Elevators:  make(map[models.Id]models.ElevatorState),
		AlivePeers: make(map[models.Id]bool),
	}
}

func (r *cache) AddRequest(req models.Request) (didChange bool) {
	status := req.Status == models.Confirmed
	if _, ok := req.Origin.Source.(models.Hall); ok {
		return r.addHallRequest(req.Origin.Floor, req.Origin.ButtonType, status)
	}

	// Must be a cab request
	return r.addCabRequest(req.Origin.Source.(models.Elevator).Id, req.Origin.Floor, status)
}

func (r *cache) addHallRequest(floor int, direction models.ButtonType, status bool) (didChange bool) {
	hr := r.Hr
	oldStatus := hr[floor][direction]
	hr[floor][direction] = status
	r.Hr = hr
	return oldStatus != status
}

func (r *cache) addCabRequest(id models.Id, floor int, status bool) (didChange bool) {
	cr, ok := r.Cr[id]
	if !ok {
		cr = make(CabRequests, models.NumFloors)
	}
	oldStatus := cr[floor]
	cr[floor] = status
	r.Cr[id] = cr
	return oldStatus != status
}

func (r *cache) AddElevatorState(state models.ElevatorState) (didChange bool) {
	oldState, ok := r.Elevators[state.Id]
	if !ok {
		r.Elevators[state.Id] = state
		r.Cr[state.Id] = make(CabRequests, models.NumFloors) // Initialize cab requests for the new elevator
		return true
	}

	if oldState == state {
		return false
	}

	r.Elevators[state.Id] = state
	return true
}

func (r *cache) ProcessAliveUpdate(alive []models.Id) {
	newAlive := make(map[models.Id]bool)
	for _, id := range alive {
		newAlive[id] = true
		r.AlivePeers[id] = true
	}

	// Check if a peer died
	for id := range r.AlivePeers {
		if _, ok := newAlive[id]; !ok {
			// peer died - remove the peer from the states and alivePeers map to exclude it from the calculations
			delete(r.Elevators, id)
			delete(r.Cr, id)
			delete(r.AlivePeers, id)
		}
	}
}

func (r *cache) IsConsistent() bool {
	// Check that all alive elevators have cab requests and elevator states in the cache
	for id := range r.AlivePeers {
		if _, ok := r.Cr[id]; !ok {
			return false
		}
		if _, ok := r.Elevators[id]; !ok {
			return false
		}
	}

	// Check that all cab requests have a corresponding elevator state
	for id := range r.Cr {
		if _, ok := r.Elevators[id]; !ok {
			return false
		}
	}

	// Check that all elevator states have a corresponding cab request
	for id := range r.Elevators {
		if _, ok := r.Cr[id]; !ok {
			return false
		}
	}

	return true
}
