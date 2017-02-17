package tart

import (
	"sync"
)

// Message can be anything.
type Message interface{}

// Behavior is a function executed by an Actor on receipt of a Message.
type Behavior func(*Context, Message)

// NonSerialBehavior is a function executed by a non-serial Actor on receipt of a Message.
type NonSerialBehavior func(*NonSerialContext, Message)

// Sponsor is a capability to create a new Actor with specified Behavior.
type Sponsor func(Behavior) Actor

// NonSerialSponsor is a capability to create a new non-serial Actor with specified NonSerialBehavior.
type NonSerialSponsor func(NonSerialBehavior) Actor

// Actor type is a capability to send a message to the Actor.
type Actor func(Message)

// Actor execution context.
type Context struct {
	// Become a non-Serial Actor with corresponding NonSerialBehavior.
	BecomeNonSerial func(NonSerialBehavior)
	// Actor behavior. Setting Behavior will change how the actor handles the next message it receives.
	Behavior Behavior
	// Relay function to the deliver function implementation called when sending message to this Actor.
	relay func(Message)
	// Capability to Sponsor (create) new Actors.
	Sponsor Sponsor
	// Capability to Sponsor (create) new non-Serial Actors.
	SponsorNonSerial NonSerialSponsor
	// Capability to send messages to Self.
	Self Actor
}

type NonSerialContext struct {
	// Actor non-serial behavior
	behavior NonSerialBehavior
	// Capability to Sponsor (create) new Actors.
	Sponsor Sponsor
	// Capability to Sponsor (create) new non-Serial Actors.
	SponsorNonSerial NonSerialSponsor
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
func Minimal(options *Options) (Sponsor, NonSerialSponsor) {
	var dispatch func(deliver)
	var fail func(interface{})
	var sponsor func(Behavior) Actor
	var sponsorNonSerial func(NonSerialBehavior) Actor
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
		var relay func(Message)
		mutex := &sync.Mutex{} // required for serial actors only
		relay = func(message Message) {
			dispatch(func() {
				mutex.Lock()
				defer func() {
					if p := recover(); p != nil {
						fail(p)
					}
					mutex.Unlock()
				}()
				context.Behavior(context, message)
			})
		}
		actor = func(message Message) {
			context.relay(message)
		}
		becomeNonSerial := func(nonSerialBehavior NonSerialBehavior) {
			context.relay = context.SponsorNonSerial(nonSerialBehavior)
		}
		context = &Context{BecomeNonSerial: becomeNonSerial, Behavior: behavior, relay: relay, Sponsor: sponsor, SponsorNonSerial: sponsorNonSerial, Self: actor}
		return actor
	}
	sponsorNonSerial = func(behavior NonSerialBehavior) Actor {
		var actor Actor
		var context *NonSerialContext
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
		context = &NonSerialContext{behavior: behavior, Sponsor: sponsor, SponsorNonSerial: sponsorNonSerial, Self: actor}
		return actor
	}
	return sponsor, sponsorNonSerial
}
