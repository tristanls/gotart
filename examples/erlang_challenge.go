package main

import (
	"fmt"
	tart "github.com/tristanls/gotart"
	"sync"
	"time"
)

type InitialMessage struct {
	Int   int
	Actor tart.Actor
}

func main() {
	var waitGroup sync.WaitGroup
	M := 100000
	N := 10

	var constructionStartTime time.Time
	var constructionEndTime time.Time
	loopCompletionTimes := make([]time.Time, 0)

	reportProcessTimes := func() {
		fmt.Printf("\ndone\n")
		constructionTime := constructionEndTime.UnixNano() - constructionStartTime.UnixNano()
		fmt.Printf("all times in NANOSECONDS\n")
		fmt.Printf("construction time: %v\n", constructionTime)
		fmt.Printf("%v\n", constructionTime)
		fmt.Printf("loop times:\n")
		loopIntervals := make([]int64, 0)
		prevTime := constructionEndTime
		var counter int64 = 0
		for _, time := range loopCompletionTimes {
			counter += 1
			interval := time.UnixNano() - prevTime.UnixNano()
			fmt.Printf("%v\n", interval)
			loopIntervals = append(loopIntervals, interval)
			prevTime = time
		}
		fmt.Printf("loop average:")
		var averageInterval int64 = 0
		for _, interval := range loopIntervals {
			averageInterval += interval
		}
		averageInterval = averageInterval / counter
		fmt.Printf("%v\n", averageInterval)
		waitGroup.Done()
	}

	sponsor, _ := tart.Minimal(nil)

	ringLink := func(next tart.Actor) tart.NonSerialBehavior {
		return func(context *tart.NonSerialContext, message tart.Message) {
			next(message)
		}
	}

	ringLast := func(endTime time.Time, first tart.Actor) tart.Behavior {
		return func(context *tart.Context, message tart.Message) {
			loopCompletionTimes = append(loopCompletionTimes, time.Now())
			n := message.(int)
			n -= 1
			if n > 0 {
				fmt.Printf(".")
				first(n)
			} else {
				context.BecomeNonSerial(func(context *tart.NonSerialContext, message tart.Message) {})
				fmt.Printf(".")
				constructionEndTime = endTime
				reportProcessTimes()
			}
		}
	}

	var ringBuilder func(int) tart.Behavior
	ringBuilder = func(m int) tart.Behavior {
		return func(context *tart.Context, message tart.Message) {
			m -= 1
			if m > 0 {
				next := context.Sponsor(ringBuilder(m))
				next(message)
				context.BecomeNonSerial(ringLink(next))
			} else {
				msg := message.(InitialMessage)
				now := time.Now()
				fmt.Printf("sending %v messages\n", msg.Int)
				first := msg.Actor
				first(msg.Int)
				context.Behavior = ringLast(now, first)
			}
		}
	}

	waitGroup.Add(1)
	fmt.Printf("starting %v actor ring\n", M)
	constructionStartTime = time.Now()
	ring := sponsor(ringBuilder(M))
	ring(InitialMessage{Int: N, Actor: ring})
	waitGroup.Wait()
}
