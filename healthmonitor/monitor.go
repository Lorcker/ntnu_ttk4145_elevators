package healthmonitor

import "group48.ttk4145.ntnu/elevators/models"

func RunMonitor(
	ping <-chan models.Id,
	alive chan<- []models.Id) {
}
