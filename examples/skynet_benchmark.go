package main

import (
	"fmt"
	tart "github.com/tristanls/gotart"
	"sync"
	"time"
)

type Message struct {
	Num int
	Size int
	Div int
}

func node(parent tart.Actor, numDescendants int) tart.Behavior {
	accumulator, count := 0, 0
	accumulatorBeh := func(c *tart.Context, m tart.Message) {
		// fmt.Printf("accumulatorBeh: %v\n", m)
		count += 1
		accumulator += m.(int)
		if count == numDescendants {
			parent(accumulator)
		}
	}
	initialBeh := func(c *tart.Context, m tart.Message) {
		msg := m.(Message)
		num, size, div := msg.Num, msg.Size, msg.Div
		// fmt.Printf("initialBeh: %v, %v, %v\n", num, size, div)
		if size == 1 {
			parent(num)
			return
		}
		for i := 0; i < div; i++ {
			subNum := num + i*(size/div)
			descendant := c.Sponsor(node(c.Self, div))
			descendant(Message{Num: subNum, Size: size / div, Div: div})
		}
		c.Become(accumulatorBeh)
	}
	return initialBeh
}

func main() {
	var wg sync.WaitGroup
	sponsor, _ := tart.Minimal(nil)

	size := 1000000
	div := 10

	start := time.Now()
	wg.Add(1)

	reporter := sponsor(func(c *tart.Context, m tart.Message) {
		result := m.(int)
		took := time.Since(start)
		fmt.Printf("Result: %d in %d ms.\n", result, took.Nanoseconds()/1e6)
		wg.Done()
	})

	root := sponsor(node(reporter, div))
	root(Message{Num: 0, Size: size, Div: div})

	wg.Wait()
}
