package elevatordriver

import (
	"fmt"

	"group48.ttk4145.ntnu/elevators/elevatorio"
	"group48.ttk4145.ntnu/elevators/models"
)

var NButtons int = 3
var NFloors int = 4

func HandleOrderEvent(elevator *models.ElevatorState, orders models.Orders, recieverDoorTimer chan<- bool) {
	switch elevator.Behavior {
	case models.Idle:
		RequestChooseDirection(elevator, orders, recieverDoorTimer) // Updates the elevator states if new orders are in
		switch elevator.Behavior {
		case models.DoorOpen:
			//Start timer
			RequestClearAtCurrentFloor(*elevator, &orders)
			break

		case models.Moving:
			elevatorio.SetMotorDirection(elevator.Direction)
			break

		case models.Idle:
			break
		}

	case models.DoorOpen:
		//NB RequestShouldClearImmediatly not implementd!!//
		if RequestShouldClearImmediatly(*elevator, orders) {
			recieverDoorTimer <- true
			RequestClearAtCurrentFloor(*elevator, &orders)
		}
		break

	case models.Moving:
		break

	}
}

func HandleFloorsensorEvent(elevator *models.ElevatorState, orders models.Orders, floor int, recieverDoorTimer chan<- bool) {
	elevator.Floor = floor
	elevatorio.SetFloorIndicator(floor)
	switch elevator.Behavior {
	case models.Moving:
		if RequestShouldStop(*elevator, orders) {
			elevatorio.SetMotorDirection((0))
			elevatorio.SetDoorOpenLamp(true)
			RequestClearAtCurrentFloor(*elevator, &orders)
			setAllElevatorLights(orders)
			recieverDoorTimer <- true
		}
		break
	default:
		break
	}
}

func HandleRequestButtonEvent(elevator models.ElevatorState, button models.ButtonType) {

}

// When timer is done, close the door, and set in Idle/Moving.
func HandleDoorTimerEvent(elevator *models.ElevatorState, orders models.Orders, recieverDoorTimer chan<- bool) {
	//fmt.Printf("Dette fungerer!")
	switch elevator.Behavior {
	case models.DoorOpen:
		RequestChooseDirection(elevator, orders, recieverDoorTimer)

		switch elevator.Behavior {
		case models.DoorOpen:
			recieverDoorTimer <- true
			RequestClearAtCurrentFloor(*elevator, &orders)
			setAllElevatorLights(orders)
			break

		case models.Moving, models.Idle:
			elevatorio.SetDoorOpenLamp(false)
			elevatorio.SetMotorDirection(elevator.Direction)
			break
		}
		break

	default:
		break
	}
	// Remember to check for obstruction
}

func OpenDoor(elevator *models.ElevatorState) {
	fmt.Printf("Door open\n")
	elevatorio.SetDoorOpenLamp(true)
	elevator.Behavior = models.DoorOpen
}

// Little bit inspired by the given C-code :)
func RequestChooseDirection(e *models.ElevatorState, orders models.Orders, recieverDoorTimer chan<- bool) {
	switch e.Direction {
	case models.Up:
		if RequestAbove(*e, orders) {
			e.Direction = models.Up
			e.Behavior = models.Moving
		} else if RequestHere(*e, orders) {
			e.Direction = models.Stop
			recieverDoorTimer <- true

		} else if RequestBelow(*e, orders) {
			e.Direction = models.Down
			e.Behavior = models.Moving
		} else {
			e.Direction = models.Stop
			e.Behavior = models.Idle
		}

	case models.Down:
		if RequestBelow(*e, orders) {
			e.Direction = models.Down
			e.Behavior = models.Moving
		} else if RequestHere(*e, orders) {
			e.Direction = models.Stop
			recieverDoorTimer <- true
		} else if RequestAbove(*e, orders) {
			e.Direction = models.Up
			e.Behavior = models.Moving
		} else {
			e.Direction = models.Stop
			e.Behavior = models.Idle
		}

	case models.Stop:
		if RequestHere(*e, orders) {
			e.Direction = models.Stop
			recieverDoorTimer <- true
		} else if RequestAbove(*e, orders) {
			e.Direction = models.Up
			e.Behavior = models.Moving
		} else if RequestBelow(*e, orders) {
			e.Direction = models.Down
			e.Behavior = models.Moving
		} else {
			e.Direction = models.Stop
			e.Behavior = models.Idle
		}
	}
}

