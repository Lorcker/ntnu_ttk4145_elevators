package healthmonitor

import (
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
	ping <-chan models.Id,
	alive chan<- []models.Id) {

	var lastSeen = make(lastSeen)
	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

	for {
		select {
		case id := <-ping:
			lastSeen[id] = time.Now()
		case <-ticker.C:
			alive <- getAlive(lastSeen)
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
