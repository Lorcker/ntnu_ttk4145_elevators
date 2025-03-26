package obstructionmonitor

import (
	"log"
	"time"

	"group48.ttk4145.ntnu/elevators/internal/models/elevator"
	"group48.ttk4145.ntnu/elevators/internal/models/message"
)

const obstructionTimeout = time.Second * 10

func RunObstructionMonitor(local elevator.Id,
	oFromElevio <-chan message.Obstruction,
	toHealthMonitor chan<- message.PeerSignal) {

	obstructionTimer := time.NewTimer(obstructionTimeout)
	obstructionTimer.Stop()

	isDead := false
	isObstructed := false

	for {
		select {
		case <-oFromElevio:
			if isObstructed && isDead {
				toHealthMonitor <- message.PeerSignal{Id: local, Alive: true}
				isDead = false
				log.Printf("[enginemonitor] The obstruction has been cleared!")
			}

			if isObstructed {
				obstructionTimer.Stop()
			} else {
				obstructionTimer.Reset(obstructionTimeout)
			}

			isObstructed = !isObstructed
		case <-obstructionTimer.C:
			toHealthMonitor <- message.PeerSignal{Id: local, Alive: false}
			isDead = true
			log.Print("[enginemotor] The elevator is currently permantly obstructed. We are considered dead")
		}
	}
}
