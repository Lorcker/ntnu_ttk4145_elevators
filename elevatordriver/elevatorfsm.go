package elevatordriver

import (
	"fmt"
	"log"

	"group48.ttk4145.ntnu/elevators/elevatorio"
	"group48.ttk4145.ntnu/elevators/models"
)

var EverybodyGoesOn bool = false

func onInitBetweenFloors() {
	if floor := elevatorio.GetFloor(); floor != 0 {
		elevatorio.SetMotorDirection(-1)
		for elevatorio.GetFloor() != 0 {
		}
		elevatorio.SetMotorDirection(0)
	}
}

func HandleOrderEvent(elevator *models.ElevatorState, orders models.Orders, recieverDoorTimer chan<- bool, resolvedRequests chan<- models.RequestMessage) {
	switch elevator.Behavior {
	case models.Idle:
		RequestChooseDirection(elevator, orders, recieverDoorTimer)
		if elevator.Behavior == models.DoorOpen {
			RequestClearAtCurrentFloor(*elevator, &orders, resolvedRequests)
		} else if elevator.Behavior == models.Moving {
			elevatorio.SetMotorDirection(elevator.Direction)
		}

	case models.DoorOpen:
		if RequestShouldClearImmediatly(*elevator, orders) {
			recieverDoorTimer <- true
			RequestClearAtCurrentFloor(*elevator, &orders, resolvedRequests)
		}
	}
}

func HandleFloorsensorEvent(elevator *models.ElevatorState, orders models.Orders, floor int, recieverDoorTimer chan<- bool, resolvedRequests chan<- models.RequestMessage) {
	elevator.Floor = floor
	elevatorio.SetFloorIndicator(floor)
	if elevator.Behavior == models.Moving && RequestShouldStop(*elevator, orders) {
		elevatorio.SetMotorDirection((0))
		elevatorio.SetDoorOpenLamp(true)
		RequestClearAtCurrentFloor(*elevator, &orders, resolvedRequests)
		recieverDoorTimer <- true
	}
}

func HandleRequestButtonEvent(elevator models.ElevatorState, button models.ButtonType) {

}

// When timer is done, close the door, and go in desired direction.
func HandleDoorTimerEvent(elevator *models.ElevatorState, orders models.Orders, recieverDoorTimer chan<- bool, resolvedRequests chan<- models.RequestMessage) {
	if (*elevator).Behavior == models.DoorOpen {
		RequestChooseDirection(elevator, orders, recieverDoorTimer)
		if elevator.Behavior == models.DoorOpen {
			recieverDoorTimer <- true
			RequestClearAtCurrentFloor(*elevator, &orders, resolvedRequests)
		} else {
			elevatorio.SetDoorOpenLamp(false)
			elevatorio.SetMotorDirection(elevator.Direction)
		}
	}
}

func OpenDoor(elevator *models.ElevatorState) {
	log.Printf("[elevatorfsm] Door open\n")
	elevatorio.SetDoorOpenLamp(true)
	elevator.Behavior = models.DoorOpen
}

func EmergencyStop(elevator *models.ElevatorState) {
	log.Printf("[elevatorfsm] Stop button not implemented :(\n")
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
	if e.Floor >= (int(models.NumFloors) - 1) {
		return false
	} //Already at top floor

	for i := (e.Floor + 1); i < int(models.NumFloors); i++ {
		for j := range len(orders[e.Floor]) {
			if orders[i][j] {
				return true
			}
		}
	}
	return false
}

