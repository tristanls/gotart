package main

import (
    "fmt"
    "sync"
    tart "github.com/tristanls/gotart"
)

func main() {
    var waitGroup sync.WaitGroup
    sponsor := tart.Minimal(nil)

    sinkBeh := func(context *tart.Context, message tart.Message) error {
        fmt.Printf("%v sinkBehDone\n", message)
        waitGroup.Done()
        return nil
    }

    oneShot := func(destination tart.Actor) tart.Behavior {
        return func(context *tart.Context, message tart.Message) error {
            destination(message)
            context.Behavior = sinkBeh
            return nil
        }
    }

    destination := sponsor(func(context *tart.Context, message tart.Message) error {
        fmt.Printf("%v destinationDone\n", message)
        waitGroup.Done()
        return nil
    })

    oneShotActor := sponsor(oneShot(destination))

    waitGroup.Add(2)

    oneShotActor([]interface{}{"first"})
    oneShotActor([]interface{}{"second"})

    waitGroup.Wait()
}
