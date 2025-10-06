// Package fsm provides finite state machine implementation with builder pattern
package fsm

import "fmt" // Standard library for string formatting and error messages

// FSMBuilder implements the Builder interface for fluent FSM construction
// This struct provides a chainable API for constructing finite state machines
type FSMBuilder struct {
	machine      *StateMachine // The state machine being constructed
	initialState State         // The state this FSM will start in when initialized
}

// NewBuilder creates a new FSM builder
// Factory function that returns a ready-to-use builder for constructing FSMs
func NewBuilder() Builder {
	return &FSMBuilder{
		machine: NewStateMachine(), // Create empty state machine to be configured
	}
}

// AddState adds a single state to the FSM
// This method adds one state to the set of valid states and returns the builder for chaining
func (b *FSMBuilder) AddState(state State) Builder {
	b.machine.AddState(state) // Register the state in the underlying state machine
	return b                  // Return builder to enable method chaining
}

// AddStates adds multiple states to the FSM
// Convenience method that accepts variadic parameters to add many states at once
func (b *FSMBuilder) AddStates(states ...State) Builder {
	for _, state := range states { // Iterate through all provided states
		b.machine.AddState(state) // Add each state to the underlying state machine
	}
	return b // Return builder to enable method chaining
}

// AddEvent adds a single event to the FSM
// This method adds one event to the set of valid events that can trigger transitions
func (b *FSMBuilder) AddEvent(event Event) Builder {
	b.machine.AddEvent(event) // Register the event in the underlying state machine
	return b                  // Return builder to enable method chaining
}

// AddEvents adds multiple events to the FSM
// Convenience method that accepts variadic parameters to add many events at once
func (b *FSMBuilder) AddEvents(events ...Event) Builder {
	for _, event := range events { // Iterate through all provided events
		b.machine.AddEvent(event) // Add each event to the underlying state machine
	}
	return b // Return builder to enable method chaining
}

// AddTransition adds a basic transition without conditions or actions
// Creates a simple state transition that occurs whenever the specified event is triggered
func (b *FSMBuilder) AddTransition(from State, event Event, to State) Builder {
	transition := Transition{ // Create transition structure with basic parameters
		From:  from,  // Source state where transition begins
		Event: event, // Event that triggers this transition
		To:    to,    // Destination state where transition ends
	}

	// Automatically add states and events if they don't exist
	b.machine.AddState(from)  // Ensure source state is registered in the FSM
	b.machine.AddState(to)    // Ensure destination state is registered in the FSM
	b.machine.AddEvent(event) // Ensure triggering event is registered in the FSM

	b.machine.AddTransition(transition) // Add the transition rule to the state machine
	return b                             // Return builder to enable method chaining
}

// AddTransitionWithCondition adds a transition with a guard condition
// Creates a conditional transition that only occurs if the condition function returns true
func (b *FSMBuilder) AddTransitionWithCondition(from State, event Event, to State, condition TransitionCondition) Builder {
	transition := Transition{ // Create transition structure with condition guard
		From:      from,      // Source state where transition begins
		Event:     event,     // Event that triggers this transition
		To:        to,        // Destination state where transition ends
		Condition: condition, // Guard function that must return true for transition to occur
	}

	// Automatically add states and events if they don't exist
	b.machine.AddState(from)  // Ensure source state is registered in the FSM
	b.machine.AddState(to)    // Ensure destination state is registered in the FSM
	b.machine.AddEvent(event) // Ensure triggering event is registered in the FSM

	b.machine.AddTransition(transition) // Add the conditional transition rule to the state machine
	return b                             // Return builder to enable method chaining
}

