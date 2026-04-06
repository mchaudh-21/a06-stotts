package main

// CONSTANTS
const numCustomers = 20
const waitCapacity = 5
const custArrivMin = 500 // ms
const custArrivMax = 2000 // ms
const cutDurMin = 1000 // ms
const cutDurMax = 4000 // ms
const satisfactionWaitThresh = 3.0 // sec per star lost
const gracePeriod = 8000 // ms

// Universal message type
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
    MsgWRStatsReply
	MsgBarberStatsReply
    MsgShutdown
)
type Message struct {
	Kind MsgKind
	From chan Message // reply-to channel
	CustomerID int
	Value int	// payload
	ArrivalMs int64 // arrival timestamp
	// Waiting Room stats
	QueueLengthStat int 
	TurnawayCountStat int
	// Barber stats 
	CutCountStat int
	AvgCutDurStat float64
	AvgRatingStat float64

}

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

func waitingRoom(mailbox chan Message, barberChannel chan Message) {
	// todo: clarify when to send wakeup signal; ensure is sent only once per cycle
	turnaway_count := 0
	cust_queue := make([]int, 0, waitCapacity)
	barberIsSleeping := false
	for {
		msg := <- mailbox
		switch msg.Kind {
		case MsgArrive:
			if len(cust_queue) >= waitCapacity {
				msg.From <- Message{
					Kind: MsgTurnedAway,
				}
				turnaway_count += 1
			} else {
				if barberIsSleeping {
					barberChannel <- Message{
						Kind: MsgWakeUp,
					}
					barberIsSleeping = false
				}
				cust_queue = append(cust_queue, msg.CustomerID)
				msg.From <- Message{
					Kind: MsgAdmitted,
				}
			}

		case MsgNextCustomer:
			if len(cust_queue) == 0 {
				barberChannel <- Message{
					Kind: MsgNoneWaiting,
				}
			} else {
				if barberIsSleeping {
					barberChannel <- Message{
						Kind: MsgWakeUp,
					}
				}
				next_customer := cust_queue[0]
				cust_queue = cust_queue[1:]

				barberChannel <- Message{
					Kind: MsgCustomerReady,
					CustomerID: next_customer,
				}
			}
		
		case MsgGetStats:
			msg.From <- Message{
				Kind: MsgWRStatsReply,
				QueueLengthStat: len(cust_queue),
				TurnawayCountStat: turnaway_count,
			}
		
		case MsgShutdown:
			// todo: shutdown
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