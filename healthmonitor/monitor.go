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
	var lastAlive = make([]models.Id, 0)

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
			a = append(a, local) // The local elevator is always alive
			if slicesEqual(lastAlive, a) {
				// Send no msg if new information is present
				continue
			}

			log.Printf("[healthmonitor] Notifying [orders] and [requests] that he alive list changed: %v", a)
			alivenessToOrders <- a
			alivenessToRequests <- a

			lastAlive = a
		}
	}
}

func getAlive(ls lastSeen) []models.Id {
	var a []models.Id
	for id, t := range ls {
		if time.Since(t) < Timeout {
			a = append(a, id)
		}
	}
	return a
}

func slicesEqual(a, b []models.Id) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
