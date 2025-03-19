package comms

import (
	"Network-go/network/bcast"
	"fmt"
	"log"
	"time"

	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/message"
)

const SendInterval = time.Millisecond * 100

type udpMessage struct {
	Source   elevator.Id
	Registry requestRegistry
	EState   elevator.State
}

// # RunComms runs the communication module
//
// It listens for updates on the local elevator state and validated requests channels.
// It send UDP messages with the local elevator state and all system requests to the broadcast address in a regular interval.
// It listens for incoming UDP messages and sends the elevator state and changed requests to the outgoing channels.
// It sends a health monitor ping on the health monitor ping channel when it receives an update from the local elevator state or validated requests channels.
func RunComms(
	local elevator.Id,
	port int,
	fromDriver <-chan message.ElevatorStateUpdate,
	fromRequests <-chan message.RequestStateUpdate,
	toOrders chan<- message.ElevatorStateUpdate,
	toRequest chan<- message.RequestStateUpdate,
	toHealthMonitor chan<- message.PeerHeartbeat) {

	var sendTicker = time.NewTicker(SendInterval)
	var internalEs = elevator.State{}
	var registry = newRequestRegistry()

	sendUdp := make(chan udpMessage)
	receiveUdp := make(chan udpMessage)
	go bcast.Transmitter(port, sendUdp)
	go bcast.Receiver(port, receiveUdp)

	for {
		select {
		case msg := <-fromDriver:
			if internalEs != msg.State {
				log.Printf("[comms] Received new local elevator state update from [driver]: %v", msg)
				internalEs = msg.State
			}

		case msg := <-fromRequests:
			before := fmt.Sprintf("%v", registry)

			registry.update(msg.Request)

			after := fmt.Sprintf("%v", registry)
			if before != after {
				log.Printf("[comms] Updated registry after getting msg from [requests]:\n\tRequest: %v\n\tBefore: %v\n\tAfter: %v", msg, before, after)
			}

		case <-sendTicker.C:
			if internalEs == (elevator.State{}) {
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

			toHealthMonitor <- message.PeerHeartbeat{Id: msg.Source}
			toOrders <- message.ElevatorStateUpdate{Elevator: msg.Source, State: msg.EState}

			changedRequests := registry.diff(msg.Source, msg.Registry)
			if len(changedRequests) > 0 {
				log.Printf("[comms] Received an external registry that changed state:\n\tFromPeer: %d\n\tChangedReqs: %v\n\tInternalRegistry:%v\n\tExternalRegistry:%v", msg.Source, changedRequests, registry, msg.Registry)
			}
			for _, msg := range changedRequests {
				toRequest <- msg
			}
		}
	}

}
