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
	subscribers []chan<- message.RequestStateUpdate) {

	var requestManager = newRequestManager(local)

	for {
		select {
		case msg := <-requestStateUpdates:
			req := requestManager.process(msg)
			log.Printf("[requests] Processed a new request:\n\tIncoming: %v\n\tProcessed: %v", msg.Request, req)

			setButtonLighting(local, req)

			for _, s := range subscribers {
				s <- message.RequestStateUpdate{
					Source:  local,
					Request: req,
				}
			}

		case ap := <-currentAlivePeers:
			log.Printf("[requests] Received alive peers: %v", ap)
			requestManager.alivePeers = ap.Peers
		}
	}
}

// setButtonLighting sets the button lighting for the request.
//
// If the request is not for the local elevator, the lighting is not set.
func setButtonLighting(local elevator.Id, req request.Request) {
	targetState := req.Status == request.Confirmed

	if request.IsCab(req) && req.Origin.(request.Cab).Id == local {
		elevatorio.SetButtonLamp(elevator.Cab, req.Origin.GetFloor(), targetState)
		log.Printf("[requests] Set button lamp: %v, %v, %v", elevator.Cab, req.Origin.GetFloor(), targetState)
	} else if request.IsFromHall(req) {
		hall := req.Origin.(request.Hall)
		if hall.Direction == request.Up {
			elevatorio.SetButtonLamp(elevator.HallUp, hall.Floor, targetState)
			log.Printf("[requests] Set button lamp: %v, %v, %v", elevator.HallUp, hall.Floor, targetState)
		} else {
			elevatorio.SetButtonLamp(elevator.HallDown, hall.Floor, targetState)
			log.Printf("[requests] Set button lamp: %v, %v, %v", elevator.HallDown, hall.Floor, targetState)
		}
	}
}