func RequestHere(e models.ElevatorState, orders models.Orders) bool {
	for _, order := range orders[e.Floor] {
		if order {
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
		for j := range len(orders[e.Floor]) {
			if orders[i][j] {
				return true
			}
		}
	}
	return false

}

func RequestClearAtCurrentFloor(e models.ElevatorState, orders *models.Orders, resolvedRequests chan<- models.RequestMessage) {
	//Definisjon. True: Alle ordre skal fjernes fra etasjen (alle går på). False: Bare de i samme retning.
	if EverybodyGoesOn {
		for j := range len(orders[e.Floor]) {
			(*orders)[e.Floor][j] = false
			sendResolvedRequestsHallUp(e, resolvedRequests)
			sendResolvedRequestsHallDown(e, resolvedRequests)
			sendResolvedRequestsCabCall(e, resolvedRequests)
		}
	} else {
		(*orders)[e.Floor][models.Cab] = false
		sendResolvedRequestsCabCall(e, resolvedRequests)

		switch e.Direction {
		case models.Up:
			if !RequestAbove(e, (*orders)) && !(*orders)[e.Floor][models.HallUp] {
				(*orders)[e.Floor][models.HallDown] = false
				sendResolvedRequestsHallDown(e, resolvedRequests)
			}
			(*orders)[e.Floor][models.HallUp] = false
			sendResolvedRequestsHallUp(e, resolvedRequests)

		case models.Down:
			if !RequestBelow(e, (*orders)) && !(*orders)[e.Floor][models.HallDown] {
				(*orders)[e.Floor][models.HallUp] = false
				sendResolvedRequestsHallUp(e, resolvedRequests)
			}
			(*orders)[e.Floor][models.HallDown] = false
			sendResolvedRequestsHallDown(e, resolvedRequests)

		case models.Stop:
			fallthrough
		default:
			(*orders)[e.Floor][models.HallDown] = false
			(*orders)[e.Floor][models.HallUp] = false
			sendResolvedRequestsHallDown(e, resolvedRequests)
			sendResolvedRequestsHallUp(e, resolvedRequests)

		}

	}

}

func RequestShouldStop(e models.ElevatorState, orders models.Orders) bool {
	switch e.Direction {
	case models.Down:
		return orders[e.Floor][models.HallDown] || orders[e.Floor][models.Cab] || !RequestBelow(e, orders)
	case models.Up:
		return orders[e.Floor][models.HallUp] || orders[e.Floor][models.Cab] || !RequestAbove(e, orders)
	default:
		return true
	}
}

// Decision: Have to decide if everyone will get in the elevator, even tho they might be going in the opposite direction.

func RequestShouldClearImmediatly(e models.ElevatorState, orders models.Orders) bool {
	if EverybodyGoesOn {
		for i := range len(orders[e.Floor]) {
			if orders[e.Floor][i] {
				return true
			}
		}
		return false
	} else {
		switch e.Direction {
		case models.Up:
			return orders[e.Floor][models.HallUp]
		case models.Down:
			return orders[e.Floor][models.HallDown]
		case models.Stop:
			return orders[e.Floor][models.Cab]
		default:
			return false
		}
	}
}

// Debug functions.
func printOrders(orders models.Orders) {
	// Iterate through the outer slice (rows)
	log.Printf("Floor\t Up\t Down\t Cab\n")
	for i := range len(orders) {
		// Iterate through the inner slice (columns) at each row
		fmt.Printf("%d\t", i)
		for j := range len(orders[i]) {
			// Print the Order information
			fmt.Printf("%t\t ", orders[i][j])
		}
		fmt.Printf("\n\n")
	}
}

func printElevatorState(elevator models.ElevatorState) {
	log.Printf("[elevatorfsm]\n\nFloor: %d\n", elevator.Floor)
	log.Printf("Behavior: %d\n", elevator.Behavior)
	log.Printf("Direction: %d\n\n", elevator.Direction)
}

func initOrders(numFloors int) models.Orders {
	var orders models.Orders = models.Orders{}
	for i := range numFloors {
		for j := range 3 {
			orders[i][j] = false
		}
	}
	return orders
}

func sendResolvedRequestsHallUp(e models.ElevatorState, resolvedRequests chan<- models.RequestMessage) {
	o := models.Origin{Source: models.Hall{}, Floor: e.Floor, ButtonType: models.HallUp}
	r := models.Request{Origin: o, Status: models.Absent}
	resolvedRequests <- models.RequestMessage{Source: e.Id, Request: r}
	log.Printf("[Elevatorfsm] Resolved request HallUp from floor: %d", e.Floor)
}
func sendResolvedRequestsHallDown(e models.ElevatorState, resolvedRequests chan<- models.RequestMessage) {
	o := models.Origin{Source: models.Hall{}, Floor: e.Floor, ButtonType: models.HallDown}
	r := models.Request{Origin: o, Status: models.Absent}
	resolvedRequests <- models.RequestMessage{Source: e.Id, Request: r}
	log.Printf("[Elevatorfsm] Resolved request HallDown from floor: %d", e.Floor)

}
func sendResolvedRequestsCabCall(e models.ElevatorState, resolvedRequests chan<- models.RequestMessage) {
	o := models.Origin{Source: models.Elevator{Id: e.Id}, Floor: e.Floor, ButtonType: models.Cab}
	r := models.Request{Origin: o, Status: models.Absent}
	resolvedRequests <- models.RequestMessage{Source: e.Id, Request: r}
	log.Printf("[Elevatorfsm] Resolved cab request from floor: %d", e.Floor)

}
