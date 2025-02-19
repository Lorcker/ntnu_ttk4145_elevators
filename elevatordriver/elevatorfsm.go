package elevatordriver

import (
	"group48.ttk4145.ntnu/elevators/models"
)

func HandleOrderEvent(elevator models.ElevatorState, orders models.Orders) {}

func HandleFloorsensorEvent(elevator models.ElevatorState, floor int) {}

func HandleDoorTimerEvent(elevator models.ElevatorState, timer bool) {
	// Remember to check for obstruction
}
