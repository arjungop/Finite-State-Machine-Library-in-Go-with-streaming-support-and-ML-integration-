// Package fsm implements a finite state machine framework for Go applications
package fsm

import (
	"fmt"  // Standard library for string formatting and printing
	"time" // Standard library for time operations and timestamps
)

// State represents a state in the finite state machine
// This is a string-based type that provides type safety for state names
type State string

// Event represents an event that can trigger state transitions
// Events are the catalysts that cause the FSM to move between states
type Event string

// TransitionCondition is a function that determines if a transition should occur
// It takes a Context and returns true if the transition is allowed
type TransitionCondition func(context Context) bool

// TransitionAction is a function executed during a state transition
// It receives the source state, destination state, triggering event, and context
// Returns an error if the action fails, which will abort the transition
type TransitionAction func(from, to State, event Event, context Context) error

// Context holds data that can be accessed during transitions
// This interface provides a key-value store for sharing data between transitions
type Context interface {
	Get(key string) interface{}        // Retrieves a value by key from the context
	Set(key string, value interface{}) // Stores a key-value pair in the context
	GetAll() map[string]interface{}    // Returns all key-value pairs as a map
}

// Transition defines a state transition rule in the finite state machine
// This struct encapsulates all information needed for a single transition
type Transition struct {
	From      State               // The source state that the transition starts from
	Event     Event               // The event that triggers this transition
	To        State               // The destination state that the transition leads to
	Condition TransitionCondition // Optional guard condition that must be true for transition
	Action    TransitionAction    // Optional action to execute when transition occurs
}

// String returns a string representation of the transition for debugging and logging
// Format: "source_state --event--> destination_state"
func (t Transition) String() string {
	return fmt.Sprintf("%s --%s--> %s", t.From, t.Event, t.To) // Creates human-readable transition description
}

// TransitionResult contains the result of a transition attempt
// This struct provides comprehensive information about what happened during a transition
type TransitionResult struct {
	Success     bool      // Indicates whether the transition completed successfully
	FromState   State     // The state the machine was in before the transition
	ToState     State     // The state the machine is in after the transition
	Event       Event     // The event that triggered this transition attempt
	Error       error     // Any error that occurred during the transition (nil if successful)
	Timestamp   time.Time // When the transition occurred for auditing and debugging
	ExecutionID string    // Unique identifier for this transition execution
}

// Hook represents a callback function for FSM events
// Hooks allow external code to respond to state machine events and transitions
type Hook func(result TransitionResult, context Context)

// HookType defines when a hook should be executed
// This enumeration specifies the timing of hook execution relative to transitions
type HookType int

// Hook execution timing constants
const (
	BeforeTransition  HookType = iota // Hook executes before a transition begins
	AfterTransition                   // Hook executes after a transition completes successfully
	OnTransitionError                 // Hook executes when a transition fails with an error
	OnStateEnter                      // Hook executes when entering any state
	OnStateExit                       // Hook executes when exiting any state
)

// FSMError represents errors specific to finite state machine operations
// This custom error type provides detailed context about FSM-related failures
type FSMError struct {
	Type    string // Classification of the error (e.g., "InvalidTransition")
	Message string // Human-readable description of what went wrong
	State   State  // The state involved in the error (if applicable)
	Event   Event  // The event involved in the error (if applicable)
}

// Error implements the error interface for FSMError
// Returns a formatted string containing all error details for debugging
func (e FSMError) Error() string {
	return fmt.Sprintf("FSM Error [%s]: %s (State: %s, Event: %s)", // Formats error with type, message, and context
		e.Type, e.Message, e.State, e.Event)
}

// NewInvalidTransitionError creates an error for invalid transitions
// Used when an event is triggered from a state that has no valid transition for that event
func NewInvalidTransitionError(from State, event Event) FSMError {
	return FSMError{
		Type:    "InvalidTransition",                                                             // Error classification
		Message: fmt.Sprintf("No valid transition from state '%s' with event '%s'", from, event), // Descriptive message
		State:   from,                                                                            // Source state
		Event:   event,                                                                           // Triggering event
	}
}

// NewStateNotFoundError creates an error when a state doesn't exist
// Used during FSM validation when referencing undefined states
// NewStateNotFoundError creates an error when a state doesn't exist
// Used during FSM validation when referencing undefined states
func NewStateNotFoundError(state State) FSMError {
	return FSMError{
		Type:    "StateNotFound",                                             // Error classification for missing states
		Message: fmt.Sprintf("State '%s' is not defined in this FSM", state), // Human-readable error message
		State:   state,                                                       // The undefined state that caused the error
	}
}

