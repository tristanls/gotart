package main

import (
	"fmt"
	"sync"
	"time"

	tart "github.com/tristanls/gotart"
)

func main() {
	var waitGroup sync.WaitGroup
	sponsor, _ := tart.Minimal(nil)

	nonSerialTimeBeh := func(context *tart.NonSerialContext, message tart.Message) {
		fmt.Printf("non-serial - %s timeBeh\n", time.Now().UTC().Format(time.RFC3339Nano))
	}

	serialTime := func(count int) tart.Behavior {
		return func(context *tart.Context, message tart.Message) {
			count -= 1
			fmt.Printf("%v - %s timeBeh\n", count, time.Now().UTC().Format(time.RFC3339Nano))
			// send two messages for each one... will be "throttled" by serial nature of this actor
			// until counter ends up at zero
			context.Self(nil)
			context.Self(nil)
			if count == 0 {
				context.BecomeNonSerial(nonSerialTimeBeh)
			}
		}
	}

	timeActor := sponsor(serialTime(10))

	waitGroup.Add(1)

	timeActor(nil)

	waitGroup.Wait()
}
