// Description: This package is responsible for monitoring the health of the elevators.
// It listens for pings from the elevators and tracks which elevators are alive.
// It also informs the orders and requests modules about the alive elevators.
// The RunMonitor function runs the health monitor.
package healthmonitor

import (
	"log"
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

// alivePeers is a map of the alive elevators.
type alivePeers = map[elevator.Id]bool

// RunMonitor runs the health monitor
//
// It listens for pings from the elevators and tracks which elevators are alive.
func RunMonitor(
	local elevator.Id,
	pingFromComms <-chan message.PeerSignal,
	alivenessToRequests chan<- message.ActivePeers,
	alivenessToOrders chan<- message.ActivePeers) {

	lastSeen := make(lastSeen)
	alivePeers := make(alivePeers)

	ticker := time.NewTicker(PollInterval)

	localAlive := true

	for {
		select {
		case msg := <-pingFromComms:
			processPing(msg, lastSeen)

		case <-ticker.C:
			if !updateAliveList(lastSeen, alivePeers, local, localAlive) {
				continue
			}
			msg := message.ActivePeers{
				Peers: mapToSlice(alivePeers),
			}
			log.Printf("[healthmonitor] Alive peers: %v", msg.Peers)
			alivenessToOrders <- msg
			alivenessToRequests <- msg
		}
	}
}

func processPing(msg message.PeerSignal, lastSeen lastSeen) {
	if msg.Alive {
		if _, ok := lastSeen[msg.Id]; !ok {
			log.Printf("[healthmonitor] A new peer with id %v is alive", msg.Id)
		}
		lastSeen[msg.Id] = time.Now()
	} else if !msg.Alive {
		lastSeen[msg.Id] = time.Now().Add(-Timeout)
	}
}

func updateAliveList(lastSeen lastSeen, alivePeers alivePeers, local elevator.Id, localAlive bool) bool {
	changed := false
	for id, t := range lastSeen {
		if time.Since(t) < Timeout {
			if !alivePeers[id] {
				alivePeers[id] = true
				changed = true
			}
		} else if alivePeers[id] {
			delete(alivePeers, id)
			changed = true
			log.Printf("[healthmonitor] The Peer with id %v has died", id)
		}
	}
	if localAlive && !alivePeers[local] {
		alivePeers[local] = true
		changed = true
	} else if !localAlive && alivePeers[local] {
		log.Printf("[healthmonitor] The local peer has died")
		delete(alivePeers, local)
		changed = true
	}

	return changed
}

// mapToSlice converts a map to a slice
func mapToSlice(m map[elevator.Id]bool) []elevator.Id {
	s := make([]elevator.Id, 0, len(m))
	for id := range m {
		s = append(s, id)
	}
	return s
}
