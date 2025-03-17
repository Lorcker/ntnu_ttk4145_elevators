package orderserver

import (
	"reflect"
	"testing"

	m "group48.ttk4145.ntnu/elevators/internal/models"
)

func Test_calculateOrders(t *testing.T) {
	type args struct {
		hr        HallRequests
		cr        map[m.Id]CabRequests
		elevators map[m.Id]m.ElevatorState
	}
	tests := []struct {
		name string
		args args
		want map[m.Id]m.Orders
	}{
		{
			name: "Test 1",
			args: args{
				hr: HallRequests{
					[2]bool{true, false},
					[2]bool{false, true},
					[2]bool{true, false},
					[2]bool{false, true},
				},
				cr: map[m.Id]CabRequests{
					1: []bool{true, false, true, false},
					2: []bool{false, true, false, true},
				},
				elevators: map[m.Id]m.ElevatorState{
					1: {
						Id:        1,
						Behavior:  m.Idle,
						Floor:     0,
						Direction: m.Up,
					},
					2: {
						Id:        2,
						Behavior:  m.Moving,
						Floor:     1,
						Direction: m.Down,
					},
				},
			},
			want: map[m.Id]m.Orders{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateOrders(tt.args.hr, tt.args.cr, tt.args.elevators); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateOrders() = %v, want %v", got, tt.want)
			}
		})
	}
}
