package orderserver

import (
	"group48.ttk4145.ntnu/elevators/models"
)

func RunOrderServer(
	validatedRequests <-chan models.Request,
	state <-chan models.Elevator,
	alive <-chan []uint8,
	orders chan<- models.Orders) {
}
