package models

type Floor uint8

var NumFloors Floor = 4

type ElevatorState struct {
	Id        Id
	Floor     int
	Behavior  ElevatorBehavior
	Direction MotorDirection
}

type ElevatorBehavior int

const (
	Idle ElevatorBehavior = iota
	DoorOpen
	Moving
)

// Orders is a 2D array of bools, where the first dimension is the floor and the second dimension is the button type.
type Orders [][3]bool

type MotorDirection int

const (
	Up   MotorDirection = 1
	Down                = -1
	Stop                = 0
)

type ButtonType int

const (
	HallUp   ButtonType = 0
	HallDown            = 1
	Cab                 = 2
)

func IsEStateEqual(a, b ElevatorState) bool {
	return a.Id == b.Id && a.Floor == b.Floor && a.Behavior == b.Behavior && a.Direction == b.Direction
}
