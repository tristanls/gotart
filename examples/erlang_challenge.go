package main

import (
	"fmt"
	tart "github.com/tristanls/gotart"
	"sync"
	"time"
)

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

	sponsor := tart.Minimal(nil)

	ringLink := func(next tart.Actor) tart.Behavior {
		return func(context *tart.Context, message tart.Message) {
			next(message)
		}
	}

	ringLast := func(endTime time.Time, first tart.Actor) tart.Behavior {
		return func(context *tart.Context, message tart.Message) {
			loopCompletionTimes = append(loopCompletionTimes, time.Now())
			n := message[0].(int)
			n -= 1
			if n > 0 {
				fmt.Printf(".")
				first([]interface{}{n})
			} else {
				context.Behavior = func(context *tart.Context, message tart.Message) {}
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
				context.Behavior = ringLink(next)
			} else {
				now := time.Now()
				fmt.Printf("sending %v messages\n", message[0])
				first := message[1].(tart.Actor)
				first([]interface{}{message[0]})
				context.Behavior = ringLast(now, first)
			}
		}
	}

	waitGroup.Add(1)
	fmt.Printf("starting %v actor ring\n", M)
	constructionStartTime = time.Now()
	ring := sponsor(ringBuilder(M))
	ring([]interface{}{N, ring})
	waitGroup.Wait()
}
