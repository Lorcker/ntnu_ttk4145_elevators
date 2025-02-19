package requests

import "group48.ttk4145.ntnu/elevators/models"

func RunRequestServer(
	unValidatedRequest <-chan models.RequestMessage,
	aliveStatus <-chan []int,
	validatedRequests chan<- []models.Request) {
}
