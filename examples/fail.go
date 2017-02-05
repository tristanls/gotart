package main

import (
	"fmt"
	tart "github.com/tristanls/gotart"
	"sync"
)

func main() {
	var waitGroup sync.WaitGroup
	failHandler := func(r interface{}) {
		fmt.Printf("%v boom!\n", r)
		waitGroup.Done()
	}
	sponsor := tart.Minimal(&tart.Options{Fail: failHandler})

	failing := sponsor(func(context *tart.Context, message tart.Message) {
		panic("boom!")
	})

	waitGroup.Add(1)

	failing([]interface{}{"go"})

	waitGroup.Wait()
}
