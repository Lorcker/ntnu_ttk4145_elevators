package orderserver

import (
	"fmt"
	"time"

	"sort"

	"group48.ttk4145.ntnu/elevators/models"
)

type State struct {
	models.ElevatorState
	CabRequests []bool
	time        time.Time
}

type Reqest struct {
	active     bool
	assignedTo models.Id
}

func optimalHallRequests(elevators elevators) [][3]bool {
	fmt.Println("optimalHallRequests started")
	reqs := addRequests(elevators)
	states := initialStates(elevators)

	for i := range states {
		performInitialMove(&states[i], &reqs)
	}
	for {
		// Sort states by time
		sort.Slice(states, func(i, j int) bool {
			return states[i].time.Before(states[j].time)
		})
		done := true
		if anyUnassigned(reqs) {
			assignImmediate(&reqs, &states)
			done = false
		}
		if unvisitedAreImmediatelyAssignable(reqs, states) {
			assignImmediate(&reqs, &states)
			done = true
		}
		if done {
			break
		}
		performSingleMove(&states[0], &reqs)
	}
	result := make([][3]bool, numFloors)
	for f, floorRequests := range reqs {
		for c := range 3 {
			result[f][c] = floorRequests[c].active
		}
	}

	return result

}

func addRequests(e elevators) [][2]Reqest {
	// initialize a 2D array of requests, one for each floor and direction
	reqs := make([][2]Reqest, numFloors)
	for i := range reqs {
		reqs[i] = [2]Reqest{
			{active: false, assignedTo: 0},
			{active: false, assignedTo: 0},
		}
	}
	// add the requests from the hall buttons
	for f, floorRequests := range e.requests {
		for c := range 2 {
			reqs[f][c].active = floorRequests[c]
		}
	}
	return reqs
}
func initialStates(e elevators) []State {
	states := make([]State, len(e.states))
	for i, elevator := range e.states {
		states[i] = State{
			ElevatorState: elevator.ElevatorState,
			CabRequests:   make([]bool, numFloors),
			time:          time.Now(),
		}
	}
	return states
}
func performInitialMove(s *State, req *[][2]Reqest) {
	switch s.Behavior {
	case models.DoorOpen: // if the elevator is at a floor with the door open, wait for it to close
		s.time = s.time.Add(doorOpenDuration / 2)
		s.Behavior = models.Idle
		fallthrough
	case models.Idle: // if the elevator is idle, move to the first floor with a request
		for c := 0; c < 2; c++ {
			if (*req)[s.Floor][c].active {
				(*req)[s.Floor][c].active = false
				s.time = s.time.Add(doorOpenDuration)
			}
		}
	case models.Moving:
		s.Floor += int(s.Direction)
		s.time = s.time.Add(travelDuration / 2)
	}
}

func performSingleMove(s *State, req *[][2]Reqest) {

	e := anyUnassignedElevator(s, req)

	onClearRequest := func(c models.ButtonType) {
		switch c {
		case models.HallUp, models.HallDown:
			(*req)[s.Floor][c].assignedTo = s.Id
		case models.Cab:
			s.CabRequests[s.Floor] = false
		}
	}

	switch s.Behavior {
	case models.Moving:
		if shouldStop(e) {
			s.Behavior = models.DoorOpen
			s.time = s.time.Add(doorOpenDuration)
			clearReqsAtFloor(e, onClearRequest)
		} else {
			s.Floor += int(s.Direction)
			s.time = s.time.Add(travelDuration)
		}
	case models.Idle, models.DoorOpen:
		s.Direction = chooseDirection(e)
		if s.Direction == models.Stop {
			if anyRequestsAtFloor(e) {
				s.Behavior = models.DoorOpen
				s.time = s.time.Add(doorOpenDuration)
				clearReqsAtFloor(e, onClearRequest)
			} else {
				s.Behavior = models.Idle
			}
		} else {
			s.Behavior = models.Moving
			s.time = s.time.Add(travelDuration)
			s.Floor += int(s.Direction)
		}

	}
}
func anyUnassignedElevator(s *State, req *[][2]Reqest) localElevator {
	e := localElevator{
		ElevatorState: s.ElevatorState,
		requests:      make([][3]bool, numFloors),
	}

	for f, floorRequests := range *req {
		for c, req := range floorRequests {
			if req.active && req.assignedTo == 0 {
				e.requests[f][c] = true
			}
		}
	}
	return e
}

// no remaining cab requests, no floors with multiple hall requests, and all *unvisited* hall requests are at floors with elevators
func unvisitedAreImmediatelyAssignable(reqs [][2]Reqest, states []State) bool {
	for _, state := range states {
		if any(state.CabRequests) {
			return false
		}
	}
	for f, reqsAtFloor := range reqs {
		activeCount := 0
		for _, req := range reqsAtFloor {
			if req.active {
				activeCount++
			}
		}
		if activeCount == 2 {
			return false
		}
		for _, req := range reqsAtFloor {
			if req.active && req.assignedTo == 0 {
				found := false
				for _, state := range states {
					if state.Floor == f && !any(state.CabRequests) {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		}
	}
	return true
}
func assignImmediate(reqs *[][2]Reqest, states *[]State) {
	for f, reqsAtFloor := range *reqs {
		for _, req := range reqsAtFloor {
			for i := range *states {
				s := &(*states)[i]
				if req.active && req.assignedTo == 0 {
					if s.Floor == f && !any(s.CabRequests) {
						req.assignedTo = models.Id(s.Id)
						s.time = s.time.Add(doorOpenDuration)
					}
				}
			}
		}
	}
}
func any(arr []bool) bool {
	for _, req := range arr {
		if req {
			return true
		}
	}
	return false
}

// helper function to check if any element in a slice of Reqest is unassigned
func anyUnassigned(reqs [][2]Reqest) bool {
	for _, floorReqs := range reqs {
		for _, req := range floorReqs {
			if req.active && req.assignedTo == 0 {
				return true
			}
		}
	}
	return false
}
