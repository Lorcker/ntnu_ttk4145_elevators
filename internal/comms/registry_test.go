package comms

import (
	"reflect"
	"testing"

	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/message"
	"group48.ttk4145.ntnu/elevators/internal/models/request"
)

func TestRegistryDiff(t *testing.T) {
	var tests = []struct {
		name     string
		internal requestRegistry
		external requestRegistry
		peer     elevator.Id
		expected []message.RequestState
	}{
		{
			name: "Discovered bug 1",
			internal: requestRegistry{
				HallUp:   []request.Status{0, 1, 1, 1},
				HallDown: []request.Status{0, 1, 1, 1},
				Cab: map[string][]request.Status{
					"1": []request.Status{0, 1, 1, 0},
					"2": []request.Status{0, 1, 1, 1},
				},
			},
			external: requestRegistry{
				HallUp:   []request.Status{0, 3, 1, 1},
				HallDown: []request.Status{0, 1, 1, 1},
				Cab: map[string][]request.Status{
					"1": []request.Status{0, 0, 1, 0},
					"2": []request.Status{0, 1, 1, 1},
				},
			},
			peer:     2,
			expected: []message.RequestState{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := tt.internal.diff(tt.peer, tt.external)
			if !reflect.DeepEqual(diff, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, diff)
			}
		})
	}
}
