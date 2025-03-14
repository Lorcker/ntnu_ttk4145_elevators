package orderserver

import (
	"encoding/json"
	"log"
	"os/exec"
	"strconv"

	m "group48.ttk4145.ntnu/elevators/models"
)

const pathToAssigner = "./orderserver/hall_request_assigner"

type HallRequests = [int(m.NumFloors)][2]bool
type CabRequests = [int(m.NumFloors)]bool

type elevator = struct {
	Behavior    string      `json:"behavior"`
	Floor       int         `json:"floor"`
	Direction   string      `json:"direction"`
	CabRequests CabRequests `json:"cabRequests"`
}

type jsonState = struct {
	HallRequests HallRequests        `json:"hallRequests"`
	States       map[string]elevator `json:"states"`
}

func calculateOrders(hr HallRequests, cr map[m.Id]CabRequests, elevators map[m.Id]m.ElevatorState) map[m.Id]m.Orders {
	jsonState := convertToJson(hr, cr, elevators)

	// Send jsonState to the assigner
	cmd := exec.Command(pathToAssigner, "-i", jsonState, "--includeCab")

	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("[orderserver] Error running hall_request_assigner: %v", err)
	}

	return convertFromJson(string(out))
}

func convertFromJson(j string) map[m.Id]m.Orders {
	var o map[string]m.Orders
	err := json.Unmarshal([]byte(j), &o)
	if err != nil {
		log.Fatalf("[orderserver] Error unmarshalling json: %v", err)
	}

	orders := make(map[m.Id]m.Orders)
	for k, v := range o {
		id, err := strconv.Atoi(k)
		if err != nil {
			log.Fatalf("[orderserver] Error converting id to int: %v", err)
		}
		orders[m.Id(id)] = v
	}
	return orders
}

func convertToJson(hr HallRequests, cr map[m.Id]CabRequests, elevators map[m.Id]m.ElevatorState) string {
	jsonState := jsonState{
		HallRequests: hr,
		States:       make(map[string]elevator),
	}

	for _, e := range elevators {
		id := strconv.Itoa(int(e.Id))
		jsonState.States[id] = elevator{
			Behavior:    behaviorToString(e.Behavior),
			Floor:       e.Floor,
			Direction:   directionToString(e.Direction),
			CabRequests: cr[e.Id],
		}
	}

	json, err := json.Marshal(jsonState)
	if err != nil {
		log.Fatalf("[orderserver] Error marshalling json: %v", err)
	}

	return string(json)
}

func behaviorToString(behavior m.ElevatorBehavior) string {
	switch behavior {
	case m.Idle:
		return "idle"
	case m.DoorOpen:
		return "doorOpen"
	case m.Moving:
		return "moving"
	default:
		return "Unknown"
	}
}

func directionToString(direction m.MotorDirection) string {
	switch direction {
	case m.Up:
		return "up"
	case m.Down:
		return "down"
	case m.Stop:
		return "stop"
	default:
		return "Unknown"
	}
}
