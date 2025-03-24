package request

import (
	"fmt"

	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
)

//------------------------------------------------------------------------------
// Cab Origin Type
//------------------------------------------------------------------------------

// Cab represents a request originating from within an elevator cab.
type Cab struct {
	// Id identifies which elevator this cab request belongs to
	Id elevator.Id
	// Floor indicates the destination floor selected
	Floor elevator.Floor
}

//------------------------------------------------------------------------------
// Cab Origin Methods
//------------------------------------------------------------------------------

// isSource implements the Origin interface.
func (Cab) isSource() {}

// GetFloor returns the destination floor for this cab request.
func (c Cab) GetFloor() elevator.Floor {
	return c.Floor
}

// GetButtonType returns the button type for cab requests.
func (c Cab) GetButtonType() elevator.ButtonType {
	return elevator.Cab
}

// String returns a readable string representation of a Cab request origin.
func (c Cab) String() string {
	return fmt.Sprintf("Cab{Id: %v, Floor: %v}", c.Id, c.Floor)
}

//------------------------------------------------------------------------------
// Factory Functions
//------------------------------------------------------------------------------

// NewCabRequest creates a new Request for a cab button press.
func NewCabRequest(f elevator.Floor, id elevator.Id, s Status) Request {
	return Request{
		Origin: Cab{Floor: f, Id: id},
		Status: s,
	}
}
