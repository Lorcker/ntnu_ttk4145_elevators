package requests

import m "group48.ttk4145.ntnu/elevators/models"

func isConfirmed(ledgers map[m.Id]bool, alive []m.Id) bool {
	if len(ledgers) != len(alive) {
		return false
	}

	for _, id := range alive {
		if _, ok := ledgers[id]; !ok {
			return false
		}
	}

	return true
}
