// Package fsm provides finite state machine implementation
package fsm

import (
	"crypto/rand" // Used for generating cryptographically secure random bytes
	"fmt"         // Standard library for string formatting and printing
	"sync"        // Provides synchronization primitives for thread safety
	"time"        // Standard library for time operations and timestamps
)

// StateMachine is the core implementation of the Machine interface
// This struct contains all the data and logic needed for a functional FSM
type StateMachine struct {
	mu           sync.RWMutex          // Read-write mutex for thread-safe access to FSM state
	currentState State                 // The state the machine is currently in
	states       map[State]bool        // Set of all valid states (map used as set with bool values)
	events       map[Event]bool        // Set of all valid events that can trigger transitions
	transitions  map[string]Transition // Map of transition rules, keyed by "from_state:event"
	hooks        map[HookType][]Hook   // Map of hook functions organized by when they should execute
	context      Context               // Shared data store accessible during transitions
	running      bool                  // Flag indicating whether the FSM is currently active
	initialState State                 // The state this FSM should start in when initialized
}

// NewStateMachine creates a new finite state machine
// Factory function that returns a properly initialized StateMachine instance
func NewStateMachine() *StateMachine {
	return &StateMachine{
		states:      make(map[State]bool),        // Initialize empty set of states
		events:      make(map[Event]bool),        // Initialize empty set of events
		transitions: make(map[string]Transition), // Initialize empty map of transitions
		hooks:       make(map[HookType][]Hook),   // Initialize empty map of hook collections
		context:     NewContext(),                // Create new context instance for data sharing
		running:     false,                       // FSM starts in stopped state
	}
}

// generateExecutionID creates a unique identifier for transition execution
// Uses cryptographic random bytes to ensure uniqueness across all transitions
func generateExecutionID() string {
	bytes := make([]byte, 8)        // Create 8-byte array for random data
	rand.Read(bytes)                // Fill array with cryptographically secure random bytes
	return fmt.Sprintf("%x", bytes) // Convert bytes to hexadecimal string representation
}

// transitionKey creates a key for the transitions map
// Combines from state and event into a unique string identifier
func transitionKey(from State, event Event) string {
	return fmt.Sprintf("%s:%s", from, event) // Format: "source_state:event_name"
}

// CurrentState returns the current state of the machine
// Thread-safe method that provides read-only access to the current state
func (sm *StateMachine) CurrentState() State {
	sm.mu.RLock()         // Acquire read lock to safely access current state
	defer sm.mu.RUnlock() // Ensure lock is released when function exits
	return sm.currentState
}

// SetState manually sets the current state (used for initialization)
func (sm *StateMachine) SetState(state State) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.states[state] {
		return NewStateNotFoundError(state)
	}

	oldState := sm.currentState
	sm.currentState = state

	// Execute state exit hooks for old state
	if oldState != "" {
		sm.executeHooks(OnStateExit, TransitionResult{
			Success:     true,
			FromState:   oldState,
			ToState:     state,
			Timestamp:   time.Now(),
			ExecutionID: generateExecutionID(),
		})
	}

	// Execute state enter hooks for new state
	sm.executeHooks(OnStateEnter, TransitionResult{
		Success:     true,
		FromState:   oldState,
		ToState:     state,
		Timestamp:   time.Now(),
		ExecutionID: generateExecutionID(),
	})

	return nil
}

// IsValidState checks if a state is defined in the machine
func (sm *StateMachine) IsValidState(state State) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.states[state]
}

