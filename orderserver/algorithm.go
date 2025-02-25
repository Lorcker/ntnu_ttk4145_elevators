package orderserver

import (
	"slices"

	"group48.ttk4145.ntnu/elevators/models"
)

func requestsAbove(e localElevator) bool {
	for _, floorRequests := range e.requests[e.Floor+1:] {
		if any(floorRequests[:]) {
			return true
		}
	}
	return false
}

func requestsBelow(e localElevator) bool {
	for _, floorRequests := range e.requests[:e.Floor] {
		if any(floorRequests[:]) {
			return true
		}
	}
	return false
}

func anyRequests(e localElevator) bool {
	for _, floorRequests := range e.requests {
		if slices.Contains(floorRequests[:], true) {
			return true
		}
	}
	return false
}

func anyRequestsAtFloor(e localElevator) bool {
	for _, request := range e.requests[e.Floor] {
		if request {
			return true
		}
	}
	return false
}

func shouldStop(e localElevator) bool {
	switch e.Direction {
	case models.Up:
		return any(e.requests[e.Floor][:]) || !requestsAbove(e) || e.Floor == 0 || e.Floor == len(e.requests)-1
	case models.Down:
		return any(e.requests[e.Floor][:]) || !requestsBelow(e) || e.Floor == 0 || e.Floor == len(e.requests)-1
	case models.Stop:
		return true
	}
	return false
}

func chooseDirection(e localElevator) models.MotorDirection {
	switch e.Direction {
	case models.Up:
		if requestsAbove(e) {
			return models.Up
		} else if anyRequestsAtFloor(e) {
			return models.Stop
		} else if requestsBelow(e) {
			return models.Down
		} else {
			return models.Stop
		}
	case models.Down, models.Stop:
		if requestsBelow(e) {
			return models.Down
		} else if anyRequestsAtFloor(e) {
			return models.Stop
		} else if requestsAbove(e) {
			return models.Up
		} else {
			return models.Stop
		}
	}
	return models.Stop
}

func clearReqsAtFloor(e localElevator, onClearedRequest func(models.ButtonType)) localElevator {
	e2 := e

	clear := func() {
		if slices.Contains(e2.requests[e2.Floor][:], true) {
			if onClearedRequest != nil {
				onClearedRequest(models.Cab)
			}
			e2.requests[e2.Floor] = [3]bool{false, false, false}
		}
	}

	clearRequestType := "all" // or "inDirn", depending on your logic

	switch clearRequestType {
	case "all":
		clear()
	case "inDirn":
		clear()
	}

	return e2
}
