package main

import (
	"fmt"
	"math/rand"
	"time"
)

var programStart = time.Now()

// CONSTANTS
const numCustomers = 20
const waitCapacity = 5
const custArrivMin = 500 // ms
const custArrivMax = 2000 // ms
const cutDurMin = 1000 // ms
const cutDurMax = 4000 // ms
const satThresh = 3000 // ms per star lost
const gracePeriod = 5*cutDurMax // ms

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
	MsgShutdownAck
)
type Message struct {
	Kind MsgKind
	From chan Message // reply-to channel
	CustomerID int
	Value int	// payload
	ArrivalMs int64 // arrival timestamp
	Rating int // rating from customer
	WaitTime int64 // customer wait time (ms)
	// Waiting Room stats
	QueueLengthStat int 
	TurnawayCountStat int
	// Barber stats 
	CutCountStat int
	AvgWaitTime float64
	AvgRatingStat float64

}

type WaitingCustomer struct {
	ID int
	Mailbox chan Message
}

// Util
func clamp(lo, hi, v int) int {
    if v < lo {
        return lo
    }
    if v > hi {
        return hi
    }
    return v
}

func elapsedMs() int64 {
	return time.Since(programStart).Milliseconds()
}

func logf(actor string, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("[%d ms] [%s] %s\n", elapsedMs(), actor, msg)
}

func main() {
	shopOwner()
}

func shopOwner() {
	mailbox := make(chan Message)
	waitRoomChannel := make(chan Message)
	barberChannel := make(chan Message, 1)
	go waitingRoom(waitRoomChannel)
	go barber(barberChannel, waitRoomChannel)

	for i := 0; i < numCustomers; i++ {
		go customer(i, waitRoomChannel)
		logf("shop owner", "spawned customer %d", i)
		time.Sleep(time.Duration(rand.Intn(custArrivMax-custArrivMin)+custArrivMin)*time.Millisecond)
	}
	time.Sleep(gracePeriod)
	barberChannel <- Message{
		Kind: MsgGetStats,
		From: mailbox,
	}
	barberStatsReply := <- mailbox
	waitRoomChannel <- Message{
		Kind: MsgGetStats,
		From: mailbox,
	}
	waitRoomStatsReply := <- mailbox
	
	logf("shop owner", "initiating shutdown...")
	barberChannel <- Message{
		Kind: MsgShutdown,
		From: mailbox,
	}
	barberShutdownReply := <- mailbox
	if barberShutdownReply.Kind == MsgShutdownAck {
		logf("shop owner", "barber confirmed shutdown")
	}

	waitRoomChannel <- Message{
		Kind: MsgShutdown,
		From: mailbox,
	}
	waitRoomShutdownReply := <- mailbox
	if waitRoomShutdownReply.Kind == MsgShutdownAck {
		logf("shop owner", "waiting room confirmed shutdown")
	}
	cust_served := barberStatsReply.CutCountStat
	cust_rejected := waitRoomStatsReply.TurnawayCountStat
	avg_wait := barberStatsReply.AvgWaitTime
	avg_rating := barberStatsReply.AvgRatingStat

	fmt.Println("============= CLOSING REPORT ============")
	fmt.Printf("Total customers arrived: %d\n", numCustomers)
	fmt.Printf("Customers served: %d\n", cust_served)
	fmt.Printf("Customers turned away: %d\n", cust_rejected)
	fmt.Printf("Average wait duration (ms): %.2f\n", avg_wait)
	fmt.Printf("Average satisfaction rating: %.2f\n", avg_rating)
	fmt.Println("==========================================")
}


func waitingRoom(mailbox chan Message) {
	turnaway_count := 0
	cust_queue := make([]WaitingCustomer, 0, waitCapacity)
	barberIsSleeping := false
	var barberChannel chan Message
	for {
		msg := <- mailbox
		switch msg.Kind {
		case MsgNextCustomer:
			barberChannel = msg.From
			if len(cust_queue) == 0 {
				barberIsSleeping = true
				barberChannel <- Message{
					Kind: MsgNoneWaiting,
				}
			} else {
				next_customer := cust_queue[0]
				cust_queue = cust_queue[1:]
				barberChannel <- Message{
					Kind: MsgCustomerReady,
					CustomerID: next_customer.ID,
					From: next_customer.Mailbox,
				}
			}
		case MsgArrive:
			if len(cust_queue) >= waitCapacity {
				msg.From <- Message{
					Kind: MsgTurnedAway,
				}
				turnaway_count += 1
				logf("WR", "customer %d turned away. queue length: %d", msg.CustomerID, len(cust_queue))
			} else {
				if barberIsSleeping {
					barberChannel <- Message{
						Kind: MsgWakeUp,
					}
					barberIsSleeping = false
				}
				cust_queue = append(cust_queue, WaitingCustomer{ID: msg.CustomerID, Mailbox: msg.From})
				msg.From <- Message{
					Kind: MsgAdmitted,
				}
				logf("WR", "customer %d admitted. %d in queue", msg.CustomerID, len(cust_queue))
			}
		
		case MsgGetStats:
			msg.From <- Message{
				Kind: MsgWRStatsReply,
				QueueLengthStat: len(cust_queue),
				TurnawayCountStat: turnaway_count,
			}
		
		case MsgShutdown:
			if msg.From != nil {
				msg.From <- Message{Kind: MsgShutdownAck}
			}
			logf("WR", "shutting down")
			return
		}
	}

}

