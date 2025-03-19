package message

import (
	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/request"
)

// FloorSensor is a message that is sent when the elevator reaches a new floor
//
// Module Flows:
//
//	[elevio] -> [driver]
type FloorSensor struct {
	Floor elevator.Floor
}

// ObstructionSwitch is a message that is sent when the obstruction switch is toggled
//
// Module Flows:
//
//	[elevio] -> [driver]
type ObstructionSwitch struct{}

// ElevatorStateUpdate is a message that is sent when the state of an elevator changes
//
// Module Flows:
//
//	[driver] -> [comms]
//	[driver] -> [orders]
//	[comms] -> [order]
type ElevatorStateUpdate struct {
	Elevator elevator.Id
	State    elevator.State
}

// Order is a message that is sent when new orders are calculated
//
// Module Flows:
//
//	[orders] -> [driver]
type Order struct {
	Order elevator.Order
}

// RequestStateUpdate is a message that is sent when the state of a request changes
//
// Module Flows:
//
//	[request] <-> [comms]
type RequestStateUpdate struct {
	Source  elevator.Id
	Request request.Request
}

// PeerHeartbeat is a message that is sent when a message is received from a peer
//
// Module Flows:
//
//	[comms] -> [healthmonitor]
type PeerHeartbeat struct {
	Id elevator.Id
}

// AlivePeersUpdate is a message that is sent when the list of alive peers changes
//
// Module Flows:
//
//	[healthmonitor] -> [requests]
//	[healthmonitor] -> [orders]
type AlivePeersUpdate struct {
	Peers []elevator.Id
}
