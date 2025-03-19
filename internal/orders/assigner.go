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

// jsonState struct used for json marshalling and unmarshaling
//
// It is used to represent the state of the system as expected by the assigner executable
type jsonState = struct {
	HallRequests hallRequests        `json:"hallRequests"`
	States       map[string]jsonElev `json:"states"`
}

// jsonElev struct used for json marshalling and unmarshaling
//
// It is used to represent the state of an jsonElev as expected by the assigner executable
type jsonElev = struct {
	Behavior    string      `json:"behavior"`
	Floor       int         `json:"floor"`
	Direction   string      `json:"direction"`
	CabRequests cabRequests `json:"cabRequests"`
}

// dirToString converts a motor direction to the corresponding string
var dirToString = map[elevator.MotorDirection]string{
	elevator.Up:   "up",
	elevator.Down: "down",
	elevator.Stop: "stop",
}

// behaviorToString converts an elevator behavior to the corresponding string
var behaviorToString = map[elevator.Behavior]string{
	elevator.Idle:     "idle",
	elevator.Moving:   "moving",
	elevator.DoorOpen: "doorOpen",
}

// calculateOrders calculates the orders for the elevators
//
// It sends the state of the system to the hall_request_assigner executable and returns the orders.
// This approach is used to avoid having to implement the assigner logic in Go which would lead to code duplication and potential bugs.
func calculateOrders(hr hallRequests, cr map[elevator.Id]cabRequests, elevators map[elevator.Id]elevator.State) map[elevator.Id]elevator.Order {
	jsonState := convertToJson(hr, cr, elevators)

	// Send jsonState to the assigner
	cmd := exec.Command(pathToAssigner, "-i", jsonState, "--includeCab")

	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("[orderserver] Error running hall_request_assigner: %v", err)
	}

	return convertFromJson(string(out))
}

// convertFromJson converts a json string to a map of orders
func convertFromJson(j string) map[elevator.Id]elevator.Order {
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

// convertToJson converts the state of the system to a json string
func convertToJson(hr hallRequests, cr map[elevator.Id]cabRequests, elevators map[elevator.Id]elevator.State) string {
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
