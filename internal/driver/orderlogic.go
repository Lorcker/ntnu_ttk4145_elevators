package driver

import "group48.ttk4145.ntnu/elevators/internal/models/elevator"

func orderAbove(e elevator.State, orders elevator.Order) bool {
	if e.Floor >= elevator.NumFloors-1 {
		return false
	} //Already at top floor

	for i := (e.Floor + 1); i < elevator.NumFloors; i++ {
		for j := range orders[e.Floor] {
			if orders[i][j] {
				return true
			}
		}
	}
	return false
}

func orderHere(e elevator.State, orders elevator.Order) bool {
	for _, order := range orders[e.Floor] {
		if order {
			return true
		}
	}
	return false
}

func orderBelow(e elevator.State, orders elevator.Order) bool {
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

// orderClearAtCurrentFloor clears the orders that are excecuted
func orderClearAtCurrentFloor(e elevator.State, orders *elevator.Order, rr resolvedRequests) {
	// If EverybodyGoesOn is set to true, then all orders clear if the elevator stops at a floor.
	if EverybodyGoesOn {
		for j := range len((*orders)[e.Floor]) {
			(*orders)[e.Floor][j] = false
			rr(elevator.HallUp, e.Floor)
			rr(elevator.HallDown, e.Floor)
			rr(elevator.Cab, e.Floor)
		}
	} else {
		(*orders)[e.Floor][elevator.Cab] = false //Cab orders are always cleared.
		rr(elevator.Cab, e.Floor)

		switch e.Direction {
		case elevator.Up:
			if !orderAbove(e, (*orders)) && !(*orders)[e.Floor][elevator.HallUp] { // Hall Down request is only cleared if it is not going further up
				(*orders)[e.Floor][elevator.HallDown] = false
				rr(elevator.HallDown, e.Floor)
			}
			(*orders)[e.Floor][elevator.HallUp] = false
			rr(elevator.HallUp, e.Floor)

		case elevator.Down:
			if !orderBelow(e, (*orders)) && !(*orders)[e.Floor][elevator.HallDown] { // Hall Up request is only cleared if it is not going further down
				(*orders)[e.Floor][elevator.HallUp] = false
				rr(elevator.HallUp, e.Floor)
			}
			(*orders)[e.Floor][elevator.HallDown] = false
			rr(elevator.HallDown, e.Floor)

		case elevator.Stop:
			fallthrough
		default:
			(*orders)[e.Floor][elevator.HallDown] = false
			(*orders)[e.Floor][elevator.HallUp] = false
			rr(elevator.HallDown, e.Floor)
			rr(elevator.HallUp, e.Floor)
		}

	}

}

// orderElevatorShouldStop returns true if elevator should stop on that floor
func orderElevatorShouldStop(e elevator.State, orders elevator.Order) bool {
	switch e.Direction {
	case elevator.Down:
		return orders[e.Floor][elevator.HallDown] || orders[e.Floor][elevator.Cab] || !orderBelow(e, orders)
	case elevator.Up:
		return orders[e.Floor][elevator.HallUp] || orders[e.Floor][elevator.Cab] || !orderAbove(e, orders)
	default:
		return true
	}
}

// orderShouldClearImmediatly returns true if a order that comes in should be handled immediatly
func orderShouldClearImmediatly(e elevator.State, orders elevator.Order) bool {
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
