package request

import (
	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
)

//------------------------------------------------------------------------------
// Origin Interface
//------------------------------------------------------------------------------

// Origin represents the source of a request (which button was pressed).
// This interface allows for different types of request sources (hall buttons, cab buttons).
type Origin interface {
	// isSource is a marker method to identify implementations of Origin
	isSource()
	// GetFloor returns the floor associated with this request origin
	GetFloor() elevator.Floor
	// GetButtonType returns the type of button that was pressed
	GetButtonType() elevator.ButtonType
}

//------------------------------------------------------------------------------
// Helper Functions
//------------------------------------------------------------------------------

// IsFromHall checks if a request originated from a hall call button.
func IsFromHall(r Request) bool {
	_, ok := r.Origin.(Hall)
	return ok
}

// IsCab checks if a request originated from within an elevator cab.
func IsCab(r Request) bool {
	_, ok := r.Origin.(Cab)
	return ok
}