// SendEvent triggers an event and potentially causes a state transition
func (sm *StateMachine) SendEvent(event Event) (*TransitionResult, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.running {
		return nil, FSMError{
			Type:    "MachineNotRunning",
			Message: "Cannot send event to a stopped machine",
			Event:   event,
		}
	}

	if !sm.events[event] {
		return nil, FSMError{
			Type:    "EventNotFound",
			Message: fmt.Sprintf("Event '%s' is not defined in this FSM", event),
			Event:   event,
		}
	}

	key := transitionKey(sm.currentState, event)
	transition, exists := sm.transitions[key]

	if !exists {
		err := NewInvalidTransitionError(sm.currentState, event)
		result := &TransitionResult{
			Success:     false,
			FromState:   sm.currentState,
			ToState:     sm.currentState,
			Event:       event,
			Error:       err,
			Timestamp:   time.Now(),
			ExecutionID: generateExecutionID(),
		}

		sm.executeHooks(OnTransitionError, *result)
		return result, err
	}

	// Check guard condition if present
	if transition.Condition != nil && !transition.Condition(sm.context) {
		err := FSMError{
			Type:    "ConditionNotMet",
			Message: fmt.Sprintf("Transition condition not met for %s", transition),
			State:   sm.currentState,
			Event:   event,
		}

		result := &TransitionResult{
			Success:     false,
			FromState:   sm.currentState,
			ToState:     sm.currentState,
			Event:       event,
			Error:       err,
			Timestamp:   time.Now(),
			ExecutionID: generateExecutionID(),
		}

		sm.executeHooks(OnTransitionError, *result)
		return result, err
	}

	result := &TransitionResult{
		Success:     true,
		FromState:   sm.currentState,
		ToState:     transition.To,
		Event:       event,
		Timestamp:   time.Now(),
		ExecutionID: generateExecutionID(),
	}

	// Execute before transition hooks
	sm.executeHooks(BeforeTransition, *result)

	// Execute state exit hooks
	sm.executeHooks(OnStateExit, *result)

	// Execute transition action if present
	if transition.Action != nil {
		if err := transition.Action(sm.currentState, transition.To, event, sm.context); err != nil {
			result.Success = false
			result.Error = err
			sm.executeHooks(OnTransitionError, *result)
			return result, err
		}
	}

	// Update state
	sm.currentState = transition.To

	// Execute state enter hooks
	sm.executeHooks(OnStateEnter, *result)

	// Execute after transition hooks
	sm.executeHooks(AfterTransition, *result)

	return result, nil
}

// CanTransition checks if an event can trigger a transition from the current state
func (sm *StateMachine) CanTransition(event Event) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if !sm.running || !sm.events[event] {
		return false
	}

	key := transitionKey(sm.currentState, event)
	transition, exists := sm.transitions[key]

	if !exists {
		return false
	}

	// Check guard condition if present
	if transition.Condition != nil {
		return transition.Condition(sm.context)
	}

	return true
}

// GetValidEvents returns all events that can be triggered from the current state
func (sm *StateMachine) GetValidEvents() []Event {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var validEvents []Event

	for event := range sm.events {
		if sm.canTransitionUnsafe(event) {
			validEvents = append(validEvents, event)
		}
	}

	return validEvents
}

// canTransitionUnsafe is an internal method that doesn't acquire locks
func (sm *StateMachine) canTransitionUnsafe(event Event) bool {
	if !sm.running || !sm.events[event] {
		return false
	}

	key := transitionKey(sm.currentState, event)
	transition, exists := sm.transitions[key]

	if !exists {
		return false
	}

	if transition.Condition != nil {
		return transition.Condition(sm.context)
	}

	return true
}

// AddTransition adds a new transition to the machine
func (sm *StateMachine) AddTransition(transition Transition) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Validate states exist
	if !sm.states[transition.From] {
		return NewStateNotFoundError(transition.From)
	}
	if !sm.states[transition.To] {
		return NewStateNotFoundError(transition.To)
	}

	// Validate event exists
	if !sm.events[transition.Event] {
		return FSMError{
			Type:    "EventNotFound",
			Message: fmt.Sprintf("Event '%s' is not defined in this FSM", transition.Event),
			Event:   transition.Event,
		}
	}

	key := transitionKey(transition.From, transition.Event)
	sm.transitions[key] = transition

	return nil
}

// RemoveTransition removes a transition from the machine
func (sm *StateMachine) RemoveTransition(from State, event Event) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := transitionKey(from, event)
	if _, exists := sm.transitions[key]; !exists {
		return FSMError{
			Type:    "TransitionNotFound",
			Message: fmt.Sprintf("No transition found from state '%s' with event '%s'", from, event),
			State:   from,
			Event:   event,
		}
	}

	delete(sm.transitions, key)
	return nil
}

