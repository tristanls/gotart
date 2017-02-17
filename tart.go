package tart

import (
	"sync"
)

// Message can be anything.
type Message interface{}

// Behavior is a function executed by an Actor on receipt of a Message.
type Behavior func(*Context, Message)

// Sponsor is a capability to create a new Actor with specified Behavior.
type Sponsor func(Behavior) Actor

// Actor type is a capability to send a message to the Actor.
type Actor func(Message)

// Actor execution context.
type Context struct {
	// Become changes how the actor will handle the next message it receives.
	Become func(Behavior)
	// Actor behavior.
	behavior Behavior
	// Capability to Sponsor (create) new Actors.
	Sponsor Sponsor
	// Capability to send messages to Self.
	Self Actor
}

type deliver func()

// Options for Minimal implementation.
type Options struct {
	// Non-default message dispatch function to use.
	Dispatch func(deliver)
	// Non-default actor behavior panic recovery to use.
	Fail func(interface{})
}

// Default actor behavior panic recovery.
func Fail(_ interface{}) {}

// Default message dispatch function.
func Dispatch(deliver deliver) {
	go deliver()
}

// Creates a Sponsor capability to create new actors with.
func Minimal(options *Options) (Sponsor, Sponsor) {
	var dispatch func(deliver)
	var fail func(interface{})
	var sponsor func(Behavior) Actor
	var sponsorNonSerial func(Behavior) Actor
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
					if p := recover(); p != nil {
						fail(p)
					}
					mutex.Unlock()
				}()
				context.behavior(context, message)
			})
		}
		become := func(behavior Behavior) {
			context.behavior = behavior
		}
		context = &Context{Become: become, behavior: behavior, Sponsor: sponsor, Self: actor}
		return actor
	}
	sponsorNonSerial = func(behavior Behavior) Actor {
		var actor Actor
		var context *Context
		actor = func(message Message) {
			dispatch(func() {
				defer func() {
					if p := recover(); p != nil {
						fail(p)
					}
				}()
				context.behavior(context, message)
			})
		}
		context = &Context{behavior: behavior, Sponsor: sponsor, Self: actor}
		return actor
	}
	return sponsor, sponsorNonSerial
}
