package main

// CONSTANTS
const numCustomers = 20
const waitCapacity = 5
const custArrivalMin = 500 // ms
const custArrivalMax = 2000 // ms
const satisfactionWaitThresh = 3.0 // sec per star lost
const gracePeriod = 8000 // ms


// Common message type
type MsgKind int

const (
    MsgArrive MsgKind = iota
    MsgAdmitted
    MsgTurnedAway
    MsgNextCustomer
    MsgCustomerReady
    MsgNoneWaiting
    MsgWakeUp
    MsgRateRequest
    MsgRating
    MsgGetStats
    MsgStatsReply
    MsgShutdown
)

type Message struct {
	Kind MsgKind
	From chan Message // reply-to channel
	CustomerID int
	Value int	// rating, or other int payload
	ArrivalMs int64 // arrival timestamp
}


// 
func shopOwner() {
	// run program

	// start waiting room
	// start barber

	// spawn customers at random intervals

	// wait for grace period (2x max haircut duration)

	// send stats request to barber & waiting room, save responses
	// send shutdown to barber and waiting room
	// print closing report
}

func waitingRoom(mailbox chan Message) {
	// track whether barber is sleeping or not. when it sends MsgNoneWaiting, flip to True. when it sends MsgWakeUp, flip to false
	turnaway_count := 0
	cust_queue := 0
	barberIsSleeping := false
	for {
		msg := <- mailbox
		switch msg.Kind {
		case MsgArrive:
			// todo: if at capacity, send back MsgTurnedAway
			// else, add to queue and send back MsgAdmitted
			// if barber is sleeping: wake him up !!!
		 
		case MsgNextCustomer:
			// todo: if queue is empty, send MsgNoneWaiting to barber & flip flag
			// if queue is NOT empty: 
			// (1) if barber is sleeping, send MsgWakeUp
			// (2) pop first cust and send to barber with MsgCustomerReady
		
		case MsgGetStats:
			// todo: send stats to shop owner in StatsReply msg. include turnaway count & curr queue length
		
		case MsgShutdown:
			// shutdown
		}
	}

}

func customer(id int, waitRoomChannel chan Message) {
	// send arrive message to waiting room 
	// wait for response:
		// 1. turned away (log rejection; exit)
		// 2. admitted - continue to wait to be called by barber
			// when called, note time
			// aft
}

// todoo ??? idk man.
func barber() {
	// send message to waiting room
}


// todo: ?? idk
type BarberState int

const (
	Asleep BarberState = iota
	Awake
)

func (b *BarberState) sleepLoop(mailbox chan Message) {
	for {
		msg := <-mailbox
		switch msg.Kind {
			case MsgWakeUp:
				return // exit sleep loop, re-enter main
			case MsgShutdown: 
				return // todo: shutdown
		}
	}
}

func (b *BarberState) awakeLoop(mailbox chan Message) {

}