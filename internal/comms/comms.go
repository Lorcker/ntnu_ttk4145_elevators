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
	Alive    bool
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
	fromDriver <-chan message.ElevatorState,
	fromRequests <-chan message.RequestState,
	toOrders chan<- message.ElevatorState,
	toRequest chan<- message.RequestState,
	toHealthMonitor chan<- message.PeerSignal) {

	var sendTicker = time.NewTicker(SendInterval)
	var internalEsBuffer = make([]elevator.State, 0)
	var registry = newRequestRegistry()
	localAlive := true

	sendUdp := make(chan udpMessage)
	receiveUdp := make(chan udpMessage)
	go bcast.Transmitter(port, sendUdp)
	go bcast.Receiver(port, receiveUdp)

	for {
		select {
		case msg := <-fromDriver:
			handleDriverMessage(msg, &internalEsBuffer, &localAlive)

		case msg := <-fromRequests:
			handleRequestMessage(msg, &registry)

		case <-sendTicker.C:
			if len(internalEsBuffer) == 0 {
				// No internal elevator state to send yet
				continue
			}

			u := udpMessage{
				Source:   local,
				Registry: registry,
				Alive:    localAlive,
				EState:   internalEsBuffer[0],
			}
			sendUdp <- u
		case msg := <-receiveUdp:
			if msg.Source == local {
				// Ignore messages from self
				continue
			}
			toHealthMonitor <- message.PeerSignal{Id: msg.Source, Alive: msg.Alive}
			toOrders <- message.ElevatorState{Elevator: msg.Source, State: msg.EState}

			changedRequests := registry.diff(msg.Source, msg.Registry)
			logRegistryDiff(msg.Source, changedRequests, registry, msg.Registry)
			for _, msg := range changedRequests {
				toRequest <- msg
			}
		}
	}

}

func handleDriverMessage(msg message.ElevatorState, internalBuffer *[]elevator.State, localAlive *bool) {
	if len(*internalBuffer) != 0 && (*internalBuffer)[0] == msg.State {
		return
	}

	if len(*internalBuffer) == 0 {
		*internalBuffer = append(*internalBuffer, msg.State)
		log.Printf("[comms] Received initial local elevator state update from [driver]: %v", msg)
	}

	log.Printf("[comms] Received new local elevator state update from [driver]: %v", msg)
	(*internalBuffer)[0] = msg.State
	*localAlive = msg.Alive
}

func handleRequestMessage(msg message.RequestState, registry *requestRegistry) {
	before := fmt.Sprintf("%v", registry)
	registry.update(msg.Request)
	after := fmt.Sprintf("%v", registry)

	if before != after {
		log.Printf("[comms] Updated registry after getting msg from [requests]:\n\tRequest: %v\n\tBef: %v\n\tAft: %v", msg, before, after)
	}
}

func logRegistryDiff(peer elevator.Id, changed []message.RequestState, internal, external requestRegistry) {
	if len(changed) == 0 {
		return
	}
	c := ""
	for _, req := range changed {
		c += fmt.Sprintf("\t%v\n", req.Request)
	}
	registry := fmt.Sprintf("\tRegistries:\n\tInternal: %v\n\tExternal: %v", internal, external)
	changedRequests := fmt.Sprintf("\tChanged requests:\n%v", c)
	log.Printf("[comms] Received registry diff from %v that caused updates:\n%v\n%v", peer, changedRequests, registry)
}
