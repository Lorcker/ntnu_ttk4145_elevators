// Description: This package is responsible for monitoring the health of the elevators.
// It listens for pings from the elevators and tracks which elevators are alive.
// It also informs the orders and requests modules about the alive elevators.
// The RunMonitor function runs the health monitor.
package healthmonitor

import (
	"log"
	"reflect"
	"time"

	"group48.ttk4145.ntnu/elevators/internal/models"
)

// Timeout is the time after which an elevator is considered dead.
const Timeout = time.Second * 10

// PollInterval is the frequency at which the monitor informs about alive elevators.
const PollInterval = time.Second * 1

// lastSeen is a map of the last time a ping was received from an elevator.
type lastSeen = map[models.Id]time.Time

// RunMonitor runs the health monitor
//
// It listens for pings from the elevators and tracks which elevators are alive.
func RunMonitor(
	local models.Id,
	pingFromComms <-chan models.Id,
	alivenessToRequests chan<- []models.Id,
	alivenessToOrders chan<- []models.Id) {

	var lastSeen = make(lastSeen)
	var alivePeers = make(map[models.Id]bool)

	ticker := time.NewTicker(PollInterval)

	for {
		select {
		case id := <-pingFromComms:
			if _, ok := lastSeen[id]; !ok {
				log.Printf("[healthmonitor] A new pear is alive %v", id)
			}
			lastSeen[id] = time.Now()
		case <-ticker.C:
			a := getAlive(lastSeen)
			a[local] = true // Always consider yourself alive

			if reflect.DeepEqual(a, alivePeers) {
				// No need to notify if the list of alive peers is the same
				continue
			}

			alivePeers = a

			log.Printf("[healthmonitor] Notifying [orders] and [requests] that he alive list changed: %v", a)
			s := mapToSlice(alivePeers)
			alivenessToOrders <- s
			alivenessToRequests <- s
		}
	}
}

// getAlive returns a map of the alive elevators
//
// An elevator is considered alive if a ping has been received from it within the last Timeout.
func getAlive(ls lastSeen) map[models.Id]bool {
	var a = make(map[models.Id]bool)
	for id, t := range ls {
		if time.Since(t) < Timeout {
			a[id] = true
		}
	}
	return a
}

// mapToSlice converts a map to a slice
func mapToSlice(m map[models.Id]bool) []models.Id {
	s := make([]models.Id, 0, len(m))
	for id := range m {
		s = append(s, id)
	}
	return s
}
