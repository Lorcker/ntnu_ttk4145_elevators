package orderserver

import (
	"fmt"
	"testing"

	"group48.ttk4145.ntnu/elevators/statedataserver"
)

// TestPollOrders tests the PollOrders function
func TestPollOrders(t *testing.T) {
	fmt.Println("TestPollOrders")
}

// TestCalculateOrders tests the CalculateOrders function
func TestCalculateOrders(t *testing.T) {
	// Create a channel for the global state
	channels := make(chan statedataserver.GlobalState)

	go CalculateOrders(channels)

	channels <- statedataserver.GlobalState{
		HallRequests: [][]bool{
			{false, false},
			{false, false},
			{false, false},
			{false, false},
		},
		GlobalElevators: map[int]statedataserver.GlobalElevator{
			0: {
				Floor:       0,
				Behavior:    statedataserver.EB_Idle,
				CabRequests: []bool{false, false, false, false},
				Direction:   0,
				IsAlive:     true,
			},
			1: {
				Floor:       1,
				Behavior:    statedataserver.EB_Idle,
				CabRequests: []bool{false, false, false, false},
				Direction:   0,
				IsAlive:     true,
			},
			2: {
				Floor:       3,
				Behavior:    statedataserver.EB_Idle,
				CabRequests: []bool{true, false, false, false},
				Direction:   0,
				IsAlive:     true,
			},
		},
	}
	print("TestCalculateOrders done")

}
