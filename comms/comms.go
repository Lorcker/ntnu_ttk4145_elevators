package comms

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"Network-go/network/bcast"

	"group48.ttk4145.ntnu/elevators/models"
	m "group48.ttk4145.ntnu/elevators/models"
)

const SendInterval = time.Millisecond * 100

type udpMessage struct {
	Source   m.Id
	Registry RequestRegistry
	EState   m.ElevatorState
}

// # RunComms runs the communication module
//
// It listens for updates on the local elevator state and validated requests channels.
// It send UDP messages with the local elevator state and all system requests to the broadcast address in a regular interval.
// It listens for incoming UDP messages and sends the elevator state and changed requests to the outgoing channels.
// It sends a health monitor ping on the health monitor ping channel when it receives an update from the local elevator state or validated requests channels.
func RunComms(
	local m.Id,
	port int,
	fromDriver <-chan m.ElevatorState,
	fromRequests <-chan m.Request,
	toOrders chan<- m.ElevatorState,
	toRequest chan<- m.RequestMessage,
	toHealthMonitor chan<- m.Id) {

	var sendTicker = time.NewTicker(SendInterval)
	var internalEs m.ElevatorState
	var registry = NewRequestRegistry()

	sendUdp := make(chan udpMessage)
	receiveUdp := make(chan udpMessage)
	go bcast.Transmitter(port, sendUdp)
	go bcast.Receiver(port, receiveUdp)

	for {
		select {
		case es := <-fromDriver:
			if !models.IsEStateEqual(internalEs, es) {
				log.Printf("[comms] Received new local elevator state update from [driver]: %v", es)
				internalEs = es
			}

		case r := <-fromRequests:
			before := fmt.Sprintf("%v", registry)

			registry.Update(r)

			after := fmt.Sprintf("%v", registry)
			if before != after {
				log.Printf("[comms] Updated registry after getting msg from [requests]:\n\tRequest: %v\n\tBefore: %v\n\tAfter: %v", r, before, after)
			}

		case <-sendTicker.C:
			if internalEs == (m.ElevatorState{}) {
				continue
			}
			u := udpMessage{
				Source:   local,
				Registry: registry,
				EState:   internalEs,
			}
			sendUdp <- u

		case msg := <-receiveUdp:
			if msg.Source == local {
				continue
			}

			toHealthMonitor <- msg.Source
			toOrders <- msg.EState

			changedRequests := registry.Diff(msg.Source, msg.Registry)
			if len(changedRequests) > 0 {
				log.Printf("[comms] Received an external registry that changed state:\n\tFromPeer: %d\n\tChangedReqs: %v\n\tInternalRegistry:%v\n\tExternalRegistry:%v", msg.Source, changedRequests, registry, msg.Registry)
			}
			for _, r := range changedRequests {
				toRequest <- r
			}
		}
	}

}

// The RequestRegistry holds information about all system requests.
// It is needed as every sending cycle of comms must propagate all system request to other peers.
// The internal system work with sending messaged on change.
// This does not work for comms as packet loss is guaranteed to happen so changes might get lost.
// Thus the registry stores the change and can calculate the diff between two registries to
// enable the conversion back to the internal messaging model.
// Also, in case one elevator dies, the information is backed up here.
type RequestRegistry struct {
	HallUp   [m.NumFloors]m.RequestStatus
	HallDown [m.NumFloors]m.RequestStatus

	// Map uses the id of the elevator as key
	// Is a string because the json conversion of network module only allows for strings
	Cab map[string][m.NumFloors]m.RequestStatus
}

func NewRequestRegistry() RequestRegistry {
	hu := [m.NumFloors]m.RequestStatus{}
	hd := [m.NumFloors]m.RequestStatus{}
	c := make(map[string][m.NumFloors]m.RequestStatus)

	for i := m.Floor(0); i < m.NumFloors; i++ {
		hu[i] = m.Unknown
		hd[i] = m.Unknown
	}

	return RequestRegistry{
		HallUp:   hu,
		HallDown: hd,
		Cab:      c,
	}
}

// Adds a new cab to the registry
func (r *RequestRegistry) InitNewCab(id string) {
	cab := [m.NumFloors]m.RequestStatus{}
	for i := m.Floor(0); i < m.NumFloors; i++ {
		cab[i] = m.Unknown
	}

	r.Cab[id] = cab
}

// Update takes in a internal msg from the request module and replaces the stored information
// As the msg were validated by the request module no checks on the status information are needed
func (r *RequestRegistry) Update(req m.Request) {
	floor := req.Origin.Floor

	switch req.Origin.ButtonType {
	case m.HallUp:
		r.HallUp[floor] = req.Status
	case m.HallDown:
		r.HallDown[floor] = req.Status
	case m.Cab:
		id := req.Origin.Source.(m.Elevator).Id
		idS := strconv.Itoa(int(id))

		// Check is needed because if comms get info about an elevator it has not seen before
		// it need to adds to the registry and keep it there
		if _, ok := r.Cab[idS]; !ok {
			r.InitNewCab(idS)
		}

		// Reassign updated slice to map as no direct update is possible in Go
		cabRequests := r.Cab[idS]
		cabRequests[floor] = req.Status
		r.Cab[idS] = cabRequests
	}
}

// Diff calculates the difference between two registries
// and returns a slice of requestMessage where each represents a differing entry
// If both states are Unconfirmed the entry is also included to enable acknoledgement of the request
func (r *RequestRegistry) Diff(peer m.Id, other RequestRegistry) []m.RequestMessage {
	var diff []m.RequestMessage

	for f := m.Floor(0); f < m.NumFloors; f++ {
		if isDifferent(r.HallUp[f], other.HallUp[f]) {
			diff = append(diff, m.NewHallRequestMsg(peer, int(f), m.HallUp, other.HallUp[f]))
		}
		if isDifferent(r.HallDown[f], other.HallDown[f]) {
			diff = append(diff, m.NewHallRequestMsg(peer, int(f), m.HallDown, other.HallDown[f]))
		}
	}

	for id, otherCab := range other.Cab {
		localCab, ok := r.Cab[id]
		idI, err := strconv.Atoi(id)

		if err != nil {
			log.Fatalf("[comms] failed to convert stored elevator id string to its uint: %e", err)
		}

		if !ok {
			for f := m.Floor(0); f < m.NumFloors; f++ {
				diff = append(diff, m.NewCabRequestMsg(peer, m.Id(idI), int(f), otherCab[f]))
			}
			continue
		}

		for f := m.Floor(0); f < m.NumFloors; f++ {
			if isDifferent(localCab[f], otherCab[f]) {
				diff = append(diff, m.NewCabRequestMsg(peer, m.Id(idI), int(f), otherCab[f]))
			}
		}
	}

	return diff
}

// isDifferent checks if two request status are different
// If both are Unconfirmed the function returns true to enable acknoledgement of the request
func isDifferent(a, b m.RequestStatus) bool {
	if a == m.Unconfirmed && b == m.Unconfirmed {
		return true
	}
	return a != b
}
