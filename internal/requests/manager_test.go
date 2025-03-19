package requests

import (
	"testing"

	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/message"
	"group48.ttk4145.ntnu/elevators/internal/models/request"
)

func TestRequestManager_OnePeerCycle(t *testing.T) {
	// Test that the request manager processes a cycle of messages from one peer

	// Setup
	var rm = newRequestManager(elevator.Id(0))
	rm.alivePeers = []elevator.Id{1}
	var req = request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unknown}

	var msg = message.RequestStateUpdate{Source: 1, Request: req}
	var expected = req

	// With an unknown request, the request should be stored as is
	expected.Status = request.Unknown
	rm.process(msg)
	if rm.store[msg.Request.Origin] != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg.Request.Origin])
	}

	// Should change from unknown to absent
	msg.Request.Status = request.Absent
	expected.Status = request.Absent
	rm.process(msg)
	if rm.store[msg.Request.Origin] != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg.Request.Origin])
	}

	// As there is only one peer, the request should be confirmed immediately
	msg.Request.Status = request.Unconfirmed
	expected.Status = request.Confirmed
	res := rm.process(msg)
	if res != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg.Request.Origin])
	}

	// Should change from confirmed to absent
	msg.Request.Status = request.Absent
	expected.Status = request.Absent
	rm.process(msg)

	if rm.store[msg.Request.Origin] != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg.Request.Origin])
	}
}

func TestRequestManager_OnePeerFirstUnconfirmed(t *testing.T) {
	// Test that the request manager processes an unconfirmed request from one peer without a previous request correctly

	// Setup
	var rm = newRequestManager(elevator.Id(0))
	rm.alivePeers = []elevator.Id{1}
	var req = request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unknown}
	var msg = message.RequestStateUpdate{Source: 1, Request: req}
	var expected = req

	// As there is only one peer, the request should be confirmed immediately
	msg.Request.Status = request.Unconfirmed
	expected.Status = request.Confirmed
	res := rm.process(msg)
	if res != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg.Request.Origin])
	}

	// Should change from confirmed to absent
	msg.Request.Status = request.Absent
	expected.Status = request.Absent
	rm.process(msg)

	if rm.store[msg.Request.Origin] != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg.Request.Origin])
	}
}

func TestRequestManager_TwoPeerCycle(t *testing.T) {
	// Test that the request manager processes a cycle of messages from two peers

	// Setup
	var rm = newRequestManager(elevator.Id(0))
	rm.alivePeers = []elevator.Id{1, 2}
	var req = request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unknown}
	var msg1 = message.RequestStateUpdate{Source: 1, Request: req}
	var msg2 = message.RequestStateUpdate{Source: 2, Request: req}
	var expected = req

	// With an unknown request, the request should be stored as is
	expected.Status = request.Unknown
	rm.process(msg1)
	rm.process(msg2)
	if rm.store[msg1.Request.Origin] != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg1.Request.Origin])
	}

	// When the first peer changes the request to unconfirmed, the request should stay unconfirmed
	msg1.Request.Status = request.Unconfirmed
	expected.Status = request.Unconfirmed
	rm.process(msg1)
	if rm.store[msg1.Request.Origin] != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg1.Request.Origin])
	}

	// When the second peer changes the request to unconfirmed, the request should be confirmed
	msg2.Request.Status = request.Unconfirmed
	expected.Status = request.Confirmed
	rm.process(msg2)
	if rm.store[msg1.Request.Origin] != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg1.Request.Origin])
	}

	// When the first peer changes the request to absent, the request should change to absent
	msg1.Request.Status = request.Absent
	expected.Status = request.Absent
	rm.process(msg1)
	if rm.store[msg1.Request.Origin] != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg1.Request.Origin])
	}
}
