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
	// Mutex for updating context. Required for Serial Actors only.
	mutex sync.Mutex
	// Flag used for correct transition into NonSerialContext.
	isNonSerial bool
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
		var context *Context
		relay := func(message Message) {
			dispatch(func() {
				// When messages are sent to serial Actor, they wait for the context.mutext.Lock() here.
				// While waiting, the Actor may have become non-Serial and started using the NonSerialContext
				// instead. Hence the check of the context.isNonSerial flag below.
				context.mutex.Lock()
				defer func() {
					if p := recover(); p != nil {
						fail(p)
					}
					context.mutex.Unlock()
				}()
				if context.isNonSerial {
					// We've become a non-Serial Actor, but we're about to execute in Serial Context. Relay the
					// message to NonSerialContext.
					context.relay(message)
				} else {
					context.Behavior(context, message)
				}
			})
		}
		actor := func(message Message) {
			context.relay(message)
		}
		becomeNonSerial := func(nonSerialBehavior NonSerialBehavior) {
			context.relay = context.SponsorNonSerial(nonSerialBehavior)
			context.isNonSerial = true
		}
		context = &Context{BecomeNonSerial: becomeNonSerial, Behavior: behavior, relay: relay, Sponsor: sponsor, SponsorNonSerial: sponsorNonSerial, Self: actor}
		return actor
	}
	sponsorNonSerial = func(behavior NonSerialBehavior) Actor {
		var context *NonSerialContext
		actor := func(message Message) {
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
