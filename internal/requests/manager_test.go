package requests

import (
	"testing"

	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/message"
	"group48.ttk4145.ntnu/elevators/internal/models/request"
)

func TestRequestManager_OnePeerHallCycle(t *testing.T) {
	// Test that the request manager processes a cycle of hall request messages from one peer

	// Setup
	var rm = newRequestManager(elevator.Id(1))
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

	// As there is only one peer, and the request was a hall request, it should not be confirmed immediately
	// Otherwise the button light contact would be violated if the local elevator crashes
	msg.Request.Status = request.Unconfirmed
	expected.Status = request.Unconfirmed
	res := rm.process(msg)
	if res != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg.Request.Origin])
	}

	// Should stay unconfirmed when the peer changes the request to absent (should be impossible)
	msg.Request.Status = request.Absent
	expected.Status = request.Unconfirmed
	rm.process(msg)

	if rm.store[msg.Request.Origin] != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg.Request.Origin])
	}
}

func TestRequestManager_OnePeerCabCycle(t *testing.T) {
	// Test that the request manager processes a cycle of cab request messages from one peer

	// Setup
	var rm = newRequestManager(elevator.Id(1))
	rm.alivePeers = []elevator.Id{1}
	var req = request.Request{Origin: request.Cab{Id: 1, Floor: 1}, Status: request.Unknown}
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
	// This is needed to ensure that cab requests are always handled even if the local elevator crashes
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

func TestRequestManager_TwoPeerHallCycle(t *testing.T) {
	// Test that the request manager processes a cycle of hall request messages from two peers

	// Setup
	var rm = newRequestManager(elevator.Id(1))
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

func TestRequestManager_TwoPeerCabCycle(t *testing.T) {
	// Test that the request manager processes a cycle of cab request messages from two peers

	// Setup
	var rm = newRequestManager(elevator.Id(1))
	rm.alivePeers = []elevator.Id{1, 2}
	var req = request.Request{Origin: request.Cab{Id: 1, Floor: 1}, Status: request.Unknown}
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

func TestRequestManager_NoUpdateFromAbsentToConfirmed(t *testing.T) {
	// Test that the request manager does not update a request from absent to confirmed

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

	// Should not change from absent to confirmed
	msg.Request.Status = request.Confirmed
	expected.Status = request.Absent
	rm.process(msg)
	if rm.store[msg.Request.Origin] != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg.Request.Origin])
	}

}

// Test with two peers
// Peer 1 is the local and has stored a unconfirmed request
// Peer 2 sends a request with status confirmed
// The request should be updated to confirmed as the other peer could only have confirmed when all peers acknowledged
func TestRequestManager_TwoPeerHallConfirmed(t *testing.T) {
	// Setup
	var rm = newRequestManager(elevator.Id(1))
	rm.alivePeers = []elevator.Id{1, 2}
	var req = request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unconfirmed}
	var msg1 = message.RequestStateUpdate{Source: 1, Request: req}
	var msg2 = message.RequestStateUpdate{Source: 2, Request: req}
	var expected = req

	// With an unknown request, the request should be stored as is
	expected.Status = request.Unconfirmed
	rm.process(msg1)
	if rm.store[msg1.Request.Origin] != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg1.Request.Origin])
	}

	// When the second peer changes the request to confirmed, the request should be confirmed
	msg2.Request.Status = request.Confirmed
	expected.Status = request.Confirmed
	rm.process(msg2)
	if rm.store[msg1.Request.Origin] != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg1.Request.Origin])
	}
}

// Test with two peers
// Peer 1 is the local and has stored a confirmed request
// Peer 2 sends a request with status unconfirmed
// The request should stay confirmed
func TestRequestManager_TwoPeerHallUnconfirmed(t *testing.T) {
	// Setup
	var rm = newRequestManager(elevator.Id(1))
	rm.alivePeers = []elevator.Id{1, 2}
	var req = request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Confirmed}
	var msg1 = message.RequestStateUpdate{Source: 1, Request: req}
	var msg2 = message.RequestStateUpdate{Source: 2, Request: req}
	var expected = req

	// With an unknown request, the request should be stored as is
	expected.Status = request.Confirmed
	rm.process(msg1)
	if rm.store[msg1.Request.Origin] != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg1.Request.Origin])
	}

	// When the second peer changes the request to unconfirmed, the request should stay confirmed
	msg2.Request.Status = request.Unconfirmed
	expected.Status = request.Confirmed
	rm.process(msg2)
	if rm.store[msg1.Request.Origin] != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg1.Request.Origin])
	}
}

// Test with three peers
// Peer 1 is the local and has stored a unconfirmed request
// Peer 2 sends a request with status unconfirmed
// local should stay unconfirmed as not all peers have confirmed
// Peer 3 sends a request with status unconfirmed
// local should change to confirmed as all peers have confirmed
func TestRequestManager_ThreePeerHallConfirmed(t *testing.T) {
	// Setup
	var rm = newRequestManager(elevator.Id(1))
	rm.alivePeers = []elevator.Id{1, 2, 3}
	var req = request.Request{Origin: request.Hall{Floor: 1, Direction: request.Up}, Status: request.Unconfirmed}
	var msg1 = message.RequestStateUpdate{Source: 1, Request: req}
	var msg2 = message.RequestStateUpdate{Source: 2, Request: req}
	var msg3 = message.RequestStateUpdate{Source: 3, Request: req}
	var expected = req

	// With an unknown request, the request should be stored as is
	expected.Status = request.Unconfirmed
	rm.process(msg1)
	if rm.store[msg1.Request.Origin] != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg1.Request.Origin])
	}

	// When the second peer changes the request to unconfirmed, the request should stay unconfirmed
	msg2.Request.Status = request.Unconfirmed
	expected.Status = request.Unconfirmed
	rm.process(msg2)
	if rm.store[msg1.Request.Origin] != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg1.Request.Origin])
	}

	// When the third peer changes the request to unconfirmed, the request should be confirmed
	msg3.Request.Status = request.Unconfirmed
	expected.Status = request.Confirmed
	rm.process(msg3)
	if rm.store[msg1.Request.Origin] != expected {
		t.Errorf("Expected %v, got %v", expected, rm.store[msg1.Request.Origin])
	}
}
