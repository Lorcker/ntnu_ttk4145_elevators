// requests is a module responsible for processing and storing requests, and updating the button lighting.
package requests

import (
	"log"

	"group48.ttk4145.ntnu/elevators/internal/elevatorio"
	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/message"
	"group48.ttk4145.ntnu/elevators/internal/models/request"
)

// RunRequestServer should be run as a goroutine and takes care of processing requests.
//
// The processing of requests is done by a requestManager, which keeps track of the state of the requests.
// The button lighting is set for the local elevator if the request is for the local elevator.
func RunRequestServer(
	local elevator.Id,
	requestStateUpdates <-chan message.RequestStateUpdate,
	currentAlivePeers <-chan message.AlivePeersUpdate,
	notifyComms chan<- message.RequestStateUpdate,
	notifyOrders chan<- message.RequestStateUpdate) {

	var requestManager = newRequestManager(local)

	for {
		select {
		case msg := <-requestStateUpdates:
			req := requestManager.Process(msg)
			setButtonLighting(local, req)

			uMsg := message.RequestStateUpdate{
				Source:  local,
				Request: req,
			}
			notifyComms <- uMsg
			notifyOrders <- uMsg

		case ap := <-currentAlivePeers:
			requestManager.UpdateAlivePeers(ap.Peers)
		}
	}
}

// setButtonLighting sets the button lighting for the request.
func setButtonLighting(local elevator.Id, req request.Request) {
	if cab, ok := req.Origin.(request.Cab); ok && cab.Id != local {
		// The request is for another elevator, do not set the button lighting
		return
	}

	targetState := req.Status == request.Confirmed
	elevatorio.SetButtonLamp(req.Origin.GetButtonType(), req.Origin.GetFloor(), targetState)
	log.Printf("[requests] Set button lamp: %v, %v, %v", req.Origin.GetButtonType(), req.Origin.GetFloor(), targetState)
}
