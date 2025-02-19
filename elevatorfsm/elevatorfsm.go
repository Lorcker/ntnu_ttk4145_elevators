package elevatorfsm

import (
	"group48.ttk4145.ntnu/elevators/elevatorio"
	"group48.ttk4145.ntnu/elevators/orderserver"
)

type Elevator struct {
	Floor        int
	Behavior     ElevatorBehavior
	HallRequests [][]bool
	CabRequests  []bool
	Direction    elevatorio.MotorDirection
}

type ElevatorBehavior int

const (
	EB_Idle ElevatorBehavior = iota
	EB_DoorOpen
	EB_Moving
)

func HandleOrderEvent(elevator Elevator, orders orderserver.Orders) {
	switch elevator.Behavior {
	case EB_Idle:
		RequestChooseDirection(&e, orders)	// Updates the elevator states if new orders are in
		switch elevator.Behavior {
		case EB_DoorOpen:
			//Start timer
			RequestClearAtCurrentFloor(e, orders)
			break
		
		case EB_Moving:
			elevator.SetMotorDirection(elevator.Direction)
			break
		
		case EB_Idle:
			break

	case EB_DoorOpen:
		//NB RequestShouldClearImmediatly not implementd!!//
		if RequestShouldClearImmediatly(Elevaror e, orders orderserver.Orders) {
			RequestClearAtCurrentFloor(Elevator e, orders orderserver.Orders)
		}
		break
	

	case EB_Moving:
		break
		
	}
}
}

func HandleFloorsensorEvent(elevator Elevator, floor int) {

}

func HandleRequestButtonEvent(elevator Elevator, button elevatorio.ButtonEvent) {

}

func HandleDoorTimerEvent(elevator Elevator, timer bool) {
	// Remember to check for obstruction
}



// Little bit inspired by the given C-code :)
func RequestChooseDirection(Elevator* e, orders orderserver.Orders) { 
	switch e.Direction{
	case MD_Up:
		if 		RequestAbove(*e, orders) 		{*e.Direction = MD_Up; *e.Behavior=EB_Moving}
		else if RequestHere(*e, orders) 	{*e.Direction = MD_Stop; *e.Behavior=EB_DoorOpen} // In given C-code they write Direction Down
		else if RequestBelow(*e, orders) 	{*e.Direction = MD_Down; *e.Behavior=EB_Moving}
		else 								{*e.Direction = MD_Stop; *e.Behavior=EB_Idle}
	}
	case MD_Down:
		if 		RequestBelow(*e, orders) 	{*e.Direction = MD_Down; *e.Behavior=EB_Moving}
		else if RequestHere(*e, orders) 	{*e.Direction = MD_Stop; *e.Behavior=EB_DoorOpen}
		else if RequestAbove(*e, orders) 	{*e.Direction = MD_Up; *e.Behavior=EB_Moving}
		else 								{*e.Direction = MD_Stop; *e.Behavior=EB_Idle}
	}
	case MD_Stop:
		if 			RequestHere(*e, orders) 	{*e.Direction = MD_Stop; *e.Behavior=EB_DoorOpen}
		else if 	RequestAbove(*e, orders) 	{*e.Direction = MD_Up; *e.Behavior=EB_Moving}
		else if  	RequestBelow(*e, orders) 	{*e.Direction = MD_Down; *e.Behavior=EB_Moving}
		else 									{*e.Direction = MD_Stop; *e.Behavior=EB_Idle}


func RequestAbove(Elevator e, orders orderserver.Orders) int {
	if e.floor == Nfloors {return 0}	//Already at top floor

	for (int i=e.floor+1; i<Nfloors; i++) {
		for (int j=0; j<NButtons; j++) {
			if orders[i][j] == 1 {
				return 1
			}
		}
	}
	return 0
}

func RequestHere(Elevator e, orders orderserver.Orders) int {
	for (int j=0; j<NButtons; j++) {
		if (orders[e.floor][j] == 1) {
			return 1
		}
	}
	return 0
}

func RequestBelow(Elevator e, orders orderserver.Orders) int {
	if (e.floor == 0) {return 0} 			// Already at bottom floor
	for (int i=e.floor-1; i>=0; i--) {
		for (int j=0; j<NButtons; j++) {
			if orders[i][j] == 1 {
				return 1
			}
		}
	}
	return 0

}

func RequestClearAtCurrentFloor(Elevator e, orders orderserver.Orders) {
	for (int j=0; j<NButtons; j++) {
		orders[e.floor][j] = 0
	}
}

//Finish this later
func RequestShouldClearImmediatly(Elevaror e, orders orderserver.Orders) {

}

// func HandleOrderEvent(elevator Elevator, orders orderserver.Orders) {
// 	int closest_order_f = inf
// 	int current_order_floor = 0
// 	int current_order_distance = 0 
// 	Nfloors = 4; NButtons = 3;

// 	switch Elevator.Behavior {
// 	case EB_Idle:
// 		//Find out which floor the order is coming from
// 		for (int i=0; i<Nfloors*NButtons; i++) {
// 			if (orders[i] == 1) {
// 				current_order_floor = i % Nfloors
// 				current_order_distance = Elevator.Floor - current_order_floor
// 			}
// 		}
// 		continue
	
// 	case EB_DoorOpen:
// 		continue
	
// 	case EB_Moving:
// 		continue
	
	
// 	}
// }