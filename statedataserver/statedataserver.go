package statedataserver

import (
	"group48.ttk4145.ntnu/elevators/elevatorio"
)

type GlobalState struct {
	HallRequests    [][]bool
	GlobalElevators map[int]GlobalElevator
}

type GlobalElevator struct {
	Floor       int
	Behavior    ElevatorBehavior
	CabRequests []bool
	Direction   elevatorio.MotorDirection
	IsAlive     bool
}

type ElevatorBehavior int

const (
	EB_Idle ElevatorBehavior = iota
	EB_DoorOpen
	EB_Moving
)

func GlobalConsistentStateVerifier(receiver chan<- GlobalState) {}

func GlobalStatePoll(receiver chan<- GlobalState) {}

func GlobalStateUpdater(receiver <-chan GlobalState) {}
