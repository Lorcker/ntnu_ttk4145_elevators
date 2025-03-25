// Package message defines the communication data structures used between system modules.
//
// This package contains all message types used for inter-module communication within
// the elevator system. Each message type serves a specific purpose in the system's
// operation and follows defined flow patterns between modules.
package message

import (
	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/request"
)

// FloorArrival is a message sent when an elevator arrives at or passes a floor.
//
// Flow path: [elevio] -> [driver]
type FloorArrival struct {
	// Floor indicates which floor the elevator has arrived at
	Floor elevator.Floor
}

// Obstruction is a message sent when the elevator's obstruction switch is toggled.
// This switch typically detects if something is blocking the door.
//
// Flow path: [elevio] -> [driver]
type Obstruction struct{}

// ElevatorState is a message sent when the operational state of an elevator changes.
// This includes changes in floor position, behavior mode, or movement direction.
//
// Flow paths:
//   - [driver] -> [comms]  (local elevator state updates sent to peers)
//   - [driver] -> [orders] (local elevator state updates for order calculation)
//   - [comms] -> [orders]  (external elevator state updates received from peers)
type ElevatorState struct {
	// Elevator identifies which elevator's state has changed
	Elevator elevator.Id
	Alive    bool
	// State contains the updated operational state information
	State elevator.State
}

// ServiceOrder is a message sent when new service orders have been calculated for an elevator.
//
// Flow path: [orders] -> [driver]
type ServiceOrder struct {
	// Order contains the calculated service orders for the elevator
	Order elevator.Order
}

// RequestState is a message sent when the lifecycle state of a service request changes.
// This includes new requests, confirmed requests, and completed requests.
//
// Flow paths: Between [requests] and [comms] modules in both directions
type RequestState struct {
	// Source identifies which elevator initiated this update
	Source elevator.Id
	// Request contains the updated request information
	Request request.Request
}

// PeerSignal is a message sent when communication is received from another elevator.
// It serves as proof that another elevator in the system is operational.
//
// Flow path: [comms] -> [healthmonitor]
type PeerSignal struct {
	// Id identifies which elevator sent the heartbeat
	Id elevator.Id
	// Alive indicates whether the elevator is still
	// operational and sending heartbeats
	Alive bool
}

// ActivePeers is a message sent when the set of operational elevators changes.
// This includes both new elevators joining and existing elevators becoming unresponsive.
//
// Flow paths:
//   - [healthmonitor] -> [requests] (for managing request acknowledgments)
//   - [healthmonitor] -> [orders]   (for calculating optimal order assignments)
type ActivePeers struct {
	// Peers contains the IDs of all elevators currently known to be operational
	Peers []elevator.Id
}