// Machine interface defines the core FSM operations
// This interface provides the complete API for interacting with finite state machines
type Machine interface {
	// State operations - methods for managing the current state of the machine
	CurrentState() State           // Returns the current state the machine is in
	SetState(state State) error    // Directly sets the machine to a specific state (bypassing transitions)
	IsValidState(state State) bool // Checks if a given state is defined in this FSM

	// Event operations - methods for triggering and validating events
	SendEvent(event Event) (*TransitionResult, error) // Triggers an event and attempts a state transition
	CanTransition(event Event) bool                   // Checks if an event can trigger a transition from current state
	GetValidEvents() []Event                          // Returns all events that are valid from the current state

	// Transition operations - methods for managing the transition rules
	AddTransition(transition Transition) error      // Adds a new transition rule to the FSM
	RemoveTransition(from State, event Event) error // Removes a specific transition rule
	GetTransitions() []Transition                   // Returns all transition rules defined in the FSM

	// Hook operations - methods for managing callback functions
	AddHook(hookType HookType, hook Hook) // Registers a callback function for specific FSM events
	RemoveHook(hookType HookType)         // Unregisters callbacks for a specific hook type

	// Context operations - methods for managing shared data
	GetContext() Context        // Returns the current context (shared data store)
	SetContext(context Context) // Replaces the current context with a new one

	// Machine lifecycle - methods for controlling the FSM's operational state
	Start(initialState State) error // Initializes the FSM and sets it to the starting state
	Stop() error                    // Stops the FSM and prevents further state transitions
	Reset() error                   // Resets the FSM to its initial configuration
	IsRunning() bool                // Returns true if the FSM is currently active and can process events

	// Validation - method for ensuring FSM integrity
	Validate() error // Checks if the FSM configuration is valid and consistent
}

// Builder interface for fluent FSM construction
// This interface provides a chainable API for building finite state machines
type Builder interface {
	AddState(state State) Builder                                                                                        // Adds a single state to the FSM being built
	AddStates(states ...State) Builder                                                                                   // Adds multiple states in one call using variadic parameters
	AddEvent(event Event) Builder                                                                                        // Adds a single event that can trigger transitions
	AddEvents(events ...Event) Builder                                                                                   // Adds multiple events in one call using variadic parameters
	AddTransition(from State, event Event, to State) Builder                                                             // Adds a basic transition without conditions or actions
	AddTransitionWithCondition(from State, event Event, to State, condition TransitionCondition) Builder                 // Adds a transition with a guard condition
	AddTransitionWithAction(from State, event Event, to State, action TransitionAction) Builder                          // Adds a transition with an action to execute
	AddTransitionFull(from State, event Event, to State, condition TransitionCondition, action TransitionAction) Builder // Adds a transition with both condition and action
	SetInitialState(state State) Builder                                                                                 // Specifies which state the FSM should start in
	Build() (Machine, error)                                                                                             // Constructs the final FSM and returns it (or an error if invalid)
}

// ContextImpl provides a basic implementation of Context
// This struct implements the Context interface using a simple map for data storage
type ContextImpl struct {
	data map[string]interface{} // Internal map to store key-value pairs
}

// NewContext creates a new context instance
// Factory function that returns a ready-to-use Context implementation
func NewContext() Context {
	return &ContextImpl{
		data: make(map[string]interface{}), // Initialize empty map for storing context data
	}
}

// Get retrieves a value from the context
// Returns the value associated with the given key, or nil if key doesn't exist
func (c *ContextImpl) Get(key string) interface{} {
	return c.data[key] // Direct map lookup using the provided key
}

// Set stores a value in the context
// Associates the given value with the provided key in the context
func (c *ContextImpl) Set(key string, value interface{}) {
	c.data[key] = value // Store the key-value pair in the internal map
}

// GetAll returns all context data
// Creates and returns a copy of all stored key-value pairs
func (c *ContextImpl) GetAll() map[string]interface{} {
	result := make(map[string]interface{}) // Create new map to avoid exposing internal state
	for k, v := range c.data {             // Iterate through all stored data
		result[k] = v // Copy each key-value pair to the result map
	}
	return result // Return the copy of all context data
}
