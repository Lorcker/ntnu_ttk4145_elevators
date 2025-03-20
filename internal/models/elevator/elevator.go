package elevator

import "fmt"

type Floor int

const NumFloors Floor = 4

type Id uint8

type State struct {
	Floor     Floor
	Behavior  Behavior
	Direction MotorDirection
}

type Behavior int

const (
	Idle Behavior = iota
	DoorOpen
	Moving
)

type MotorDirection int

const (
	Up   MotorDirection = 1
	Down MotorDirection = -1
	Stop MotorDirection = 0
)

type ButtonType int

const (
	HallUp   ButtonType = 0
	HallDown ButtonType = 1
	Cab      ButtonType = 2
)

// Order represents the orders of the elevator
//
// An order is created when a request was verified by the request module and
// the order module has assigned the order to an elevator.
//
// The first index represents the floor of the order.
// The second index represents the button type of the order.
type Order = [NumFloors][3]bool

// orderToString converts an order to a string representation using 1s and 0s
// for true and false respectively.
// Floors are vertically stacked
func OrderToString(o Order) string {
	str := ""
	for i := 0; i < int(NumFloors); i++ {
		str += "["
		for j := 0; j < 3; j++ {
			if o[i][j] {
				str += "1"
			} else {
				str += "0"
			}
		}
		str += "] "
	}
	return str
}

// String returns a string representation of the elevator state
func (s State) String() string {
	return fmt.Sprintf("Floor: %d, Behavior: %v, Direction: %v", s.Floor, s.Behavior, s.Direction)
}

// DiffString returns a string representation of diff between two elevator states
func (s State) DiffString(s2 State) string {
	return fmt.Sprintf("Floor: %d -> %d, Behavior: %v -> %v, Direction: %v -> %v", s.Floor, s2.Floor, s.Behavior, s2.Behavior, s.Direction, s2.Direction)
}

// String returns a string representation of the elevator Behavior
func (b Behavior) String() string {
	switch b {
	case Idle:
		return "Idle"
	case DoorOpen:
		return "DoorOpen"
	case Moving:
		return "Moving"
	default:
		return "Unknown"
	}
}

// String returns a string representation of the elevator MotorDirection
func (d MotorDirection) String() string {
	switch d {
	case Up:
		return "Up"
	case Down:
		return "Down"
	case Stop:
		return "Stop"
	default:
		return "Unknown"
	}
}

// String returns a string representation of the elevator ButtonType
func (b ButtonType) String() string {
	switch b {
	case HallUp:
		return "HallUp"
	case HallDown:
		return "HallDown"
	case Cab:
		return "Cab"
	default:
		return "Unknown"
	}
}
