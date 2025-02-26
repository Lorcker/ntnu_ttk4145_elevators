package elevatordriver

import (
	// for printing
	"log"
	"testing"

	// Assuming PollRequests is in this package
	"group48.ttk4145.ntnu/elevators/elevatorio"
	"group48.ttk4145.ntnu/elevators/models" // Assuming your models are here
)

func TestStarter(t *testing.T) {
	pollObstructionSwitch := make(chan bool)
	pollFloorSensor := make(chan int)
	pollOrders := make(chan models.Orders)
	resolvedRequests := make(chan models.Request, 0)
	receiver := make([]chan<- models.ElevatorState, 0)
	id := models.Id(3)

	//For the test:
	receiverRequest := make(chan models.RequestMessage)

	go testPollOrders(pollOrders, receiverRequest)
	go testPollResolvedRequest(resolvedRequests)
	go elevatorio.PollRequests(receiverRequest)
	go elevatorio.PollFloorSensor(pollFloorSensor)
	go elevatorio.PollObstructionSwitch(pollObstructionSwitch)

	Starter(pollObstructionSwitch, pollFloorSensor, pollOrders, resolvedRequests, receiver, id)

}

func testPollOrders(reciever chan<- models.Orders, receiverRequest chan models.RequestMessage) {
	orders := initOrders(NFloors)
	for {
		select {
		case order_request := <-receiverRequest:
			orders[order_request.Request.Origin.Floor][order_request.Request.Origin.ButtonType] = true
			reciever <- orders
			log.Printf("Orders from testPollOrders: %v", orders)
		}

	}
}

func testPollResolvedRequest(reciever <-chan models.Request) {
	for {
		log.Printf("ResolvedRequest: %v", <-reciever)
	}
}
