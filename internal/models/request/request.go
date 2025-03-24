// Package request defines the data structures for representing elevator service requests.
//
// This package manages the representation of user button presses and their lifecycle
// throughout the distributed elevator system. It provides a state machine model for
// tracking requests from their initial unconfirmed state through to completion.
package request

import (
	"fmt"
)

//------------------------------------------------------------------------------
// Core Request Type
//------------------------------------------------------------------------------

// Request represents a user's request for elevator service at a specific location.
// Each request has an origin (where the button was pressed) and a status (its lifecycle state).
type Request struct {
	// Origin identifies the source of the request (which button was pressed)
	Origin Origin
	// Status represents the current state of the request in its lifecycle
	Status Status
}

//------------------------------------------------------------------------------
// Status
//------------------------------------------------------------------------------

// Status represents the lifecycle state of a request in the distributed system.
// A normal lifecycle for a request is as follows:
//  1. Unknown (initial state)
//  2. Absent (no active request exists)
//  3. Unconfirmed (request made by pressing a call button but not acknowledged)
//  4. Confirmed (request acknowledged by all peers and ready to be serviced)
//  5. Absent (request completed and no longer active)
type Status int

// Status constants define the possible lifecycle states of a request.
const (
	// Unknown indicates the request's state is not yet determined
	Unknown Status = iota
	// Absent indicates no active request exists for this origin
	Absent
	// Unconfirmed indicates a request has been made but not yet acknowledged by all peers
	Unconfirmed
	// Confirmed indicates the request has been acknowledged by all peers and is ready to be serviced
	Confirmed
)

//------------------------------------------------------------------------------
// String Representations
//------------------------------------------------------------------------------

// String returns a readable string representation of a Request.
func (r Request) String() string {
	return fmt.Sprintf("Request{Origin: %v, Status: %v}", r.Origin, r.Status)
}

// String returns a readable string representation of a request Status.
func (s Status) String() string {
	switch s {
	case Unknown:
		return "?"
	case Absent:
		return "A"
	case Unconfirmed:
		return "U"
	case Confirmed:
		return "C"
	default:
		return "?"
	}
}
