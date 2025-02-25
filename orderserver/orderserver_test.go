package orderserver

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"group48.ttk4145.ntnu/elevators/models"
)

// TestCalculateOrders tests the CalculateOrders function
func TestCalculateOrders(t *testing.T) {
	// Create a channel for the global state
	fmt.Println("TestCalculateOrders started")
	validatedRequests := make(chan models.Request, 1)
	alive := make(chan []models.Id, 1)
	orders := make(chan models.Orders, 1)
	state := make(chan models.ElevatorState, 1)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		RunOrderServer(validatedRequests, state, alive, orders)
	}()
	// Send test data to the channels
	alive <- []models.Id{1, 2}
	state <- models.ElevatorState{
		Id:        1,
		Floor:     0,
		Direction: models.Stop,
		Behavior:  models.Idle,
	}
	validatedRequests <- models.Request{
		Origin: models.Origin{
			Source:     models.Hall{},
			Floor:      1,
			ButtonType: models.HallUp,
		},
		Status: models.Confirmed,
	}

	// Wait for the goroutine to process the input
	wg.Wait()

	// Check the output from the orders channel
	select {
	case o := <-orders:
		if len(o) == 0 {
			t.Errorf("Expected non-empty orders, got empty orders")
		} else {
			fmt.Println("Orders received:", o)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("Timeout waiting for orders")
	}

	fmt.Println("TestCalculateOrders done")

}
