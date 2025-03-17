package comms

import (
	"testing"

	"group48.ttk4145.ntnu/elevators/internal/models"
)

func TestRegistryDiff(t *testing.T) {
	var tests = []struct {
		name     string
		internal RequestRegistry
		external RequestRegistry
		peer     models.Id
		expected []models.RequestMessage
	}{
		{
			name: "Discovered bug 1",
			internal: RequestRegistry{
				HallUp:   []models.RequestStatus{0, 1, 1, 1},
				HallDown: []models.RequestStatus{0, 1, 1, 1},
				Cab: map[string][]models.RequestStatus{
					"1": []models.RequestStatus{0, 1, 1, 0},
					"2": []models.RequestStatus{0, 1, 1, 1},
				},
			},
			external: RequestRegistry{
				HallUp:   []models.RequestStatus{0, 3, 1, 1},
				HallDown: []models.RequestStatus{0, 1, 1, 1},
				Cab: map[string][]models.RequestStatus{
					"1": []models.RequestStatus{0, 0, 1, 0},
					"2": []models.RequestStatus{0, 1, 1, 1},
				},
			},
			peer: 2,
			expected: []models.RequestMessage{
				{
					Request: models.Request{
						Origin: models.Origin{
							Source:     models.Hall{},
							Floor:      1,
							ButtonType: models.HallUp,
						},
						Status: models.Confirmed,
					},
					Source: 2,
				},
				{
					Request: models.Request{
						Origin: models.Origin{
							Source:     models.Elevator{Id: 1},
							Floor:      1,
							ButtonType: models.Cab,
						},
						Status: models.Unknown,
					},
					Source: 2,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := tt.internal.Diff(tt.peer, tt.external)
			if !equalRequestMessages(diff, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, diff)
			}
		})
	}
}

func equalRequestMessages(a, b []models.RequestMessage) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
