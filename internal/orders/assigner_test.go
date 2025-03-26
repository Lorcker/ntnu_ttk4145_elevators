package orders

import (
	"reflect"
	"testing"

	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
)

func Test_calculateOrders(t *testing.T) {
	type args struct {
		hr        hallRequests
		cr        map[elevator.Id]cabRequests
		elevators map[elevator.Id]elevator.State
		alive     map[elevator.Id]bool
	}
	tests := []struct {
		name string
		args args
		want map[elevator.Id]elevator.Order
	}{
		{
			name: "Test 1",
			args: args{
				hr: hallRequests{
					[2]bool{true, false},
					[2]bool{false, true},
					[2]bool{true, false},
					[2]bool{false, true},
				},
				cr: map[elevator.Id]cabRequests{
					1: {true, false, true, false},
					2: {false, true, false, true},
				},
				elevators: map[elevator.Id]elevator.State{
					1: {
						Behavior:  elevator.Idle,
						Floor:     0,
						Direction: elevator.Up,
					},
					2: {
						Behavior:  elevator.Moving,
						Floor:     1,
						Direction: elevator.Down,
					},
				},
				alive: map[elevator.Id]bool{
					1: true,
					2: true,
				},
			},
			want: map[elevator.Id]elevator.Order{},
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