func RequestAbove(e models.ElevatorState, orders models.Orders) bool {
	if e.Floor >= (NFloors - 1) {
		return false
	} //Already at top floor

	for i := (e.Floor + 1); i < NFloors; i++ {
		for j := 0; j < NButtons; j++ {
			if orders[i][j] == true {
				return true
			}
		}
	}
	return false
}

func RequestHere(e models.ElevatorState, orders models.Orders) bool {
	for j := 0; j < NButtons; j++ {
		if orders[e.Floor][j] == true {
			return true
		}
	}
	return false
}

func RequestBelow(e models.ElevatorState, orders models.Orders) bool {
	if e.Floor == 0 {
		return false
	} // Already at bottom floor
	for i := e.Floor - 1; i >= 0; i-- {
		for j := 0; j < NButtons; j++ {
			if orders[i][j] == true {
				return true
			}
		}
	}
	return false

}

// Skriv ferdig denne?
func RequestClearAtCurrentFloor(e models.ElevatorState, orders *models.Orders) {
	clearRequestVariant := true //Definisjon. True: Alle ordre skal fjernes fra etasjen (alle går på). False: Bare de i samme retning.
	if clearRequestVariant {
		for j := 0; j < NButtons; j++ {
			(*orders)[e.Floor][j] = false
		}
	} else {
		switch e.Direction {
		case models.Up:
			if !RequestAbove(e, (*orders)) && !(*orders)[e.Floor][models.HallUp] {
				(*orders)[e.Floor][models.HallDown] = false
			}
			(*orders)[e.Floor][models.HallUp] = false
			break

		case models.Down:
			if !RequestBelow(e, (*orders)) && !(*orders)[e.Floor][models.HallDown] {
				(*orders)[e.Floor][models.HallUp] = false
			}
			(*orders)[e.Floor][models.HallDown] = false
			break

		case models.Stop:
		default:
			(*orders)[e.Floor][models.HallDown] = false
			(*orders)[e.Floor][models.HallUp] = false
			break
		}

	}

}

func RequestShouldStop(e models.ElevatorState, orders models.Orders) bool {
	switch e.Direction {
	case models.Down:
		if RequestHere(e, orders) || (!RequestBelow(e, orders)) {
			return true // Stop if no orders here, or below
		} else {
			return false
		}
	case models.Up:
		if RequestHere(e, orders) || (!RequestAbove(e, orders)) {
			return true
		} else {
			return false
		}
	case models.Stop:
		{
			return true
		}
	default:
		{
			return true
		}
	}
}

// Decision: Have to decide if everyone will get in the elevator, even tho they might be going in the opposite direction.

func RequestShouldClearImmediatly(e models.ElevatorState, orders models.Orders) bool {
	EverybodyGoesOn := true

	if EverybodyGoesOn {
		for i := 0; i < NButtons; i++ {
			if orders[e.Floor][i] {
				return true
			}
		}
		return false
	} else {
		switch e.Direction {
		case models.Up:
			if orders[e.Floor][models.HallUp] {
				return true
			} else {
				return false
			}

		case models.Down:
			if orders[e.Floor][models.HallDown] {
				return true
			} else {
				return false
			}

		case models.Stop:
			if orders[e.Floor][models.Cab] {
				return true
			} else {
				return false
			}
		default:
			return false
		}
	}
}

func setAllElevatorLights(orders models.Orders) {
	for i := 0; i < len(orders); i++ {
		for j := 0; j < len(orders[i]); j++ {
			if orders[i][j] == true {
				elevatorio.SetButtonLamp(models.ButtonType(j), i, true)
			} else {
				elevatorio.SetButtonLamp(models.ButtonType(j), i, false)
			}
		}
	}
}
