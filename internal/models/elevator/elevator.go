// Package elevator defines the core data structures for representing elevator state and behavior.
//
// This package contains all the types and constants related to physical elevator properties,
// including floor information, movement state, and button types. These models are used
// throughout the system for consistent representation of elevator data.
package elevator

import "fmt"

// Floor represents a physical floor number in the building.
type Floor int

// NumFloors is the total number of floors the elevator can access.
const NumFloors Floor = 4

// Id uniquely identifies an elevator in the distributed system.
type Id uint8

// State represents the current operational state of an elevator.
// It combines information about the elevator's current floor,
// behavior mode, and movement direction.
type State struct {
	// Floor is the current floor position of the elevator
	Floor Floor
	// Behavior represents what the elevator is currently doing (idle, moving, etc.)
	Behavior Behavior
	// Direction indicates the current movement direction of the elevator
	Direction MotorDirection
}

// Behavior defines the operational mode of the elevator.
type Behavior int

// Behavior constants define the possible operational states of an elevator.
const (
	// Idle indicates the elevator is stationary with closed doors, waiting for commands
	Idle Behavior = iota
	// DoorOpen indicates the elevator is stationary with doors open
	DoorOpen
	// Moving indicates the elevator is in motion between floors
	Moving
)

// MotorDirection defines the direction of movement for the elevator motor.
type MotorDirection int

// MotorDirection constants define the possible movement directions.
const (
	// Up indicates upward movement (+1)
	Up MotorDirection = 1
	// Down indicates downward movement (-1)
	Down MotorDirection = -1
	// Stop indicates no movement (0)
	Stop MotorDirection = 0
)

// ButtonType identifies the different types of elevator call buttons.
type ButtonType int

// ButtonType constants define the types of buttons in the elevator system.
const (
	// HallUp is a hall call button requesting upward travel
	HallUp ButtonType = 0
	// HallDown is a hall call button requesting downward travel
	HallDown ButtonType = 1
	// Cab is an internal elevator button for selecting a destination floor
	Cab ButtonType = 2
)

// Order represents the service orders assigned to an elevator.
//
// An order is created when a request has been verified by the request module and
// the order module has assigned the order to a specific elevator.
//
// The first index represents the floor of the order.
// The second index represents the button type of the order.
// A true value indicates the elevator should service that request.
type Order = [NumFloors][3]bool

// OrderToString converts an order to a readable string representation using 1s and 0s
// for true and false respectively.
// Floors are vertically stacked in the output string.
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

// String returns a readable string representation of the elevator state.
func (s State) String() string {
	return fmt.Sprintf("Floor: %d, Behavior: %v, Direction: %v", s.Floor, s.Behavior, s.Direction)
}

// DiffString returns a string showing the differences between this state and another.
func (s State) DiffString(s2 State) string {
	return fmt.Sprintf("Floor: %d -> %d, Behavior: %v -> %v, Direction: %v -> %v",
		s.Floor, s2.Floor, s.Behavior, s2.Behavior, s.Direction, s2.Direction)
}

// String returns a readable string representation of the elevator Behavior.
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

// String returns a readable string representation of the elevator MotorDirection.
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

// String returns a readable string representation of the elevator ButtonType.
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
