package main

import (
	"fmt"
	tart "github.com/tristanls/gotart"
	"sync"
	"time"
)

func node(parent tart.Actor, numDescendants int) tart.Behavior {
	accumulator, count := 0, 0
	accumulatorBeh := func(c *tart.Context, m tart.Message) error {
		// fmt.Printf("accumulatorBeh: %v\n", m[0].(int))
		count += 1
		accumulator += m[0].(int)
		if count == numDescendants {
			parent([]interface{}{accumulator})
		}
		return nil
	}
	initialBeh := func(c *tart.Context, m tart.Message) error {
		num, size, div := m[0].(int), m[1].(int), m[2].(int)
		// fmt.Printf("initialBeh: %v, %v, %v\n", num, size, div)
		if size == 1 {
			parent([]interface{}{num})
			return nil
		}
		for i := 0; i < div; i++ {
			subNum := num + i*(size/div)
			descendant := c.Sponsor(node(c.Self, div))
			descendant([]interface{}{subNum, size / div, div})
		}
		c.Behavior = accumulatorBeh
		return nil
	}
	return initialBeh
}

func main() {
	var wg sync.WaitGroup
	sponsor := tart.Minimal(nil)

	size := 1000000
	div := 10

	start := time.Now()
	wg.Add(1)

	reporter := sponsor(func(c *tart.Context, m tart.Message) error {
		result := m[0]
		took := time.Since(start)
		fmt.Printf("Result: %d in %d ms.\n", result, took.Nanoseconds()/1e6)
		wg.Done()
		return nil
	})

	root := sponsor(node(reporter, div))
	root([]interface{}{0, size, div})

	wg.Wait()
}
