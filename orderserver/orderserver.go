package orderserver

import (
	"group48.ttk4145.ntnu/elevators/models"
)

func RunOrderServer(
	validatedRequests <-chan models.Request,
	state <-chan models.ElevatorState,
	alive <-chan []models.Id,
	orders chan<- models.Orders) {
}
