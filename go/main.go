package main

import (
	"fmt"
	"math/rand"
	"time"
)

// CONSTANTS
const numCustomers = 20
const waitCapacity = 5
const custArrivMin = 500 // ms
const custArrivMax = 2000 // ms
const cutDurMin = 1000 // ms
const cutDurMax = 4000 // ms
const satThresh = 3.0 // sec per star lost
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
	MsgStartHaircut
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

func clamp(lo, hi, v int) int {
    if v < lo {
        return lo
    }
    if v > hi {
        return hi
    }
    return v
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
				// todo: log admission + queue depth here
			}

		case MsgNextCustomer:
			if len(cust_queue) == 0 {
				barberChannel <- Message{
					Kind: MsgNoneWaiting,
				}
			} else {
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
	arriveTime := time.Now().UnixMilli()
	fmt.Printf("Customer %d entered waiting room at time %d\n", id, arriveTime)
	mailbox := make(chan Message) // todo: should customer channel creation be here, or in shop owner? not sure
	waitRoomChannel <- Message{
		Kind: MsgArrive,
		From: mailbox,
		CustomerID: id, 
		ArrivalMs: arriveTime,
	}

	for {
		msg := <- mailbox
		switch msg.Kind {
		case MsgTurnedAway:
			fmt.Printf("Customer %d turned away\n", id)
			return
		case MsgAdmitted:
			fmt.Printf("Customer %d admitted to waiting room", id)
			cutStartMsg := <- mailbox // wait for haircut to start
			if cutStartMsg.Kind != MsgStartHaircut {
				fmt.Printf("Error: customer %d received unexpected message instead of MsgStartHaircut", id)
				return
			}
			cutStartTime := time.Now().UnixMilli()
			fmt.Printf("Starting haircut for customer %d", id)
			wait := arriveTime - cutStartTime
			jitter := rand.Intn(3) - 1 // {-1, 0, +1}
			score := clamp(1, 5, int(5-(wait/satThresh))+jitter)
			
			rateReqMsg := <- mailbox
			if rateReqMsg.Kind != MsgRateRequest {
				fmt.Printf("Error: customer %d received unexpected message instead of MsgRateRequest", id)
				return
			}
			rateReqMsg.From <- Message{
				Kind: MsgRating,
				Value: score,
			}
			return
			}
		}	

	}

// todoo ??? idk man.
func barber() {
	// send message to waiting room
	// todo: use buffered channel
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