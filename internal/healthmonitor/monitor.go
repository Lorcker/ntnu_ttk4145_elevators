// Description: This package is responsible for monitoring the health of the elevators.
// It listens for pings from the elevators and tracks which elevators are alive.
// It also informs the orders and requests modules about the alive elevators.
// The RunMonitor function runs the health monitor.
package healthmonitor

import (
	"log"
	"reflect"
	"time"

	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/message"
)

// Timeout is the time after which an elevator is considered dead.
const Timeout = time.Second * 10

// PollInterval is the frequency at which the monitor informs about alive elevators.
const PollInterval = time.Second * 1

// lastSeen is a map of the last time a ping was received from an elevator.
type lastSeen = map[elevator.Id]time.Time

// RunMonitor runs the health monitor
//
// It listens for pings from the elevators and tracks which elevators are alive.
func RunMonitor(
	local elevator.Id,
	pingFromComms <-chan message.PeerHeartbeat,
	alivenessToRequests chan<- message.AlivePeersUpdate,
	alivenessToOrders chan<- message.AlivePeersUpdate) {

	var lastSeen = make(lastSeen)
	var alivePeers = make(map[elevator.Id]bool)

	ticker := time.NewTicker(PollInterval)

	for {
		select {
		case msg := <-pingFromComms:
			if _, ok := lastSeen[msg.Id]; !ok {
				log.Printf("[healthmonitor] A new pear is alive %v", msg.Id)
			}
			lastSeen[msg.Id] = time.Now()
		case <-ticker.C:
			a := getAlive(lastSeen)
			a[local] = true // Always consider yourself alive

			if reflect.DeepEqual(a, alivePeers) {
				// No need to notify if the list of alive peers is the same
				continue
			}

			alivePeers = a

			log.Printf("[healthmonitor] Notifying [orders] and [requests] that he alive list changed: %v", a)
			msg := message.AlivePeersUpdate{
				Peers: mapToSlice(alivePeers),
			}

			alivenessToOrders <- msg
			alivenessToRequests <- msg
		}
	}
}

// getAlive returns a map of the alive elevators
//
// An elevator is considered alive if a ping has been received from it within the last Timeout.
func getAlive(ls lastSeen) map[elevator.Id]bool {
	var a = make(map[elevator.Id]bool)
	for id, t := range ls {
		if time.Since(t) < Timeout {
			a[id] = true
		}
	}
	return a
}

// mapToSlice converts a map to a slice
func mapToSlice(m map[elevator.Id]bool) []elevator.Id {
	s := make([]elevator.Id, 0, len(m))
	for id := range m {
		s = append(s, id)
	}
	return s
}
