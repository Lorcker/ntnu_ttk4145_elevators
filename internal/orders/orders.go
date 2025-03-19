// orders the module responsible for managing the orders and distributing them to the elevators
package orders

import (
	"log"
	"reflect"
	"time"

	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/message"
	"group48.ttk4145.ntnu/elevators/internal/models/request"
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
	localPeerId elevator.Id,
	requestUpdate <-chan message.RequestStateUpdate,
	stateUpdate <-chan message.ElevatorStateUpdate,
	aliveListUpdate <-chan message.AlivePeersUpdate,
	orderUpdates chan<- message.Order,
) {

	// cache stores the latest requests, elevator states and alive information
	cache := newCache()
	// old orders stores the last calculated orders and is used to check if the orders have changed
	oldOrders := make(map[elevator.Id]elevator.Order)
	// orderRefresh is a ticker that will trigger the order server to recalculate orders
	orderRefresh := time.NewTicker(orderRefreshRate)

	for {
		select {
		case msg := <-requestUpdate:
			if msg.Request.Status == request.Unconfirmed || msg.Request.Status == request.Unknown {
				// These are not relevant for the order sever
				continue
			}

			if cache.AddRequest(msg.Request) {
				log.Printf("[orderserver] Request cache changed to:\n\tHallRequests: %v\n\tCabRequests: %v", cache.Hr, cache.Cr)
			}
		case msg := <-aliveListUpdate:
			log.Printf("[orderserver] Received alive status: %v", msg.Peers)
			cache.ProcessAliveUpdate(msg.Peers)

		case msg := <-stateUpdate:
			if cache.AddElevatorState(msg.Elevator, msg.State) {
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
			orderUpdates <- message.Order{
				Order: newOrders[localPeerId],
			}

			oldOrders = newOrders
		}

	}
}
