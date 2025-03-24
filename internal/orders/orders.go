// orders the module responsible for managing the orders and distributing them to the elevators
package orders

import (
	"fmt"
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
	requestUpdate <-chan message.RequestState,
	stateUpdate <-chan message.ElevatorState,
	aliveListUpdate <-chan message.ActivePeers,
	orderUpdates chan<- message.ServiceOrder,
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
			isUnRelevant := msg.Request.Status == request.Unconfirmed || msg.Request.Status == request.Unknown
			if isUnRelevant {
				continue
			}
			cache.AddRequest(msg.Request)

		case msg := <-aliveListUpdate:
			cache.ProcessAliveUpdate(msg.Peers)

		case msg := <-stateUpdate:
			cache.AddElevatorState(msg.Elevator, msg.State)

		case <-orderRefresh.C:
			if !cache.IsConsistent() {
				continue
			}

			newOrders := calculateOrders(cache.Hr, cache.Cr, cache.States)
			if reflect.DeepEqual(newOrders, oldOrders) {
				// Orders have not changed, no need to send an update to the elevator driver
				continue
			}

			logChangedOrders(oldOrders, newOrders)
			orderUpdates <- message.ServiceOrder{
				Order: newOrders[localPeerId],
			}

			oldOrders = newOrders
		}

	}
}

// logChangedOrders logs only the orders that have changed
func logChangedOrders(oldOrders, newOrders map[elevator.Id]elevator.Order) {
	msg := "[orderserver] Orders changed for elevators:\n"
	for id, newOrder := range newOrders {
		msg += fmt.Sprintf("\tElevator %v: ", id)
		if oldOrder, ok := oldOrders[id]; ok {
			msg += fmt.Sprintf("%v -> ", elevator.OrderToString(oldOrder))
		}
		msg += fmt.Sprintf("%v\n", elevator.OrderToString(newOrder))
	}
	log.Print(msg)
}