// GetTransitions returns all transitions in the machine
func (sm *StateMachine) GetTransitions() []Transition {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	transitions := make([]Transition, 0, len(sm.transitions))
	for _, transition := range sm.transitions {
		transitions = append(transitions, transition)
	}

	return transitions
}

// executeHooks executes all hooks of a given type
func (sm *StateMachine) executeHooks(hookType HookType, result TransitionResult) {
	if hooks, exists := sm.hooks[hookType]; exists {
		for _, hook := range hooks {
			hook(result, sm.context)
		}
	}
}

// AddHook adds a hook function for a specific hook type
func (sm *StateMachine) AddHook(hookType HookType, hook Hook) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.hooks[hookType] == nil {
		sm.hooks[hookType] = make([]Hook, 0)
	}
	sm.hooks[hookType] = append(sm.hooks[hookType], hook)
}

// RemoveHook removes all hooks of a specific type
func (sm *StateMachine) RemoveHook(hookType HookType) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.hooks, hookType)
}

// GetContext returns the machine's context
func (sm *StateMachine) GetContext() Context {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.context
}

// SetContext sets the machine's context
func (sm *StateMachine) SetContext(context Context) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.context = context
}

// Start initializes the machine with an initial state
func (sm *StateMachine) Start(initialState State) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.states[initialState] {
		return NewStateNotFoundError(initialState)
	}

	sm.initialState = initialState
	sm.currentState = initialState
	sm.running = true

	// Execute state enter hooks for initial state
	sm.executeHooks(OnStateEnter, TransitionResult{
		Success:     true,
		FromState:   "",
		ToState:     initialState,
		Timestamp:   time.Now(),
		ExecutionID: generateExecutionID(),
	})

	return nil
}

// Stop halts the machine
func (sm *StateMachine) Stop() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.running {
		// Execute state exit hooks for current state
		sm.executeHooks(OnStateExit, TransitionResult{
			Success:     true,
			FromState:   sm.currentState,
			ToState:     "",
			Timestamp:   time.Now(),
			ExecutionID: generateExecutionID(),
		})
	}

	sm.running = false
	return nil
}

// Reset resets the machine to its initial state
func (sm *StateMachine) Reset() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.initialState == "" {
		return FSMError{
			Type:    "NoInitialState",
			Message: "Cannot reset machine: no initial state defined",
		}
	}

	oldState := sm.currentState

	if sm.running && oldState != "" {
		// Execute state exit hooks for current state
		sm.executeHooks(OnStateExit, TransitionResult{
			Success:     true,
			FromState:   oldState,
			ToState:     sm.initialState,
			Timestamp:   time.Now(),
			ExecutionID: generateExecutionID(),
		})
	}

	sm.currentState = sm.initialState
	sm.running = true

	// Execute state enter hooks for initial state
	sm.executeHooks(OnStateEnter, TransitionResult{
		Success:     true,
		FromState:   oldState,
		ToState:     sm.initialState,
		Timestamp:   time.Now(),
		ExecutionID: generateExecutionID(),
	})

	return nil
}

// IsRunning returns whether the machine is currently running
func (sm *StateMachine) IsRunning() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.running
}

// Validate checks the machine configuration for consistency
func (sm *StateMachine) Validate() error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Check if there are any states
	if len(sm.states) == 0 {
		return FSMError{
			Type:    "NoStates",
			Message: "Machine has no states defined",
		}
	}

	// Check if there are any events
	if len(sm.events) == 0 {
		return FSMError{
			Type:    "NoEvents",
			Message: "Machine has no events defined",
		}
	}

	// Validate all transitions reference valid states and events
	for _, transition := range sm.transitions {
		if !sm.states[transition.From] {
			return NewStateNotFoundError(transition.From)
		}
		if !sm.states[transition.To] {
			return NewStateNotFoundError(transition.To)
		}
		if !sm.events[transition.Event] {
			return FSMError{
				Type:    "EventNotFound",
				Message: fmt.Sprintf("Event '%s' is not defined in this FSM", transition.Event),
				Event:   transition.Event,
			}
		}
	}

	return nil
}

// AddState adds a state to the machine
func (sm *StateMachine) AddState(state State) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.states[state] = true
}

// AddEvent adds an event to the machine
func (sm *StateMachine) AddEvent(event Event) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.events[event] = true
}
