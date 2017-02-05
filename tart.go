package tart

import (
	"sync"
)

// Message is a slice of data
type Message []interface{}

type Behavior func(*Context, Message)

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
	Fail     func(interface{})
}

func Fail(_ interface{}) {}

func Dispatch(deliver deliver) {
	go deliver()
}

func Minimal(options *Options) Sponsor {
	var dispatch func(deliver)
	var fail func(interface{})
	var sponsor func(Behavior) Actor
	if options != nil && options.Fail != nil {
		fail = options.Fail
	} else {
		fail = Fail
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
				defer func() {
					mutex.Unlock()
					if p := recover(); p != nil {
						fail(p)
					}
				}()
				context.Behavior(context, message)
			})
		}
		context = &Context{Behavior: behavior, Sponsor: sponsor, Self: actor}
		return actor
	}
	return sponsor
}
