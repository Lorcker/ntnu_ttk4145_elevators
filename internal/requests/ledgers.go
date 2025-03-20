package requests

import (
	"log"

	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/request"
)

// ledgerTracker keeps track of which peers have acknowledged a request
//
// It is needed to move requests from Unconfirmed to Confirmed state and is
// used by the requestManager.
type ledgerTracker struct {
	// ledgers are associated with an origin and keep track of which peers have acknowledged a request.
	// Acknowledgment is indicated by the presence of the peer id in the nested map and the value being true.
	ledgers map[request.Origin]map[elevator.Id]bool
}

// newLedgerManager creates a new ledger tracker
func newLedgerManager() *ledgerTracker {
	return &ledgerTracker{
		ledgers: make(map[request.Origin]map[elevator.Id]bool),
	}
}

// addLedger adds a ledger for the origin and id.
func (lm *ledgerTracker) addLedger(origin request.Origin, id elevator.Id) {
	if _, ok := lm.ledgers[origin]; !ok {
		lm.ledgers[origin] = make(map[elevator.Id]bool)
	}

	if lm.ledgers[origin][id] {
		// The ledger already exists, do not add it again
		return
	}

	lm.ledgers[origin][id] = true
	log.Printf("[requests] [manager] [ledgers] Added ledger for %v with id %v", origin, id)
}

// resetLedgers resets the ledgers for the origin.
func (lm *ledgerTracker) resetLedgers(origin request.Origin) {
	lm.ledgers[origin] = make(map[elevator.Id]bool)
	log.Printf("[requests] [manager] [ledgers] Reset ledgers for %v", origin)
}

// isMessageAcknowledged checks if all alive peers have acknowledged the request.
//
// A request is acknowledged if one of the following conditions is met:
//   - A hall request is acknowledged by all alive peers and at least one other peer.
//   - A cab request is acknowledged by all alive peers (can be only the local elevator).
func (lm *ledgerTracker) isMessageAcknowledged(o request.Origin, alive []elevator.Id) bool {
	if len(lm.ledgers[o]) != len(alive) {
		return false
	}

	if _, ok := o.(request.Hall); ok && len(lm.ledgers[o]) < 2 {
		// Hall requests must be acknowledged by all alive peers and at least one other peer.
		// This is because of the button light contract.
		// When the local elevator is disconnected, the no redundancy would be present.
		return false
	}

	for _, id := range alive {
		if _, ok := lm.ledgers[o][id]; !ok {
			return false
		}
	}

	return true
}
