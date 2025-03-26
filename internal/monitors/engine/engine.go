package enginemonitor

import (
	"log"
	"time"

	"group48.ttk4145.ntnu/elevators/internal/elevatorio"
	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/message"
)

const engineTimeout = time.Second * 10

func RunEngineMonitor(local elevator.Id,
	fFromElevio <-chan message.FloorArrival,
	bFromDriver <-chan message.ElevatorState,
	toHealthMonitor chan<- message.PeerSignal) {

	engineTimer := time.NewTimer(engineTimeout)
	engineTimer.Stop()

	isDead := false
	lastBeh := elevator.Idle
	lasDir := elevator.Stop
	shouldMove := false

	for {
		select {
		case <-fFromElevio:
			if isDead {
				toHealthMonitor <- message.PeerSignal{Id: local, Alive: true}
				isDead = false
				log.Printf("[enginemonitor] The motor is alive aggain!")
			}
			if shouldMove {
				engineTimer.Reset(engineTimeout)
			}
		case msg := <-bFromDriver:
			lasDir = msg.State.Direction

			current := msg.State.Behavior

			if lastBeh != elevator.Moving && current == elevator.Moving {
				engineTimer.Reset(engineTimeout)
				shouldMove = true
			}
			if lastBeh == elevator.Moving && current != elevator.Moving {
				engineTimer.Stop()
				shouldMove = false
			}

			lastBeh = current
		case <-engineTimer.C:
			toHealthMonitor <- message.PeerSignal{Id: local, Alive: false}
			isDead = true
			log.Print("[enginemotor] The motor died. Trying to move until power is restored.")
			tryMoving(lasDir)
		}
	}
}

func tryMoving(dir elevator.MotorDirection) {
	current := elevatorio.GetFloor()
	for elevatorio.GetFloor() == current {
		elevatorio.SetMotorDirection(dir)
		time.Sleep(time.Millisecond * 100)
	}
}
