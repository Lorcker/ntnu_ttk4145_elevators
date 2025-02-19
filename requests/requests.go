package requests

import (
	m "group48.ttk4145.ntnu/elevators/models"
)

func RunRequestServer(
	incomingRequests <-chan m.RequestMessage,
	peerStatus <-chan []m.Id,
	subscribers []chan<- map[m.Origin]m.Request) {

	var requestManager = newRequestManager()

	for {
		select {
		case msg := <-incomingRequests:
			requestManager.process(msg)
			for _, s := range subscribers {
				s <- requestManager.store
			}
		case alivePeers := <-peerStatus:
			requestManager.alivePeers = alivePeers
		}
	}
}

type requestManager struct {
	store      map[m.Origin]m.Request
	ledgers    map[m.Origin][]m.Id
	alivePeers []m.Id
}

func newRequestManager() *requestManager {
	return &requestManager{
		store:      make(map[m.Origin]m.Request),
		ledgers:    make(map[m.Origin][]m.Id),
		alivePeers: make([]m.Id, 0),
	}
}

func (rm *requestManager) process(msg m.RequestMessage) {
	_, ok := rm.store[msg.Request.Origin]
	if !ok {
		rm.store[msg.Request.Origin] = msg.Request
		return
	}

	switch msg.Request.Status {
	case m.Absent:
		rm.processAbsent(msg)
	case m.Unconfirmed:
		rm.processUnconfirmed(msg)
	case m.Confirmed:
		rm.processConfirmed(msg)
	case m.Unknown:
		break
	}
}

func (rm *requestManager) processAbsent(msg m.RequestMessage) {
	if msg.Request.Status != m.Absent {
		return
	}

	storedRequest := rm.store[msg.Request.Origin]
	switch storedRequest.Status {
	case m.Confirmed:
		fallthrough
	case m.Unknown:
		storedRequest.Status = m.Absent
	case m.Absent:
	case m.Unconfirmed:
	}

	rm.store[msg.Request.Origin] = storedRequest
}

func (rm *requestManager) processUnconfirmed(msg m.RequestMessage) {
	if msg.Request.Status != m.Unconfirmed {
		return
	}

	ledgers := rm.ledgers[msg.Request.Origin]
	storedRequest := rm.store[msg.Request.Origin]
	switch storedRequest.Status {
	case m.Unknown:
		fallthrough
	case m.Absent:
		fallthrough
	case m.Unconfirmed:
		ledgers = append(ledgers, msg.Source)

		isConfirmed := isSetEqual(ledgers, rm.alivePeers)
		if isConfirmed {
			storedRequest.Status = m.Confirmed
			ledgers = make([]m.Id, 0)
		} else {
			storedRequest.Status = m.Unconfirmed
		}
	case m.Confirmed:
	}

	rm.ledgers[msg.Request.Origin] = ledgers
	rm.store[msg.Request.Origin] = storedRequest
}

func (rm *requestManager) processConfirmed(msg m.RequestMessage) {
	if msg.Request.Status != m.Confirmed {
		return
	}

	storedRequest := rm.store[msg.Request.Origin]
	storedRequest.Status = m.Confirmed

	rm.store[msg.Request.Origin] = storedRequest
}

func isSetEqual(a, b []m.Id) bool {
	if len(a) != len(b) {
		return false
	}

	for _, id := range a {
		if !contains(b, id) {
			return false
		}
	}

	return true
}

func contains(s []m.Id, e m.Id) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
