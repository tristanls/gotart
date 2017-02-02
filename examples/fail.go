package main

import (
    "errors"
    "fmt"
    "sync"
    tart "github.com/tristanls/gotart"
)

func main() {
    var waitGroup sync.WaitGroup
    failHandler := func(err error) {
        fmt.Printf("%v boom!\n", err)
        waitGroup.Done()
    }
    sponsor := tart.Minimal(&tart.Options{Fail: failHandler})

    failing := sponsor(func(context *tart.Context, message tart.Message) error {
        return errors.New("boom!")
    })

    waitGroup.Add(1)

    failing([]interface{}{"go"})

    waitGroup.Wait()
}
