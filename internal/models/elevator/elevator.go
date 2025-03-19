package elevator

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
