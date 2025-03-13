// requests is a module responsible for processing and storing requests, and updating the button lighting.
package requests

import (
	"log"

	"group48.ttk4145.ntnu/elevators/elevatorio"
	"group48.ttk4145.ntnu/elevators/models"
)

// RunRequestServer should be run as a goroutine and takes care of processing requests.
//
// The processing of requests is done by a requestManager, which keeps track of the state of the requests.
// The button lighting is set for the local elevator if the request is for the local elevator.
func RunRequestServer(
	local models.Id,
	incomingRequests <-chan models.RequestMessage,
	currentAlivePeers <-chan []models.Id,
	subscribers []chan<- models.Request) {

	var requestManager = newRequestManager(local)

	for {
		select {
		case msg := <-incomingRequests:
			req := requestManager.process(msg)
			log.Printf("[requests] Processed a new request:\n\tIncoming: %v\n\tProcessed: %v", msg.Request, req)

			setButtonLighting(local, req)

			for _, s := range subscribers {
				s <- req
			}
		case ap := <-currentAlivePeers:
			log.Printf("[requests] Received alive peers: %v", ap)
			requestManager.alivePeers = ap
		}
	}
}

// setButtonLighting sets the button lighting for the request.
//
// If the request is not for the local elevator, the lighting is not set.
func setButtonLighting(local models.Id, req models.Request) {
	if elev, ok := req.Origin.Source.(models.Elevator); ok && elev.Id != local {
		return // Lighting does not concern this elevator
	}
	targetState := req.Status == models.Confirmed

	elevatorio.SetButtonLamp(req.Origin.ButtonType, req.Origin.Floor, targetState)
	log.Printf("[requests] Set button lamp: %v, %v, %v", req.Origin.ButtonType, req.Origin.Floor, targetState)
}
