package fsm

import (
	"fmt"
	"time"
)

// State represents a state in the finite state machine
type State string

// Event represents an event that can trigger state transitions
type Event string

// TransitionCondition is a function that determines if a transition should occur
type TransitionCondition func(context Context) bool

// TransitionAction is a function executed during a state transition
type TransitionAction func(from, to State, event Event, context Context) error

// Context holds data that can be accessed during transitions
type Context interface {
	Get(key string) interface{}
	Set(key string, value interface{})
	GetAll() map[string]interface{}
}

// Transition defines a state transition rule
type Transition struct {
	From      State
	Event     Event
	To        State
	Condition TransitionCondition // Optional guard condition
	Action    TransitionAction    // Optional action to execute
}

// String returns a string representation of the transition
func (t Transition) String() string {
	return fmt.Sprintf("%s --%s--> %s", t.From, t.Event, t.To)
}

// TransitionResult contains the result of a transition attempt
type TransitionResult struct {
	Success     bool
	FromState   State
	ToState     State
	Event       Event
	Error       error
	Timestamp   time.Time
	ExecutionID string
}

// Hook represents a callback function for FSM events
type Hook func(result TransitionResult, context Context)

// HookType defines when a hook should be executed
type HookType int

const (
	BeforeTransition HookType = iota
	AfterTransition
	OnTransitionError
	OnStateEnter
	OnStateExit
)

// FSMError represents errors specific to finite state machine operations
type FSMError struct {
	Type    string
	Message string
	State   State
	Event   Event
}

func (e FSMError) Error() string {
	return fmt.Sprintf("FSM Error [%s]: %s (State: %s, Event: %s)",
		e.Type, e.Message, e.State, e.Event)
}

// NewInvalidTransitionError creates an error for invalid transitions
func NewInvalidTransitionError(from State, event Event) FSMError {
	return FSMError{
		Type:    "InvalidTransition",
		Message: fmt.Sprintf("No valid transition from state '%s' with event '%s'", from, event),
		State:   from,
		Event:   event,
	}
}

// NewStateNotFoundError creates an error when a state doesn't exist
func NewStateNotFoundError(state State) FSMError {
	return FSMError{
		Type:    "StateNotFound",
		Message: fmt.Sprintf("State '%s' is not defined in this FSM", state),
		State:   state,
	}
}

// Machine interface defines the core FSM operations
type Machine interface {
	// State operations
	CurrentState() State
	SetState(state State) error
	IsValidState(state State) bool

	// Event operations
	SendEvent(event Event) (*TransitionResult, error)
	CanTransition(event Event) bool
	GetValidEvents() []Event

	// Transition operations
	AddTransition(transition Transition) error
	RemoveTransition(from State, event Event) error
	GetTransitions() []Transition

	// Hook operations
	AddHook(hookType HookType, hook Hook)
	RemoveHook(hookType HookType)

	// Context operations
	GetContext() Context
	SetContext(context Context)

	// Machine lifecycle
	Start(initialState State) error
	Stop() error
	Reset() error
	IsRunning() bool

	// Validation
	Validate() error
}

// Builder interface for fluent FSM construction
type Builder interface {
	AddState(state State) Builder
	AddStates(states ...State) Builder
	AddEvent(event Event) Builder
	AddEvents(events ...Event) Builder
	AddTransition(from State, event Event, to State) Builder
	AddTransitionWithCondition(from State, event Event, to State, condition TransitionCondition) Builder
	AddTransitionWithAction(from State, event Event, to State, action TransitionAction) Builder
	AddTransitionFull(from State, event Event, to State, condition TransitionCondition, action TransitionAction) Builder
	SetInitialState(state State) Builder
	Build() (Machine, error)
}

// ContextImpl provides a basic implementation of Context
type ContextImpl struct {
	data map[string]interface{}
}

// NewContext creates a new context instance
func NewContext() Context {
	return &ContextImpl{
		data: make(map[string]interface{}),
	}
}

// Get retrieves a value from the context
func (c *ContextImpl) Get(key string) interface{} {
	return c.data[key]
}

// Set stores a value in the context
func (c *ContextImpl) Set(key string, value interface{}) {
	c.data[key] = value
}

// GetAll returns all context data
func (c *ContextImpl) GetAll() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range c.data {
		result[k] = v
	}
	return result
}
