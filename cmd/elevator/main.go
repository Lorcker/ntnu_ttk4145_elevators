package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"group48.ttk4145.ntnu/elevators/internal/comms"
	"group48.ttk4145.ntnu/elevators/internal/driver"
	"group48.ttk4145.ntnu/elevators/internal/elevatorio"
	"group48.ttk4145.ntnu/elevators/internal/healthmonitor"
	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/message"
	"group48.ttk4145.ntnu/elevators/internal/orders"
	"group48.ttk4145.ntnu/elevators/internal/requests"
)

const channelBufferSize = 10

func main() {
	configPath := flag.String("config", "./configs/config.json", "Path to config file")
	flag.Parse()

	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	localId := elevator.Id(config.LocalPeerId)

	floorSensorUpdate := make(chan message.FloorSensor, channelBufferSize)
	obstructionSwitchUpdate := make(chan message.ObstructionSwitch, channelBufferSize)

	requestStateUpdateToRequest := make(chan message.RequestStateUpdate, channelBufferSize)
	requestStateNotifyToOrders := make(chan message.RequestStateUpdate, channelBufferSize)
	requestStateNotifyToComms := make(chan message.RequestStateUpdate, channelBufferSize)

	orderUpdates := make(chan message.Order, channelBufferSize)

	elevatorStateUpdateToOrders := make(chan message.ElevatorStateUpdate, channelBufferSize)
	elevatorStateUpdateToComms := make(chan message.ElevatorStateUpdate, channelBufferSize)

	heartbeatUpdate := make(chan message.PeerHeartbeat, channelBufferSize)
	alivePeersNotifyToOrders := make(chan message.AlivePeersUpdate, channelBufferSize)
	alivePeersNotifyToRequests := make(chan message.AlivePeersUpdate, channelBufferSize)

	elevatorio.Init(config.ElevatorAddr, localId)
	go elevatorio.PollNewRequests(requestStateUpdateToRequest)
	go elevatorio.PollFloorSensor(floorSensorUpdate)
	go elevatorio.PollObstructionSwitch(obstructionSwitchUpdate)

	go driver.RunDriver(
		obstructionSwitchUpdate,
		floorSensorUpdate,
		orderUpdates,
		requestStateUpdateToRequest,
		elevatorStateUpdateToComms,
		elevatorStateUpdateToOrders,
		localId,
	)

	requestSubscribers := []chan<- message.RequestStateUpdate{
		requestStateNotifyToOrders,
		requestStateNotifyToComms,
	}
	go requests.RunRequestServer(
		localId,
		requestStateUpdateToRequest,
		alivePeersNotifyToRequests,
		requestSubscribers,
	)

	go orders.RunOrderServer(
		localId,
		requestStateNotifyToOrders,
		elevatorStateUpdateToOrders,
		alivePeersNotifyToOrders,
		orderUpdates,
	)

	go healthmonitor.RunMonitor(
		localId,
		heartbeatUpdate,
		alivePeersNotifyToRequests,
		alivePeersNotifyToOrders,
	)

	go comms.RunComms(
		localId,
		config.LocalPort,
		elevatorStateUpdateToComms,
		requestStateNotifyToComms,
		elevatorStateUpdateToOrders,
		requestStateUpdateToRequest,
		heartbeatUpdate,
	)

	select {}
}

type Config struct {
	ElevatorAddr string `json:"elevator_addr"`
	LocalPeerId  int    `json:"local_peer_id"`
	LocalPort    int    `json:"local_port"`
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := &Config{}
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
