package elevatordriver

import (
	"fmt" // for printing
	"testing"
	"time"

	"group48.ttk4145.ntnu/elevators/elevatorio" // Assuming PollRequests is in this package
	"group48.ttk4145.ntnu/elevators/models"     // Assuming your models are here
)

func initOrders(numFloors int) models.Orders {
	var orders models.Orders = make([][3]bool, numFloors)
	for i := 0; i < numFloors; i++ {
		for j := 0; j < 3; j++ {
			orders[i][j] = false
		}
	}
	return orders
}

func printOrders(orders models.Orders) {
	// Iterate through the outer slice (rows)
	fmt.Printf("Floor\t Up\t Down\t Cab\n")
	for i := 0; i < len(orders); i++ {
		// Iterate through the inner slice (columns) at each row
		fmt.Printf("%d\t", i)
		for j := 0; j < len(orders[i]); j++ {
			// Print the Order information
			fmt.Printf("%t\t ", orders[i][j])
		}
		fmt.Printf("\n\n")
	}
}

func printElevatorState(elevator models.ElevatorState) {
	fmt.Printf("\n\nFloor: %d\n", elevator.Floor)
	fmt.Printf("Behavior: %d\n", elevator.Behavior)
	fmt.Printf("Direction: %d\n\n", elevator.Direction)
}

func onInitBetweenFloors() {
	if elevatorio.GetFloor() != 0 {
		elevatorio.SetMotorDirection(-1)
	}
	for elevatorio.GetFloor() != 0 {
	}
	elevatorio.SetMotorDirection(0)
}

func initTestElevator(nFloors int) models.Orders {
	elevatorio.Init("localhost:15680", 4)
	onInitBetweenFloors()
	orders := initOrders(nFloors)
	setAllElevatorLights(orders)

	return orders

}

func TestMain(t *testing.T) {
	var orders models.Orders = initTestElevator(4)

	elevator := models.ElevatorState{Id: 10, Floor: 0, Behavior: models.Idle, Direction: models.MotorDirection(0)}

	// Create a channel to receive the request messages
	receiverOrder := make(chan models.RequestMessage, 10)
	recieverFloorSensor := make(chan int, 10)
	recieverDoorTimer := make(chan bool, 10)
	recieverObstructionSwitch := make(chan bool, 10)
	recieverStopButton := make(chan bool, 10)

	// Run the PollRequests function in a separate goroutine
	go elevatorio.PollRequests(receiverOrder)
	go elevatorio.PollFloorSensor(recieverFloorSensor)
	go elevatorio.PollObstructionSwitch(recieverObstructionSwitch)
	go elevatorio.PollStopButton(recieverStopButton)

	// Use a timeout for the test to avoid hanging forever
	timeout := time.After(10 * time.Second)

	// Loop and print received request messages from the channel
	timer := time.NewTimer(3 * time.Second)
	timer.Stop()
	isObstructed := false

	for {
		select {
		case order_request := <-receiverOrder:
			orders[order_request.Request.Origin.Floor][order_request.Request.Origin.ButtonType] = true
			printOrders(orders)
			setAllElevatorLights(orders)
			HandleOrderEvent(&elevator, orders, recieverDoorTimer)

		case <-recieverDoorTimer:
			OpenDoor(&elevator)
			timer.Reset(3 * time.Second)

		case <-recieverObstructionSwitch:
			isObstructed = !isObstructed

		case <-timer.C:
			if elevator.Behavior == models.DoorOpen && !isObstructed {
				HandleDoorTimerEvent(&elevator, orders, recieverDoorTimer)
			} else {
				fmt.Printf("Remove Obstruction!\n")
				timer.Reset(3 * time.Second)
			}

		case floor_sensor := <-recieverFloorSensor:
			HandleFloorsensorEvent(&elevator, orders, floor_sensor, recieverDoorTimer)
			printElevatorState(elevator)

		case <-recieverStopButton:
			EmergencyStop(&elevator) // So far does nothing

		case <-timeout:
			// Timeout to stop the test if nothing happens
			// fmt.Println("No msg last 10 sec")
			timeout = time.After(10 * time.Second)
		}
	}
}
