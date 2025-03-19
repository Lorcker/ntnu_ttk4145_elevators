package request

import (
	"fmt"

	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
)

type Request struct {
	Origin Origin
	Status Status
}

type Status int

const (
	Unknown Status = iota
	Absent
	Unconfirmed
	Confirmed
)

type Origin interface {
	isSource()
	GetFloor() elevator.Floor
}

type Hall struct {
	Direction Direction
	Floor     elevator.Floor
}

type Direction int

const (
	Up   Direction = 0
	Down Direction = 1
)

func (h Hall) GetFloor() elevator.Floor {
	return h.Floor
}

func (Hall) isSource() {}

type Cab struct {
	Id    elevator.Id
	Floor elevator.Floor
}

func (c Cab) GetFloor() elevator.Floor {
	return c.Floor
}

func (Cab) isSource() {}

func NewHallRequest(f elevator.Floor, dir Direction, s Status) Request {
	return Request{
		Origin: Hall{Floor: f, Direction: dir},
		Status: s,
	}
}

func NewCabRequest(f elevator.Floor, id elevator.Id, s Status) Request {
	return Request{
		Origin: Cab{Floor: f, Id: id},
		Status: s,
	}
}

func IsFromHall(r Request) bool {
	_, ok := r.Origin.(Hall)
	return ok
}

func IsCab(r Request) bool {
	_, ok := r.Origin.(Cab)
	return ok
}

// Implement the Stringer interface for Request
func (r Request) String() string {
	return fmt.Sprintf("Request{Origin: %v, Status: %v}", r.Origin, r.Status)
}

// Implement the Stringer interface for Hall
func (h Hall) String() string {
	return fmt.Sprintf("Hall{Floor: %v, Direction: %v}", h.Floor, h.Direction)
}

// Implement the Stringer interface for Cab
func (c Cab) String() string {
	return fmt.Sprintf("Cab{Id: %v, Floor: %v}", c.Id, c.Floor)
}

// Implement the Stringer interface for Status
func (s Status) String() string {
	switch s {
	case Unknown:
		return "Unknown"
	case Absent:
		return "Absent"
	case Unconfirmed:
		return "Unconfirmed"
	case Confirmed:
		return "Confirmed"
	default:
		return "Unknown"
	}
}

// Implement the Stringer interface for Direction
func (d Direction) String() string {
	switch d {
	case Up:
		return "Up"
	case Down:
		return "Down"
	default:
		return "Unknown"
	}
}
