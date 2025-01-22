package main

import (
	"Driver-go/elevio"
	"Driver-go/fsm"
	"fmt"
)

func main() {
	const NUMFLOORS = 4

	elevio.Init("localhost:15657", NUMFLOORS)

	var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	// Init elevator object
	elevator := fsm.Elevator{

		Floor:        0,
		Direction:    elevio.MD_Stop,
		Behavior:     fsm.EB_Idle,
		HallRequests: make([][]bool, NUMFLOORS),
		CabRequests:  make([]bool, NUMFLOORS),
	}
	for i := range elevator.HallRequests {
		elevator.HallRequests[i] = make([]bool, 2)
	}

	//Move to floor if inbetween
	if elevio.GetFloor() == -1 {
		for elevio.GetFloor() == -1 {
			elevator.Behavior = fsm.EB_Moving
			elevio.SetMotorDirection(elevio.MD_Down)
		}
		elevio.SetMotorDirection(elevio.MD_Stop)
		elevator.Behavior = fsm.EB_Idle
	}

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)
			fsm.OnRequestButtonPress(&elevator, a.Floor, a.Button)
			fmt.Println(elevator.CabRequests)
			fmt.Println(elevator.HallRequests)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			if a == NUMFLOORS-1 {
				d = elevio.MD_Down
			} else if a == 0 {
				d = elevio.MD_Up
			}
			//elevio.SetMotorDirection(d)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < NUMFLOORS; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}
}
