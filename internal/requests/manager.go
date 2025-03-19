package requests

import (
	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/message"
	"group48.ttk4145.ntnu/elevators/internal/models/request"
)

// requestManager manages the state of requests.
//
// It implements a simple state machine for processing requests.
// The states of a request are:
//   - Unknown: The request is not present and the state is unknown.
//   - Absent: The request is not present.
//   - Unconfirmed: The request is present, but not all peers have acknowledged it.
//   - Confirmed: The request is present and all peers have acknowledged it.
//
// Generally, the transitions are cyclic, similar to a cyclic counter. The cycle is:
//
//	Absent -> Unconfirmed -> Confirmed -> Absent
//
// The possible transitions given an input are implemented in the process method.
type requestManager struct {
	// local id is needed to add the local elevator to the ledgers.
	local elevator.Id

	// store contains the latest state of each request.
	store map[request.Origin]request.Request

	// ledgerTracker keeps track of which peers have acknowledged a request.
	// It is needed to move requests from Unconfirmed to Confirmed state.
	ledgerTracker *ledgerTracker

	// alivePeers is used to determine if all alive peers have acknowledged a request.
	// This is needed to move requests from Unconfirmed to Confirmed state.
	alivePeers []elevator.Id
}

// newRequestManager creates a new request manager
func newRequestManager(local elevator.Id) *requestManager {
	return &requestManager{
		local:         local,
		store:         make(map[request.Origin]request.Request),
		ledgerTracker: newLedgerManager(),
		alivePeers:    make([]elevator.Id, 0),
	}
}

// process processes a request message and returns the updated request.
//
// Processed requests are stored in the request manager to keep track of the state of each request.
func (rm *requestManager) process(msg message.RequestStateUpdate) request.Request {
	if _, ok := rm.store[msg.Request.Origin]; !ok {
		rm.store[msg.Request.Origin] = msg.Request
	}

	switch msg.Request.Status {
	case request.Absent:
		return rm.processAbsent(msg)
	case request.Unconfirmed:
		return rm.processUnconfirmed(msg)
	case request.Confirmed:
		return rm.processConfirmed(msg)
	default:
		return rm.processUnknown(msg)
	}

}

// processUnknown processes a request with an Unknown status.
func (rm *requestManager) processUnknown(msg message.RequestStateUpdate) request.Request {
	// As a request with an Unknown status does not add any new information,
	// the stored request is returned as is.
	return rm.store[msg.Request.Origin]
}

// processAbsent processes a request with an Absent status.
func (rm *requestManager) processAbsent(msg message.RequestStateUpdate) request.Request {
	if msg.Request.Status != request.Absent {
		return msg.Request
	}

	storedRequest := rm.store[msg.Request.Origin]
	if storedRequest.Status == request.Confirmed || storedRequest.Status == request.Unknown {
		// Acknowledgement from the other peers is not needed,
		// as a request could only have been confirmed if all peers acknowledged it in the first place
		storedRequest.Status = request.Absent
	}

	rm.store[msg.Request.Origin] = storedRequest
	return storedRequest
}

// processUnconfirmed processes a request with an Unconfirmed status.
func (rm *requestManager) processUnconfirmed(msg message.RequestStateUpdate) request.Request {
	if msg.Request.Status != request.Unconfirmed {
		return msg.Request
	}

	storedRequest := rm.store[msg.Request.Origin]
	if storedRequest.Status == request.Confirmed {
		// The stored version is already confirmed, so we return it as is
		return storedRequest
	}

	rm.ledgerTracker.addLedger(msg.Request.Origin, msg.Source)
	rm.ledgerTracker.addLedger(msg.Request.Origin, rm.local)

	if rm.ledgerTracker.isMessageAcknowledged(msg.Request.Origin, rm.alivePeers) {
		storedRequest.Status = request.Confirmed

		// Ledgers are reset as the next time the request reaches the Unconfirmed state,
		// it must be acknowledged by all peers again.
		rm.ledgerTracker.resetLedgers(msg.Request.Origin)
	} else {
		storedRequest.Status = request.Unconfirmed
	}

	rm.store[msg.Request.Origin] = storedRequest
	return storedRequest
}

// processConfirmed processes a request with a Confirmed status.
func (rm *requestManager) processConfirmed(msg message.RequestStateUpdate) request.Request {
	if msg.Request.Status != request.Confirmed {
		return msg.Request
	}

	storedRequest := rm.store[msg.Request.Origin]

	if storedRequest.Status != request.Unconfirmed {
		// Either the request is already confirmed locally, so we can return it as is,
		// or the stored request is absent. In the latter case, we should not change the status,
		// this means that an elevator has cleared the request, and the request should not be re-added.
		return storedRequest
	}

	// If the stored request is currently Unconfirmed, the request is updated to Confirmed.
	// This is okay, as the request could only have been confirmed if all peers acknowledged it in the first place.
	storedRequest.Status = request.Confirmed
	rm.store[msg.Request.Origin] = storedRequest
	return storedRequest
}
