package orderserver

import (
	"group48.ttk4145.ntnu/elevators/statedataserver"
)

type Orders [][]bool

func PollOrders(receiver chan<- Orders) {

}

func CalculateOrders(globalState <-chan statedataserver.GlobalState) {

}
