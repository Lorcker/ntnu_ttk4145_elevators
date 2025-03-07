package requests

import (
	"log"

	"group48.ttk4145.ntnu/elevators/elevatorio"
	m "group48.ttk4145.ntnu/elevators/models"
)

func RunRequestServer(
	local m.Id,
	incomingRequests <-chan m.RequestMessage,
	peerStatus <-chan []m.Id,
	subscribers []chan<- m.Request) {

	var requestManager = newRequestManager(local)

	for {
		select {
		case msg := <-incomingRequests:
			log.Printf("[requests] Received request: %v", msg.Request)
			r := requestManager.process(msg)
			setButtonLighting(local, msg.Request)
			log.Printf("[requests] Processed request: %v", r)
			for _, s := range subscribers {
				s <- r
			}
		case alivePeers := <-peerStatus:
			log.Printf("[requests] Received alive peers: %v", alivePeers)

			requestManager.alivePeers = alivePeers
		}
	}
}

func setButtonLighting(local m.Id, req m.Request) {
	if elevator, ok := req.Origin.Source.(m.Elevator); ok && elevator.Id != local {
		return // Lighting does not concern this elevator
	}
	targetState := req.Status == m.Confirmed || !(req.Status == m.Absent)

	elevatorio.SetButtonLamp(req.Origin.ButtonType, req.Origin.Floor, targetState)
}

type requestManager struct {
	local m.Id

	store   map[m.Origin]m.Request
	ledgers map[m.Origin]map[m.Id]bool

	alivePeers []m.Id
}

func newRequestManager(local m.Id) *requestManager {
	return &requestManager{
		local:      local,
		store:      make(map[m.Origin]m.Request),
		ledgers:    make(map[m.Origin]map[m.Id]bool),
		alivePeers: make([]m.Id, 0),
	}
}

func (rm *requestManager) process(msg m.RequestMessage) m.Request {
	if _, ok := rm.store[msg.Request.Origin]; !ok {
		rm.store[msg.Request.Origin] = msg.Request
	}

	switch msg.Request.Status {
	case m.Absent:
		return rm.processAbsent(msg)
	case m.Unconfirmed:
		return rm.processUnconfirmed(msg)
	case m.Confirmed:
		return rm.processConfirmed(msg)
	case m.Unknown:
		fallthrough
	default:
		return rm.processUnknown(msg)
	}

}

func (rm *requestManager) processUnknown(msg m.RequestMessage) m.Request {
	if msg.Request.Status != m.Unknown {
		return msg.Request
	}

	storedRequest := rm.store[msg.Request.Origin]

	return storedRequest
}

func (rm *requestManager) processAbsent(msg m.RequestMessage) m.Request {
	if msg.Request.Status != m.Absent {
		return msg.Request
	}

	storedRequest := rm.store[msg.Request.Origin]
	if storedRequest.Status == m.Confirmed || storedRequest.Status == m.Unknown {
		storedRequest.Status = m.Absent
	}

	rm.store[msg.Request.Origin] = storedRequest
	return storedRequest
}

func (rm *requestManager) processUnconfirmed(msg m.RequestMessage) m.Request {
	if msg.Request.Status != m.Unconfirmed {
		return msg.Request
	}

	ledgers, ok := rm.ledgers[msg.Request.Origin]
	if !ok {
		// Origin was not stored before and needs be initialized
		ledgers = make(map[m.Id]bool)
		rm.ledgers[msg.Request.Origin] = ledgers
	}

	storedRequest := rm.store[msg.Request.Origin]

	switch storedRequest.Status {
	case m.Unknown:
		fallthrough
	case m.Absent:
		fallthrough
	case m.Unconfirmed:
		ledgers[rm.local] = true
		ledgers[msg.Source] = true

		isConfirmed := isConfirmed(ledgers, rm.alivePeers)
		if isConfirmed {
			storedRequest.Status = m.Confirmed
			// Reset ledgers
			ledgers = make(map[m.Id]bool)
		} else {
			storedRequest.Status = m.Unconfirmed
		}
	}

	rm.ledgers[msg.Request.Origin] = ledgers
	rm.store[msg.Request.Origin] = storedRequest
	return storedRequest
}

func (rm *requestManager) processConfirmed(msg m.RequestMessage) m.Request {
	if msg.Request.Status != m.Confirmed {
		return msg.Request
	}

	storedRequest := rm.store[msg.Request.Origin]

	if storedRequest.Status != m.Unconfirmed {
		return storedRequest
	}

	storedRequest.Status = m.Confirmed
	rm.store[msg.Request.Origin] = storedRequest
	return storedRequest
}
