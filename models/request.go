package models

type RequestMessage struct {
	Source  Id
	Request Request
}
type Request struct {
	Origin Origin
	Status RequestStatus
}
type Origin struct {
	Source     Source
	Floor      int
	ButtonType ButtonType
}
type Source interface {
	isSource()
}

type Hall struct{}

func (Hall) isSource() {}

type Elevator struct {
	Id Id
}

func (Elevator) isSource() {}

type Id uint8

type RequestStatus int

const (
	Absent RequestStatus = iota
	Unconfirmed
	Confirmed
	Unknown
)

func NewHallRequestMsg(peer Id, floor int, buttonType ButtonType, status RequestStatus) RequestMessage {
	return RequestMessage{
		Source: peer,
		Request: Request{
			Origin: Origin{
				Source:     Hall{},
				Floor:      floor,
				ButtonType: buttonType,
			},
			Status: status,
		},
	}
}

func NewCabRequestMsg(peer Id, elevator Id, floor int, status RequestStatus) RequestMessage {
	return RequestMessage{
		Source: peer,
		Request: Request{
			Origin: Origin{
				Source:     Elevator{Id: elevator},
				Floor:      floor,
				ButtonType: Cab,
			},
			Status: status,
		},
	}
}
