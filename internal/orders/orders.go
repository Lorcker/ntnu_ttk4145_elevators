// orders the module responsible for managing the orders and distributing them to the elevators
package orders

import (
	"log"
	"reflect"
	"time"

	"group48.ttk4145.ntnu/elevators/internal/models"
)

// orderRefreshRate is the rate at which the order server will redistribute orders
// using the latest information from the cache
const orderRefreshRate = time.Millisecond * 2000

// RunOrderServer is the main function for the order module and should be run as a goroutine
//
// The server listens for validated requests, elevator states, alive status updates and
// stores them in a cache. The server then calculates the orders based on the cache and
// sends the local orders to the elevator driver.
func RunOrderServer(
	requestUpdate <-chan models.Request,
	stateUpdate <-chan models.ElevatorState,
	aliveListUpdate <-chan []models.Id,
	orderUpdates chan<- models.Orders,
	localPeerId models.Id) {

	// cache stores the latest requests, elevator states and alive information
	cache := newCache()
	// old orders stores the last calculated orders and is used to check if the orders have changed
	oldOrders := make(map[models.Id]models.Orders)
	// orderRefresh is a ticker that will trigger the order server to recalculate orders
	orderRefresh := time.NewTicker(orderRefreshRate)

	for {
		select {
		case r := <-requestUpdate:
			if r.Status == models.Unconfirmed || r.Status == models.Unknown {
				// These are not relevant for the order sever
				continue
			}

			if cache.AddRequest(r) {
				log.Printf("[orderserver] Request cache changed to:\n\tHallRequests: %v\n\tCabRequests: %v", cache.Hr, cache.Cr)
			}
		case a := <-aliveListUpdate:
			log.Printf("[orderserver] Received alive status: %v", a)
			cache.ProcessAliveUpdate(a)

		case newState := <-stateUpdate:
			if cache.AddElevatorState(newState) {
				log.Printf("[orderserver] Elevator state cache changed to: %v", cache.States)
			}

		case <-orderRefresh.C:
			if !cache.IsConsistent() {
				// The cache is not consistent, skip this iteration
				continue
			}

			newOrders := calculateOrders(cache.Hr, cache.Cr, cache.States)
			if reflect.DeepEqual(newOrders, oldOrders) {
				// Orders have not changed, no need to send an update to the elevator driver
				continue
			}

			log.Printf("[orderserver] Derived new orders:\n\tOld: %v\n\tNew: %v", oldOrders, newOrders)
			orderUpdates <- newOrders[localPeerId]

			oldOrders = newOrders
		}

	}
}
