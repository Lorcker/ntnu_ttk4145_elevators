package driver

import "group48.ttk4145.ntnu/elevators/internal/models/elevator"

// ordersAbove returns true if there excist an order above
func (fsm *ElevatorFSM) ordersAbove() bool {
	if fsm.state.Floor >= elevator.NumFloors-1 {
		return false
	} //Already at top floor

	for i := (fsm.state.Floor + 1); i < elevator.NumFloors; i++ {
		for j := range fsm.orders[fsm.state.Floor] {
			if fsm.orders[i][j] {
				return true
			}
		}
	}
	return false
}

// ordersHere returns true if there excist an order here
func (fsm *ElevatorFSM) ordersHere() bool {
	for _, order := range fsm.orders[fsm.state.Floor] {
		if order {
			return true
		}
	}
	return false
}

// ordersBelow returns true if there excist an order below
func (fsm *ElevatorFSM) ordersBelow() bool {
	if fsm.state.Floor == 0 {
		return false
	} // Already at bottom floor
	for i := fsm.state.Floor - 1; i >= 0; i-- {
		for j := range len(fsm.orders[fsm.state.Floor]) {
			if fsm.orders[i][j] {
				return true
			}
		}
	}
	return false

}

// ordersClearAtCurrentFloor clears the orders that are excecuted
func (fsm *ElevatorFSM) ordersClearAtCurrentFloor() {
	// If EverybodyGoesOn is set to true, then all orders clear if the elevator stops at a floor.
	if EverybodyGoesOn {
		for j := range len((fsm.orders)[fsm.state.Floor]) {
			(fsm.orders)[fsm.state.Floor][j] = false
			fsm.rr(elevator.HallUp, fsm.state.Floor)
			fsm.rr(elevator.HallDown, fsm.state.Floor)
			fsm.rr(elevator.Cab, fsm.state.Floor)
		}
	} else {
		(fsm.orders)[fsm.state.Floor][elevator.Cab] = false //Cab orders are always cleared.
		fsm.rr(elevator.Cab, fsm.state.Floor)

		switch fsm.state.Direction {
		case elevator.Up:
			if !fsm.ordersAbove() && !(fsm.orders)[fsm.state.Floor][elevator.HallUp] { // Hall Down request is only cleared if it is not going further up
				(fsm.orders)[fsm.state.Floor][elevator.HallDown] = false
				fsm.rr(elevator.HallDown, fsm.state.Floor)
			}
			(fsm.orders)[fsm.state.Floor][elevator.HallUp] = false
			fsm.rr(elevator.HallUp, fsm.state.Floor)

		case elevator.Down:
			if !fsm.ordersBelow() && !(fsm.orders)[fsm.state.Floor][elevator.HallDown] { // Hall Up request is only cleared if it is not going further down
				(fsm.orders)[fsm.state.Floor][elevator.HallUp] = false
				fsm.rr(elevator.HallUp, fsm.state.Floor)
			}
			(fsm.orders)[fsm.state.Floor][elevator.HallDown] = false
			fsm.rr(elevator.HallDown, fsm.state.Floor)

		case elevator.Stop:
			fallthrough
		default:
			(fsm.orders)[fsm.state.Floor][elevator.HallDown] = false
			(fsm.orders)[fsm.state.Floor][elevator.HallUp] = false
			fsm.rr(elevator.HallDown, fsm.state.Floor)
			fsm.rr(elevator.HallUp, fsm.state.Floor)
		}

	}

}

// ordersElevatorShouldStop returns true if elevator should stop on that floor
func (fsm *ElevatorFSM) ordersElevatorShouldStop() bool {
	switch fsm.state.Direction {
	case elevator.Down:
		return fsm.orders[fsm.state.Floor][elevator.HallDown] || fsm.orders[fsm.state.Floor][elevator.Cab] || !fsm.ordersBelow()
	case elevator.Up:
		return fsm.orders[fsm.state.Floor][elevator.HallUp] || fsm.orders[fsm.state.Floor][elevator.Cab] || !fsm.ordersAbove()
	default:
		return true
	}
}

// ordersShouldClearImmediatly returns true if a order that comes in should be handled immediatly
func ordersShouldClearImmediatly(e elevator.State, orders elevator.Order) bool {
	if EverybodyGoesOn {
		for i := range len(orders[e.Floor]) {
			if orders[e.Floor][i] {
				return true
			}
		}
		return false
	} else {
		switch e.Direction {
		case elevator.Up:
			return orders[e.Floor][elevator.HallUp]
		case elevator.Down:
			return orders[e.Floor][elevator.HallDown]
		case elevator.Stop:
			return orders[e.Floor][elevator.Cab]
		default:
			return false
		}
	}
}
