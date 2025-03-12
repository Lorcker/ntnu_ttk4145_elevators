package healthmonitor

import (
	"log"
	"time"

	"group48.ttk4145.ntnu/elevators/models"
)

// Timeout is the time after which an elevator is considered dead.
const Timeout = time.Second * 10

// PollInterval is the frequency at which the monitor informs about alive elevators.
const PollInterval = time.Second * 1

type lastSeen = map[models.Id]time.Time

// RunMonitor runs the health monitor. It listens for pings from the elevators
// and tracks which elevators are alive.
func RunMonitor(
	local models.Id,
	pingFromComms <-chan models.Id,
	alivenessToRequests chan<- []models.Id,
	alivenessToOrders chan<- []models.Id) {

	var lastSeen = make(lastSeen)
	var alivePeers = make(map[models.Id]bool)

	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

	for {
		select {
		case id := <-pingFromComms:
			if _, ok := lastSeen[id]; !ok {
				log.Printf("[healthmonitor] A new pear is alive %v", id)
			}
			lastSeen[id] = time.Now()
		case <-ticker.C:
			a := getAlive(lastSeen)
			a[local] = true
			if !isDifferent(a, alivePeers) {
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

func getAlive(ls lastSeen) map[models.Id]bool {
	var a = make(map[models.Id]bool)
	for id, t := range ls {
		if time.Since(t) < Timeout {
			a[id] = true
		}
	}
	return a
}

func isDifferent(a, b map[models.Id]bool) bool {
	if len(a) != len(b) {
		return true
	}
	for id := range a {
		if _, ok := b[id]; !ok {
			return true
		}
	}
	return false
}

func mapToSlice(m map[models.Id]bool) []models.Id {
	s := make([]models.Id, 0, len(m))
	for id := range m {
		s = append(s, id)
	}
	return s
}
