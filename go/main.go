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
const satThresh = 3000 // ms per star lost
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
	Rating int // rating from customer
	WaitTime int64 // customer wait time (ms)
	// Waiting Room stats
	QueueLengthStat int 
	TurnawayCountStat int
	// Barber stats 
	CutCountStat int
	AvgCutDurStat float64
	AvgRatingStat float64

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
		time.Sleep(time.Duration(rand.Intn(custArrivMax-custArrivMin)+custArrivMin)*time.Millisecond)
	}
	time.Sleep(time.Duration(2*cutDurMax)*time.Millisecond)
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
	
	waitRoomChannel <- Message{
		Kind: MsgShutdown,
	}
	barberChannel <- Message{
		Kind: MsgShutdown,
	}

	total_cust := barberStatsReply.CutCountStat + waitRoomStatsReply.TurnawayCountStat
	cust_served := barberStatsReply.CutCountStat
	cust_rejected := waitRoomStatsReply.TurnawayCountStat
	avg_duration := barberStatsReply.AvgCutDurStat
	avg_rating := barberStatsReply.AvgRatingStat

	fmt.Println("== CLOSING REPORT ==")
	fmt.Printf("Total customers arrived: %d\n", total_cust)
	fmt.Printf("Total customers served: %d\n", cust_served)
	fmt.Printf("Total customers arrived: %d\n", cust_rejected)
	fmt.Printf("Total customers arrived: %d\n", avg_duration)
	fmt.Printf("Total customers arrived: %d\n", avg_rating)
	fmt.Println("============")
}


func waitingRoom(mailbox chan Message) {
	// todo: clarify when to send wakeup signal; ensure is sent only once per cycle
	turnaway_count := 0
	cust_queue := make([]int, 0, waitCapacity)
	barberIsSleeping := false
	var barberChannel chan Message
	for {
		msg := <- mailbox
		switch msg.Kind {
		case MsgNextCustomer:
			barberChannel = msg.From
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
		
		case MsgGetStats:
			msg.From <- Message{
				Kind: MsgWRStatsReply,
				QueueLengthStat: len(cust_queue),
				TurnawayCountStat: turnaway_count,
			}
		
		case MsgShutdown:
			return
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
			wait := cutStartTime - arriveTime
			jitter := rand.Intn(3) - 1 // {-1, 0, +1}
			score := clamp(1, 5, int(5-(wait/satThresh))+jitter)
			
			rateReqMsg := <- mailbox
			if rateReqMsg.Kind != MsgRateRequest {
				fmt.Printf("Error: customer %d received unexpected message instead of MsgRateRequest", id)
				return
			}
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
	var avg_duration float64 = 0
	var avg_rating float64 = 0

	waitRoomChannel <- Message{
		Kind: MsgNextCustomer,
		From: mailbox,
	}
	
	for {
		msg := <- mailbox
		switch msg.Kind {
		case MsgCustomerReady:
			msg.From <- Message{
				Kind: MsgStartHaircut,
			}
			cutLength := time.Duration(rand.Intn((cutDurMax-cutDurMin)) + cutDurMin) * time.Millisecond
			time.Sleep(cutLength)
			ratingReplyChannel := make(chan Message)
			msg.From <- Message{
				Kind: MsgRateRequest,
				From: ratingReplyChannel,
			}
			ratingResponse:= <- ratingReplyChannel
			num_completed += 1
			avg_rating = (avg_rating + float64(ratingResponse.Rating)) / float64(num_completed)
			avg_duration = (avg_duration + float64(ratingResponse.WaitTime)) / float64(num_completed)

		case MsgNoneWaiting:
			shouldExit := barberSleepLoop(mailbox) // block until sleep loop is exited
			if shouldExit {
				return
			}
			continue
		case MsgGetStats:
			msg.From <- Message{
				Kind: MsgBarberStatsReply,
				CutCountStat: num_completed,
				AvgCutDurStat: avg_duration,
				AvgRatingStat: avg_rating,
			}
			continue
		case MsgWakeUp:
			continue
		}
	}
}

func barberSleepLoop(mailbox chan Message) bool {
	for {
		msg := <-mailbox
		switch msg.Kind {
			case MsgWakeUp:
				return false // return
			case MsgShutdown:
				return true // terminate barber
		}
	}
}