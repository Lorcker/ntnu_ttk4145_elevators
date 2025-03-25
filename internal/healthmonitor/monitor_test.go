package healthmonitor

import (
	"testing"
	"time"

	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/message"
)

func TestUpdateAliveList(t *testing.T) {
	localID := elevator.Id(0)
	peerID1 := elevator.Id(1)
	peerID2 := elevator.Id(2)

	tests := []struct {
		name       string
		lastSeen   lastSeen
		alivePeers alivePeers
		expected   alivePeers
		changed    bool
	}{
		{
			name: "No peers, only local alive",
			lastSeen: lastSeen{
				localID: time.Now(),
			},
			alivePeers: alivePeers{},
			expected: alivePeers{
				localID: true,
			},
			changed: true,
		},
		{
			name: "One peer alive",
			lastSeen: lastSeen{
				localID: time.Now(),
				peerID1: time.Now(),
			},
			alivePeers: alivePeers{},
			expected: alivePeers{
				localID: true,
				peerID1: true,
			},
			changed: true,
		},
		{
			name: "One peer dead",
			lastSeen: lastSeen{
				localID: time.Now(),
				peerID1: time.Now().Add(-Timeout * 2),
			},
			alivePeers: alivePeers{
				peerID1: true,
			},
			expected: alivePeers{
				localID: true,
			},
			changed: true,
		},
		{
			name: "No change in alive peers",
			lastSeen: lastSeen{
				localID: time.Now(),
				peerID1: time.Now(),
			},
			alivePeers: alivePeers{
				localID: true,
				peerID1: true,
			},
			expected: alivePeers{
				localID: true,
				peerID1: true,
			},
			changed: false,
		},
		{
			name: "Multiple peers, some alive, some dead",
			lastSeen: lastSeen{
				localID: time.Now(),
				peerID1: time.Now(),
				peerID2: time.Now().Add(-Timeout * 2),
			},
			alivePeers: alivePeers{
				peerID2: true,
			},
			expected: alivePeers{
				localID: true,
				peerID1: true,
			},
			changed: true,
		},
		{
			name: "Local dead",
			lastSeen: lastSeen{
				localID: time.Now().Add(-Timeout * 2),
				peerID1: time.Now(),
				peerID2: time.Now(),
			},
			alivePeers: alivePeers{
				localID: true,
				peerID1: true,
				peerID2: true,
			},
			expected: alivePeers{
				peerID1: true,
				peerID2: true,
			},
			changed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed := updateAliveList(tt.lastSeen, tt.alivePeers, localID, true)
			if changed != tt.changed {
				t.Errorf("expected changed to be %v, got %v", tt.changed, changed)
			}
			for id, alive := range tt.expected {
				if tt.alivePeers[id] != alive {
					t.Errorf("expected peer %v to be %v, got %v", id, alive, tt.alivePeers[id])
				}
			}
		})
	}
}

func TestRunHealthMonitor(t *testing.T) {
	localID := elevator.Id(0)
	peerID1 := elevator.Id(1)
	peerID2 := elevator.Id(2)

	pingFromComms := make(chan message.PeerSignal)
	alivenessToOrders := make(chan message.ActivePeers)
	alivenessToRequests := make(chan message.ActivePeers)

	go func() {
		pingFromComms <- message.PeerSignal{Id: peerID1, Alive: true}
		pingFromComms <- message.PeerSignal{Id: peerID2, Alive: true}
		pingFromComms <- message.PeerSignal{Id: peerID1, Alive: false}
		pingFromComms <- message.PeerSignal{Id: peerID2, Alive: false}
	}()

	go RunMonitor(
		localID,
		pingFromComms,
		alivenessToOrders,
		alivenessToRequests,
	)

	time.Sleep(10000 * time.Millisecond)
}
