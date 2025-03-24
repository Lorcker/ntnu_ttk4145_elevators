package requests

import (
	"testing"

	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/message"
	"group48.ttk4145.ntnu/elevators/internal/models/request"
)

func TestRequestManager(t *testing.T) {
	tests := []struct {
		name           string
		alivePeers     []elevator.Id
		initialRequest request.Request
		updates        []struct {
			update         message.RequestState
			expectedStatus request.Status
		}
	}{
		{
			name:           "OnePeerHallCycle",
			alivePeers:     []elevator.Id{1},
			initialRequest: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unknown},
			updates: []struct {
				update         message.RequestState
				expectedStatus request.Status
			}{
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unknown}}, expectedStatus: request.Unknown},
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Absent}}, expectedStatus: request.Absent},
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unconfirmed}}, expectedStatus: request.Unconfirmed},
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Absent}}, expectedStatus: request.Unconfirmed},
			},
		},
		{
			name:           "OnePeerCabCycle",
			alivePeers:     []elevator.Id{1},
			initialRequest: request.Request{Origin: request.Cab{Id: 1, Floor: 1}, Status: request.Unknown},
			updates: []struct {
				update         message.RequestState
				expectedStatus request.Status
			}{
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Cab{Id: 1, Floor: 1}, Status: request.Unknown}}, expectedStatus: request.Unknown},
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Cab{Id: 1, Floor: 1}, Status: request.Absent}}, expectedStatus: request.Absent},
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Cab{Id: 1, Floor: 1}, Status: request.Unconfirmed}}, expectedStatus: request.Confirmed},
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Cab{Id: 1, Floor: 1}, Status: request.Absent}}, expectedStatus: request.Absent},
			},
		},
		{
			name:           "TwoPeerHallCycle",
			alivePeers:     []elevator.Id{1, 2},
			initialRequest: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unknown},
			updates: []struct {
				update         message.RequestState
				expectedStatus request.Status
			}{
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unknown}}, expectedStatus: request.Unknown},
				{update: message.RequestState{Source: 2, Request: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unknown}}, expectedStatus: request.Unknown},
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unconfirmed}}, expectedStatus: request.Unconfirmed},
				{update: message.RequestState{Source: 2, Request: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unconfirmed}}, expectedStatus: request.Confirmed},
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Absent}}, expectedStatus: request.Absent},
			},
		},
		{
			name:           "TwoPeerCabCycle",
			alivePeers:     []elevator.Id{1, 2},
			initialRequest: request.Request{Origin: request.Cab{Id: 1, Floor: 1}, Status: request.Unknown},
			updates: []struct {
				update         message.RequestState
				expectedStatus request.Status
			}{
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Cab{Id: 1, Floor: 1}, Status: request.Unknown}}, expectedStatus: request.Unknown},
				{update: message.RequestState{Source: 2, Request: request.Request{Origin: request.Cab{Id: 1, Floor: 1}, Status: request.Unknown}}, expectedStatus: request.Unknown},
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Cab{Id: 1, Floor: 1}, Status: request.Unconfirmed}}, expectedStatus: request.Unconfirmed},
				{update: message.RequestState{Source: 2, Request: request.Request{Origin: request.Cab{Id: 1, Floor: 1}, Status: request.Unconfirmed}}, expectedStatus: request.Confirmed},
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Cab{Id: 1, Floor: 1}, Status: request.Absent}}, expectedStatus: request.Absent},
			},
		},
		{
			name:           "NoUpdateFromAbsentToConfirmed",
			alivePeers:     []elevator.Id{1},
			initialRequest: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unknown},
			updates: []struct {
				update         message.RequestState
				expectedStatus request.Status
			}{
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unknown}}, expectedStatus: request.Unknown},
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Absent}}, expectedStatus: request.Absent},
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Confirmed}}, expectedStatus: request.Absent},
			},
		},
		{
			name:           "TwoPeerHallConfirmed",
			alivePeers:     []elevator.Id{1, 2},
			initialRequest: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unconfirmed},
			updates: []struct {
				update         message.RequestState
				expectedStatus request.Status
			}{
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unconfirmed}}, expectedStatus: request.Unconfirmed},
				{update: message.RequestState{Source: 2, Request: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Confirmed}}, expectedStatus: request.Confirmed},
			},
		},
		{
			name:           "TwoPeerHallUnconfirmed",
			alivePeers:     []elevator.Id{1, 2},
			initialRequest: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Confirmed},
			updates: []struct {
				update         message.RequestState
				expectedStatus request.Status
			}{
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Confirmed}}, expectedStatus: request.Confirmed},
				{update: message.RequestState{Source: 2, Request: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unconfirmed}}, expectedStatus: request.Confirmed},
			},
		},
		{
			name:           "ThreePeerHallConfirmed",
			alivePeers:     []elevator.Id{1, 2, 3},
			initialRequest: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unconfirmed},
			updates: []struct {
				update         message.RequestState
				expectedStatus request.Status
			}{
				{update: message.RequestState{Source: 1, Request: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unconfirmed}}, expectedStatus: request.Unconfirmed},
				{update: message.RequestState{Source: 2, Request: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unconfirmed}}, expectedStatus: request.Unconfirmed},
				{update: message.RequestState{Source: 3, Request: request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unconfirmed}}, expectedStatus: request.Confirmed},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := newRequestManager(elevator.Id(1))
			rm.alivePeers = tt.alivePeers
			rm.statusByOrigin[tt.initialRequest.Origin] = tt.initialRequest.Status

			for _, u := range tt.updates {
				res := rm.Process(u.update)
				if got := res.Status; got != u.expectedStatus {
					t.Errorf("Expected %v, got %v", u.expectedStatus, got)
				}
			}
		})
	}
}