// AddTransitionWithAction adds a transition with an action
// Creates a transition that executes a specific function when the transition occurs
func (b *FSMBuilder) AddTransitionWithAction(from State, event Event, to State, action TransitionAction) Builder {
	transition := Transition{ // Create transition structure with action callback
		From:   from,   // Source state where transition begins
		Event:  event,  // Event that triggers this transition
		To:     to,     // Destination state where transition ends
		Action: action, // Function to execute when transition occurs
	}

	// Automatically add states and events if they don't exist
	b.machine.AddState(from)  // Ensure source state is registered in the FSM
	b.machine.AddState(to)    // Ensure destination state is registered in the FSM
	b.machine.AddEvent(event) // Ensure triggering event is registered in the FSM

	b.machine.AddTransition(transition)
	return b
}

// AddTransitionFull adds a transition with both condition and action
// Creates a fully-featured transition with both guard condition and action callback
func (b *FSMBuilder) AddTransitionFull(from State, event Event, to State, condition TransitionCondition, action TransitionAction) Builder {
	transition := Transition{ // Create transition structure with both condition and action
		From:      from,      // Source state where transition begins
		Event:     event,     // Event that triggers this transition
		To:        to,        // Destination state where transition ends
		Condition: condition, // Guard function that must return true for transition to occur
		Action:    action,    // Function to execute when transition occurs
	}

	// Automatically add states and events if they don't exist
	b.machine.AddState(from)  // Ensure source state is registered in the FSM
	b.machine.AddState(to)    // Ensure destination state is registered in the FSM
	b.machine.AddEvent(event) // Ensure triggering event is registered in the FSM

	b.machine.AddTransition(transition) // Add the full-featured transition rule to the state machine
	return b                             // Return builder to enable method chaining
}

// SetInitialState sets the initial state for the FSM
// Specifies which state the finite state machine should start in when initialized
func (b *FSMBuilder) SetInitialState(state State) Builder {
	b.initialState = state           // Store the initial state for use during Build()
	b.machine.AddState(state)        // Ensure the initial state is registered in the FSM
	return b                         // Return builder to enable method chaining
}

// Build creates and validates the FSM, returning it ready for use
// Final method in the builder chain that constructs the complete finite state machine
func (b *FSMBuilder) Build() (Machine, error) {
	// Validate the machine configuration
	if err := b.machine.Validate(); err != nil { // Check if FSM configuration is valid
		return nil, err // Return error if validation fails
	}

	// Set initial state if specified
	if b.initialState != "" {                                 // Check if initial state was configured
		if err := b.machine.Start(b.initialState); err != nil { // Start FSM in initial state
			return nil, err // Return error if starting fails
		}
	}

	return b.machine, nil // Return completed and validated state machine
}

// BuilderWithHooks extends the builder with hook functionality
type BuilderWithHooks struct {
	*FSMBuilder
}

// NewBuilderWithHooks creates a builder that supports adding hooks during construction
func NewBuilderWithHooks() *BuilderWithHooks {
	return &BuilderWithHooks{
		FSMBuilder: &FSMBuilder{
			machine: NewStateMachine(),
		},
	}
}

// AddBeforeTransitionHook adds a hook that executes before transitions
func (b *BuilderWithHooks) AddBeforeTransitionHook(hook Hook) *BuilderWithHooks {
	b.machine.AddHook(BeforeTransition, hook)
	return b
}

// AddAfterTransitionHook adds a hook that executes after transitions
func (b *BuilderWithHooks) AddAfterTransitionHook(hook Hook) *BuilderWithHooks {
	b.machine.AddHook(AfterTransition, hook)
	return b
}

// AddOnStateEnterHook adds a hook that executes when entering states
func (b *BuilderWithHooks) AddOnStateEnterHook(hook Hook) *BuilderWithHooks {
	b.machine.AddHook(OnStateEnter, hook)
	return b
}

// AddOnStateExitHook adds a hook that executes when exiting states
func (b *BuilderWithHooks) AddOnStateExitHook(hook Hook) *BuilderWithHooks {
	b.machine.AddHook(OnStateExit, hook)
	return b
}

// AddOnTransitionErrorHook adds a hook that executes when transitions fail
func (b *BuilderWithHooks) AddOnTransitionErrorHook(hook Hook) *BuilderWithHooks {
	b.machine.AddHook(OnTransitionError, hook)
	return b
}