func customer(id int, waitRoomChannel chan Message) {
	actor := fmt.Sprintf("cust-%d", id)
	arriveTime := elapsedMs()
	logf(actor, "arrived at waiting room")
	mailbox := make(chan Message)
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
			return
		case MsgAdmitted:
			cutStartMsg := <- mailbox // wait for haircut to start
			if cutStartMsg.Kind != MsgStartHaircut {
				logf(actor, "error: received unexpected message instead of MsgStartHaircut")
				return
			}
			cutStartTime := elapsedMs()
			wait := cutStartTime - arriveTime
			jitter := rand.Intn(3) - 1 // {-1, 0, +1}
			score := clamp(1, 5, int(5-(wait/satThresh))+jitter)
			
			rateReqMsg := <- mailbox
			if rateReqMsg.Kind != MsgRateRequest {
				logf(actor, "error: received unexpected message instead of MsgRateRequest")
				return
			}
			logf(actor, "rates %d stars, waited %d ms", score, wait)
			rateReqMsg.From <- Message{
				Kind: MsgRating,
				Rating: score,
				WaitTime: wait,
			}
			return
			}
		}	

	}


func barber(mailbox chan Message, waitRoomChannel chan Message) {
	// main "awake" loop for Barber
	num_completed := 0
	var avg_wait float64 = 0
	var avg_rating float64 = 0
	requestNextCustomer := func() {
		waitRoomChannel <- Message{
			Kind: MsgNextCustomer,
			From: mailbox,
		}
	}
	requestNextCustomer()
	for {
		msg := <- mailbox
		switch msg.Kind {
		case MsgCustomerReady:
			msg.From <- Message{
				Kind: MsgStartHaircut,
			}
			cutLength := time.Duration(rand.Intn((cutDurMax-cutDurMin)) + cutDurMin) * time.Millisecond
			logf("barber", "starting haircut for customer %d", msg.CustomerID)
			time.Sleep(cutLength)
			logf("barber", "finished haircut for customer %d", msg.CustomerID)
			ratingReplyChannel := make(chan Message)
			msg.From <- Message{
				Kind: MsgRateRequest,
				From: ratingReplyChannel,
			}
			ratingResponse:= <- ratingReplyChannel
			num_completed += 1
			recordedWait := float64(ratingResponse.WaitTime)
			recordedRating := float64(ratingResponse.Rating)
			avg_wait += (recordedWait - avg_wait) / float64(num_completed)
			avg_rating += (recordedRating - avg_rating) / float64(num_completed)

			logf("barber", "average rating: %.2f stars | average wait time: %.2f ms", avg_rating, avg_wait)
			requestNextCustomer()

		case MsgNoneWaiting:
			logf("barber", "no customers in queue, sleeping")
			shouldExit := barberSleepLoop(mailbox, num_completed, avg_wait, avg_rating) // block until sleep loop is exited
			if shouldExit {
				return
			}
			requestNextCustomer()
			continue
		case MsgGetStats:
			msg.From <- Message{
				Kind: MsgBarberStatsReply,
				CutCountStat: num_completed,
				AvgWaitTime: avg_wait,
				AvgRatingStat: avg_rating,
			}
			continue
		case MsgShutdown:
			if msg.From != nil {
				msg.From <- Message{Kind: MsgShutdownAck}
			}
			logf("barber", "shutting down")
			return
		case MsgWakeUp:
			continue
		}
	}
}

func barberSleepLoop(mailbox chan Message, numCompleted int, avgDuration float64, avgRating float64) bool {
	for {
		msg := <-mailbox
		switch msg.Kind {
			case MsgWakeUp:
				logf("barber", "wakes up")
				return false
			case MsgGetStats:
				msg.From <- Message{
					Kind: MsgBarberStatsReply,
					CutCountStat: numCompleted,
					AvgWaitTime: avgDuration,
					AvgRatingStat: avgRating,
				}
			case MsgShutdown:
				if msg.From != nil {
					msg.From <- Message{Kind: MsgShutdownAck}
				}
				logf("barber", "shutting down")
				return true
		}
	}
}