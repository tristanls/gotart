package main

import (
	"fmt"
	tart "github.com/tristanls/gotart"
	"sync"
)

func main() {
	var waitGroup sync.WaitGroup
	sponsor := tart.Minimal(nil)

	sinkBeh := func(context *tart.Context, message tart.Message) {
		fmt.Printf("%v sinkBehDone\n", message)
		waitGroup.Done()
	}

	oneShot := func(destination tart.Actor) tart.Behavior {
		return func(context *tart.Context, message tart.Message) {
			destination(message)
			context.Behavior = sinkBeh
		}
	}

	destination := sponsor(func(context *tart.Context, message tart.Message) {
		fmt.Printf("%v destinationDone\n", message)
		waitGroup.Done()
	})

	oneShotActor := sponsor(oneShot(destination))

	waitGroup.Add(2)

	oneShotActor([]interface{}{"first"})
	oneShotActor([]interface{}{"second"})

	waitGroup.Wait()
}
