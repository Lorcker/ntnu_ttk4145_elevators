package request

import (
	"fmt"
	"log"

	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
)

//------------------------------------------------------------------------------
// Hall Origin Type
//------------------------------------------------------------------------------

// Hall represents a request originating from a hall call button.
type Hall struct {
	// Direction indicates whether the hall call is for upward or downward travel
	Direction Direction
	// Floor indicates which floor the hall call button was pressed on
	Floor elevator.Floor
}

// Direction represents the intended direction of travel for a hall call.
type Direction int

//------------------------------------------------------------------------------
// Direction Constants
//------------------------------------------------------------------------------

// Direction constants define the possible travel directions for hall calls.
const (
	// Up indicates the user wants to travel upward
	Up Direction = 0
	// Down indicates the user wants to travel downward
	Down Direction = 1
)

//------------------------------------------------------------------------------
// Hall Origin Methods
//------------------------------------------------------------------------------

// isSource implements the Origin interface.
func (Hall) isSource() {}

// GetFloor returns the floor where this hall request originated.
func (h Hall) GetFloor() elevator.Floor {
	return h.Floor
}

// GetButtonType returns the button type for this hall request.
func (h Hall) GetButtonType() elevator.ButtonType {
	switch h.Direction {
	case Up:
		return elevator.HallUp
	case Down:
		return elevator.HallDown
	default:
		log.Fatal("Hall direction of the origin has an illegal value")
		return elevator.HallUp // Default to avoid compilation error
	}
}

// String returns a readable string representation of a Hall request origin.
func (h Hall) String() string {
	return fmt.Sprintf("Hall{Floor: %v, Direction: %v}", h.Floor, h.Direction)
}

// String returns a readable string representation of a Direction.
func (d Direction) String() string {
	switch d {
	case Up:
		return "U"
	case Down:
		return "D"
	default:
		return "?"
	}
}

//------------------------------------------------------------------------------
// Factory Functions
//------------------------------------------------------------------------------

// NewHallRequest creates a new Request for a hall call.
func NewHallRequest(f elevator.Floor, dir Direction, s Status) Request {
	return Request{
		Origin: Hall{Floor: f, Direction: dir},
		Status: s,
	}
}
