package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"group48.ttk4145.ntnu/elevators/internal/comms"
	"group48.ttk4145.ntnu/elevators/internal/driver"
	"group48.ttk4145.ntnu/elevators/internal/elevatorio"
	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/message"
	enginemonitor "group48.ttk4145.ntnu/elevators/internal/monitors/engine"
	obstructionmonitor "group48.ttk4145.ntnu/elevators/internal/monitors/obstruction"
	healthmonitor "group48.ttk4145.ntnu/elevators/internal/monitors/peers"
	"group48.ttk4145.ntnu/elevators/internal/orders"
	"group48.ttk4145.ntnu/elevators/internal/requests"
)

// channelBufferSize can be used to control the buffer size of the channels
// Without buffer the channels will block until the message is received
// This can lead to deadlocks when modules are waiting for each other
const channelBufferSize = 10

func main() {
	configPath := flag.String("config", "configs/config.json", "Path to config file")
	flag.Parse()

	config := LoadConfig(*configPath)
	localId := elevator.Id(config.LocalPeerId)

	// The channels are structured as follows:
	// 	- Update channels are responsible for sending input from one ore more modules to another module.
	// 	- Notify channels are triggered when a module receives a msg on the update channel and the state of data has changed.

	// These channels are responsible for sending updates from the elevator IO to the [driver] and [enginemonitor] module.
	// Updates are triggered by values read from the elevator IO.
	floorSensorToDriver := make(chan message.FloorArrival, channelBufferSize)
	floorSensorToMotorMonitor := make(chan message.FloorArrival, channelBufferSize)
	obstructionSwitchUpdateToDriver := make(chan message.Obstruction, channelBufferSize)
	obstructionSwitchUpdateToMonitor := make(chan message.Obstruction, channelBufferSize)

	// These channels are responsible to transport all updates concerning requests.
	// All modules that want to update the state of a request should send a message to requestStateUpdateToRequest.
	// These are:
	// 	- [elevatorio] When a button is pressed on the elevator it sends a unconfirmed request to the [request] module
	// 	- [driver] When a request is resolved by the local elevator is sends a request with absent status to the [request] module
	// 	- [comms] When the local peer receives a request from another peer it sends a request to the [request] module
	// The [request] module then updates the state of the request and sends a notification to the [orders] and [comms] module.
	requestStateUpdateToRequest := make(chan message.RequestState, channelBufferSize)
	requestStateNotifyToOrders := make(chan message.RequestState, channelBufferSize)
	requestStateNotifyToComms := make(chan message.RequestState, channelBufferSize)

	// This channel is responsible for sending newly calculated orders from the [orders] module to the [driver] module.
	// Messages are only sent when the orders have changed.
	orderUpdates := make(chan message.ServiceOrder, channelBufferSize)

	// These channels are responsible for sending updates concerning the state of the elevator.
	// The [driver] module sends updates to the [orders] and [comms] module.
	// The updates are sent periodically using a ticker defined in the [driver] module.
	elevatorStateUpdateToOrders := make(chan message.ElevatorState, channelBufferSize)
	elevatorStateUpdateToComms := make(chan message.ElevatorState, channelBufferSize)
	elevatorStateUpdateToEngineMonitor := make(chan message.ElevatorState, channelBufferSize)

	// These channels are responsible for sending updates concerning the aliveness of the peers.
	// The [comms] module send heartbeats to the [healthmonitor] module if it receives messages from another peer.
	// If the health of peer changes (i.e a peer has died or a new peer has joined),
	// the [healthmonitor] module sends a notification to the [requests], [comms], and [orders] module.
	alivePeersUpdate := make(chan message.PeerSignal, channelBufferSize)
	alivePeersNotifyToOrders := make(chan message.ActivePeers, channelBufferSize)
	alivePeersNotifyToRequests := make(chan message.ActivePeers, channelBufferSize)
	alivePeersNotifyToComms := make(chan message.ActivePeers, channelBufferSize)

	// The [elevatorio] module is responsible for communicating with the elevator hardware.
	// It produces outputs:
	//  - Updates to the [request] module (unconfirmed requests) when a button is pressed
	//  - Updates to the [driver] module (floor sensor and obstruction switch) when the hardware is triggered
	//  - Updates to the [enginemonitor] module (floor sensor) hwen the hardware is triggered
	elevatorio.Init(config.ElevatorAddr, localId)
	go elevatorio.PollNewRequests(requestStateUpdateToRequest)
	go elevatorio.PollFloorSensor(floorSensorToDriver)
	go elevatorio.PollFloorSensor(floorSensorToMotorMonitor)
	go elevatorio.PollObstructionSwitch(obstructionSwitchUpdateToDriver)
	go elevatorio.PollObstructionSwitch(obstructionSwitchUpdateToMonitor)

	// The [driver] module is responsible for controlling the elevator hardware.
	// It takes as input:
	// 	- Updates from the elevator hardware (floor sensor, obstruction switch)
	// 	- Updates from the [orders] module (new orders)
	// It produces outputs:
	//  - Updates to the [request] module (resolved requests) when a request is resolved
	//  - Updates to the [comms] and [order] module (elevator state) based on a polling rate
	//  - Sends a heartbeat update to the [healthmonitor] module to indicate that the local peer is dead, due to failure
	go driver.RunDriver(
		obstructionSwitchUpdateToDriver,
		floorSensorToDriver,
		orderUpdates,
		requestStateUpdateToRequest,
		elevatorStateUpdateToComms,
		elevatorStateUpdateToOrders,
		elevatorStateUpdateToEngineMonitor,
		localId,
	)

	// The [requests] module is responsible for managing the state of the requests.
	// This includes the acknowledgment of other peers to ensure redundancy.
	// It takes as input:
	// 	- Updates from the [elevatorio] module (unconfirmed requests) which are triggered by button presses
	// 	- Updates from the [driver] module (absent requests) which are triggered by the local elevator when a request is resolved
	// 	- Updates from the [comms] module (requests from other peers)
	//  - Updates from the [healthmonitor] module (peer aliveness) to determine acknowledgment status
	// It produces outputs:
	//  - Notifications to the [orders] and [comms] module when the state of a request has changed
	go requests.RunRequestServer(
		localId,
		requestStateUpdateToRequest,
		alivePeersNotifyToRequests,
		requestStateNotifyToComms,
		requestStateNotifyToOrders,
	)

	// The [orders] module is responsible for managing the orders and calculating the orders for the local elevator.
	// An order includes all requests that should be handled by the local elevator.
	// It takes as input:
	// 	- Updates from the [requests] module (request state updates)
	// 	- Updates from the [driver] and [comms] module (local and external elevator state updates)
	// 	- Updates from the [healthmonitor] module (peer aliveness) to exclude dead peers from the order calculations
	// It produces outputs:
	//  - Updates to the [driver] module (new orders) when the orders have changed
	go orders.RunOrderServer(
		localId,
		requestStateNotifyToOrders,
		elevatorStateUpdateToOrders,
		alivePeersNotifyToOrders,
		orderUpdates,
	)

	// The [healthmonitor] module is responsible for monitoring the health of the peers.
	// It takes as input:
	// 	- Updates from the [comms] module (peer heartbeats) to store the last time a peer was seen
	// It produces outputs:
	//  - Notifications to the [requests] and [orders] module when the aliveness of a peer has changed (death or new peer)
	go healthmonitor.RunMonitor(
		localId,
		alivePeersUpdate,
		alivePeersNotifyToRequests,
		alivePeersNotifyToOrders,
		alivePeersNotifyToComms,
	)

	// The [enginemonitor] module is responsible for monitoring the health of the engine
	// It takes as input:
	//  - Updated from the [elevio] module (floor) to check that the elevator moved
	//  - Updated from the [driver] module (state) to register that the elevator should be moving
	// It produced ouputs:
	// 	- Notification to the [healthmonitor] module when state of the engine changed (dead <-> alive)
	go enginemonitor.RunEngineMonitor(
		localId,
		floorSensorToMotorMonitor,
		elevatorStateUpdateToEngineMonitor,
		alivePeersUpdate,
	)

	// The [obstruct] module is responsible for monitoring the the status of the obstruction switch
	// If obstructed for a long period of time we consider ourselv not functional
	// It takes as input:
	//  - Updated from the [elevio] module (obstruction) to register obstruction
	// It produced ouputs:
	// 	- Notification to the [healthmonitor] module when state of perma interruption changes
	go obstructionmonitor.RunObstructionMonitor(
		localId,
		obstructionSwitchUpdateToMonitor,
		alivePeersUpdate,
	)

	// The [comms] module is responsible for handling the communication between the peers.
	// This includes sending the local elevator state and all information about about the requests of the local and external peers.
	// These messages are sent via UDP broadcast based on a regular interval defined in the [comms] module.
	// It takes as input:
	// 	- Updates from the [driver] module (local elevator state) which are cached and propagated to the other peers
	// 	- Updates from the [requests] module (request state updates) which are cached and propagated to the other peers
	// It produces outputs:
	//  - Notifications to the [orders] and [requests] module when an external peer has a different state of a request
	//  - Notifications to the [orders] module about the elevator state of the external peers
	//  - Notifications to the [healthmonitor] module to update the aliveness of the peers
	go comms.RunComms(
		localId,
		config.LocalPort,
		elevatorStateUpdateToComms,
		requestStateNotifyToComms,
		alivePeersNotifyToComms,
		elevatorStateUpdateToOrders,
		requestStateUpdateToRequest,
		alivePeersUpdate,
	)

	// Block forever
	select {}
}

type Config struct {
	// ElevatorAddr is the address of the elevator simulator or the elevator hardware.
	ElevatorAddr string `json:"elevator_addr"`

	// LocalPeerId is the id of the local elevator.
	LocalPeerId int `json:"local_peer_id"`

	// LocalPort is the port the local [comms] module listens to and sends broadcasts on.
	LocalPort int `json:"local_port"`
}

// LoadConfig loads the configuration from a file
func LoadConfig(filename string) *Config {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("[main] Failed to open config file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := &Config{}
	err = decoder.Decode(config)
	if err != nil {
		log.Fatalf("[main] Failed to decode config file: %v", err)
	}

	return config
}
