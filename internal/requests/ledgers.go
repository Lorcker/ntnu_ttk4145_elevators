package requests

import (
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
	lm.ledgers[origin][id] = true
}

// resetLedgers resets the ledgers for the origin.
func (lm *ledgerTracker) resetLedgers(origin request.Origin) {
	lm.ledgers[origin] = make(map[elevator.Id]bool)
}

// haveAllAlivePeersAcknowledged checks if all alive peers have acknowledged the request.
func (lm *ledgerTracker) haveAllAlivePeersAcknowledged(o request.Origin, alive []elevator.Id) bool {
	if len(lm.ledgers[o]) != len(alive) {
		return false
	}

	for _, id := range alive {
		if _, ok := lm.ledgers[o][id]; !ok {
			return false
		}
	}

	return true
}
