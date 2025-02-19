package communicationmodule

import "group48.ttk4145.ntnu/elevators/models"

func RunComms(
	estates <-chan models.ElevatorState,
	request <-chan models.Request,
	estatesOut chan<- models.ElevatorState,
	ping chan<- models.Id) {
}
