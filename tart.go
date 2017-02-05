package tart

import (
	"sync"
)

// Message is a slice of data
type Message []interface{}

type Behavior func(*Context, Message) error

type Sponsor func(Behavior) Actor

type Actor func(Message)

type Context struct {
	Behavior Behavior
	Sponsor  Sponsor
	Self     Actor
}

type deliver func()

type Options struct {
	Dispatch func(deliver)
	Fail     func(error)
}

func Dispatch(deliver deliver) {
	go deliver()
}

func Minimal(options *Options) Sponsor {
	var dispatch func(deliver)
	var fail func(error)
	var sponsor func(Behavior) Actor
	if options != nil && options.Fail != nil {
		fail = options.Fail
	} else {
		fail = func(_ error) {}
	}
	if options != nil && options.Dispatch != nil {
		dispatch = options.Dispatch
	} else {
		dispatch = Dispatch
	}
	sponsor = func(behavior Behavior) Actor {
		var actor Actor
		var context *Context
		mutex := &sync.Mutex{} // required for serial actors only
		actor = func(message Message) {
			dispatch(func() {
				mutex.Lock()
				err := context.Behavior(context, message)
				mutex.Unlock()
				if err != nil {
					fail(err)
				}
			})
		}
		context = &Context{Behavior: behavior, Sponsor: sponsor, Self: actor}
		return actor
	}
	return sponsor
}
