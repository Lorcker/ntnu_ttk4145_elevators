package orders

import (
	"encoding/json"
	"log"
	"os/exec"
	"runtime"
	"strconv"

	"path/filepath"

	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
)

func getBasePath() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Dir(b)
}

// Path to the hall_request_assigner executable
var pathToAssigner = filepath.Join(getBasePath(), "../../external/assigner/hall_request_assigner")

// jsonState struct is used for json marshalling and unmarshaling
type jsonState = struct {
	HallRequests hallRequests        `json:"hallRequests"`
	States       map[string]jsonElev `json:"states"`
}
type jsonElev = struct {
	Behavior    string      `json:"behavior"`
	Floor       int         `json:"floor"`
	Direction   string      `json:"direction"`
	CabRequests cabRequests `json:"cabRequests"`
}

var dirToString = map[elevator.MotorDirection]string{
	elevator.Up:   "up",
	elevator.Down: "down",
	elevator.Stop: "stop",
}
var behaviorToString = map[elevator.Behavior]string{
	elevator.Idle:     "idle",
	elevator.Moving:   "moving",
	elevator.DoorOpen: "doorOpen",
}

// calculateOrders calculates the orders for the elevators
//
// It sends the state of the system to the hall_request_assigner executable and returns the orders.
// This approach is used to avoid having to implement the assigner logic in Go which would lead to code duplication and potential bugs.
func calculateOrders(hr hallRequests, cr map[elevator.Id]cabRequests, elevators map[elevator.Id]elevator.State, al map[elevator.Id]bool) map[elevator.Id]elevator.Order {
	//only include states of alive elevators
	newStates := make(map[elevator.Id]elevator.State)
	for id, state := range elevators {
		if al[id] {
			newStates[id] = state
		}
	}
	elevators = newStates

	jsonState := marshal(hr, cr, elevators)

	cmd := exec.Command(pathToAssigner, "-i", jsonState, "--includeCab")

	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("[orderserver] Error running hall_request_assigner: %v", err)
	}

	return unmarshal(string(out))
}

// unmarshal converts a json string to a map of orders
func unmarshal(j string) map[elevator.Id]elevator.Order {
	var o map[string]elevator.Order
	err := json.Unmarshal([]byte(j), &o)
	if err != nil {
		log.Fatalf("[orderserver] Error unmarshalling json: %v", err)
	}

	orders := make(map[elevator.Id]elevator.Order)
	for k, v := range o {
		id, err := strconv.Atoi(k)
		if err != nil {
			log.Fatalf("[orderserver] Error converting id to int: %v", err)
		}
		orders[elevator.Id(id)] = v
	}
	return orders
}

// marshal converts the state of the system to a json string
func marshal(hr hallRequests, cr map[elevator.Id]cabRequests, elevators map[elevator.Id]elevator.State) string {
	jsonState := jsonState{
		HallRequests: hr,
		States:       make(map[string]jsonElev),
	}

	for id, e := range elevators {
		idStr := strconv.Itoa(int(id))
		jsonState.States[idStr] = jsonElev{
			Behavior:    behaviorToString[e.Behavior],
			Floor:       int(e.Floor),
			Direction:   dirToString[e.Direction],
			CabRequests: cr[id],
		}
	}

	json, err := json.Marshal(jsonState)
	if err != nil {
		log.Fatalf("[orderserver] Error marshalling json: %v", err)
	}

	return string(json)
}