// Override methods to return *BuilderWithHooks instead of Builder

// AddStates adds multiple states to the FSM
func (b *BuilderWithHooks) AddStates(states ...State) *BuilderWithHooks {
	b.FSMBuilder.AddStates(states...)
	return b
}

// AddEvents adds multiple events to the FSM
func (b *BuilderWithHooks) AddEvents(events ...Event) *BuilderWithHooks {
	b.FSMBuilder.AddEvents(events...)
	return b
}

// AddTransition adds a basic transition without conditions or actions
func (b *BuilderWithHooks) AddTransition(from State, event Event, to State) *BuilderWithHooks {
	b.FSMBuilder.AddTransition(from, event, to)
	return b
}

// AddTransitionWithAction adds a transition with an action
func (b *BuilderWithHooks) AddTransitionWithAction(from State, event Event, to State, action TransitionAction) *BuilderWithHooks {
	b.FSMBuilder.AddTransitionWithAction(from, event, to, action)
	return b
}

// AddTransitionWithCondition adds a transition with a condition
func (b *BuilderWithHooks) AddTransitionWithCondition(from State, event Event, to State, condition TransitionCondition) *BuilderWithHooks {
	b.FSMBuilder.AddTransitionWithCondition(from, event, to, condition)
	return b
}

// AddTransitionFull adds a transition with both condition and action
func (b *BuilderWithHooks) AddTransitionFull(from State, event Event, to State, condition TransitionCondition, action TransitionAction) *BuilderWithHooks {
	b.FSMBuilder.AddTransitionFull(from, event, to, condition, action)
	return b
}

// SetInitialState sets the initial state for the FSM
func (b *BuilderWithHooks) SetInitialState(state State) *BuilderWithHooks {
	b.FSMBuilder.SetInitialState(state)
	return b
}

// Common transition conditions that can be used with the builder

// AlwaysTrue is a condition that always allows transitions
func AlwaysTrue() TransitionCondition {
	return func(context Context) bool {
		return true
	}
}

// AlwaysFalse is a condition that never allows transitions
func AlwaysFalse() TransitionCondition {
	return func(context Context) bool {
		return false
	}
}

// ContextHasKey returns a condition that checks if a key exists in context
func ContextHasKey(key string) TransitionCondition {
	return func(context Context) bool {
		return context.Get(key) != nil
	}
}

// ContextEquals returns a condition that checks if a context value equals the expected value
func ContextEquals(key string, expectedValue interface{}) TransitionCondition {
	return func(context Context) bool {
		return context.Get(key) == expectedValue
	}
}

// ContextGreaterThan returns a condition that checks if a numeric context value is greater than threshold
func ContextGreaterThan(key string, threshold float64) TransitionCondition {
	return func(context Context) bool {
		value := context.Get(key)
		if floatValue, ok := value.(float64); ok {
			return floatValue > threshold
		}
		if intValue, ok := value.(int); ok {
			return float64(intValue) > threshold
		}
		return false
	}
}

// Common transition actions that can be used with the builder

// LogTransition creates an action that logs transition information
func LogTransition(logger func(string)) TransitionAction {
	return func(from, to State, event Event, context Context) error {
		logger(fmt.Sprintf("Transition: %s --%s--> %s", from, event, to))
		return nil
	}
}

// SetContextValue creates an action that sets a value in the context
func SetContextValue(key string, value interface{}) TransitionAction {
	return func(from, to State, event Event, context Context) error {
		context.Set(key, value)
		return nil
	}
}

// IncrementCounter creates an action that increments a counter in the context
func IncrementCounter(key string) TransitionAction {
	return func(from, to State, event Event, context Context) error {
		current := context.Get(key)
		if current == nil {
			context.Set(key, 1)
		} else if intValue, ok := current.(int); ok {
			context.Set(key, intValue+1)
		}
		return nil
	}
}
