package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"group48.ttk4145.ntnu/elevators/comms"
	"group48.ttk4145.ntnu/elevators/elevatordriver"
	"group48.ttk4145.ntnu/elevators/elevatorio"
	"group48.ttk4145.ntnu/elevators/healthmonitor"
	"group48.ttk4145.ntnu/elevators/models"
	"group48.ttk4145.ntnu/elevators/orderserver"
	"group48.ttk4145.ntnu/elevators/requests"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to config file")
	flag.Parse()

	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	models.NumFloors = models.Floor(config.NumFloors)

	// Elevator IO module initialization
	var unvalidatedRequests = make(chan models.RequestMessage, 10)
	var floorSensorUpdates = make(chan int, 10)
	var obstructionSwitchUpdates = make(chan bool, 10)

	elevatorio.Init(config.ElevatorAddr, models.Id(config.LocalPeerId))
	go elevatorio.PollRequests(unvalidatedRequests)
	go elevatorio.PollFloorSensor(floorSensorUpdates)
	go elevatorio.PollObstructionSwitch(obstructionSwitchUpdates)

	// Elevator Driver module initialization
	var orders = make(chan models.Orders, 10)
	var internalEStateToComms = make(chan models.ElevatorState, 10)
	var eStatesUpdatesToOrders = make(chan models.ElevatorState, 10)
	go elevatordriver.Starter(
		obstructionSwitchUpdates,
		floorSensorUpdates,
		orders,
		unvalidatedRequests,
		internalEStateToComms,
		eStatesUpdatesToOrders,
		models.Id(config.LocalPeerId))

	// Order module initialization
	var aliveStatusOrders = make(chan []models.Id, 10)
	var validatedRequestsToOrder = make(chan models.Request, 10)
	go orderserver.RunOrderServer(
		validatedRequestsToOrder,
		eStatesUpdatesToOrders,
		aliveStatusOrders,
		orders,
		models.Id(config.LocalPeerId))

	// Health monitor module initialization
	var ping = make(chan models.Id, 10)
	var alivenessToRequests = make(chan []models.Id, 10)
	go healthmonitor.RunMonitor(
		models.Id(config.LocalPeerId),
		ping,
		alivenessToRequests,
		aliveStatusOrders)

	// Comms module initialization
	var internalValidatedRequestsToComms = make(chan models.Request, 10)

	go comms.RunComms(
		models.Id(config.LocalPeerId),
		config.LocalPort,
		internalEStateToComms,
		internalValidatedRequestsToComms,
		eStatesUpdatesToOrders,
		unvalidatedRequests,
		ping)

	// Request module initialization
	var validatedRequests = make([]chan<- models.Request, 2)
	validatedRequests[0] = validatedRequestsToOrder
	validatedRequests[1] = internalValidatedRequestsToComms
	go requests.RunRequestServer(
		models.Id(config.LocalPeerId),
		unvalidatedRequests,
		alivenessToRequests,
		validatedRequests)

	select {}
}

type Config struct {
	ElevatorAddr string `json:"elevator_addr"`
	NumFloors    int    `json:"num_floors"`
	LocalPeerId  int    `json:"local_peer_id"`
	LocalAddr    string `json:"local_addr"`
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
