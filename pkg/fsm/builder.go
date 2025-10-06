package fsm

import "fmt"

// FSMBuilder implements the Builder interface for fluent FSM construction
type FSMBuilder struct {
	machine      *StateMachine
	initialState State
}

// NewBuilder creates a new FSM builder
func NewBuilder() Builder {
	return &FSMBuilder{
		machine: NewStateMachine(),
	}
}

// AddState adds a single state to the FSM
func (b *FSMBuilder) AddState(state State) Builder {
	b.machine.AddState(state)
	return b
}

// AddStates adds multiple states to the FSM
func (b *FSMBuilder) AddStates(states ...State) Builder {
	for _, state := range states {
		b.machine.AddState(state)
	}
	return b
}

// AddEvent adds a single event to the FSM
func (b *FSMBuilder) AddEvent(event Event) Builder {
	b.machine.AddEvent(event)
	return b
}

// AddEvents adds multiple events to the FSM
func (b *FSMBuilder) AddEvents(events ...Event) Builder {
	for _, event := range events {
		b.machine.AddEvent(event)
	}
	return b
}

// AddTransition adds a basic transition without conditions or actions
func (b *FSMBuilder) AddTransition(from State, event Event, to State) Builder {
	transition := Transition{
		From:  from,
		Event: event,
		To:    to,
	}
	
	// Automatically add states and events if they don't exist
	b.machine.AddState(from)
	b.machine.AddState(to)
	b.machine.AddEvent(event)
	
	b.machine.AddTransition(transition)
	return b
}

// AddTransitionWithCondition adds a transition with a guard condition
func (b *FSMBuilder) AddTransitionWithCondition(from State, event Event, to State, condition TransitionCondition) Builder {
	transition := Transition{
		From:      from,
		Event:     event,
		To:        to,
		Condition: condition,
	}
	
	// Automatically add states and events if they don't exist
	b.machine.AddState(from)
	b.machine.AddState(to)
	b.machine.AddEvent(event)
	
	b.machine.AddTransition(transition)
	return b
}

// AddTransitionWithAction adds a transition with an action
func (b *FSMBuilder) AddTransitionWithAction(from State, event Event, to State, action TransitionAction) Builder {
	transition := Transition{
		From:   from,
		Event:  event,
		To:     to,
		Action: action,
	}
	
	// Automatically add states and events if they don't exist
	b.machine.AddState(from)
	b.machine.AddState(to)
	b.machine.AddEvent(event)
	
	b.machine.AddTransition(transition)
	return b
}

// AddTransitionFull adds a transition with both condition and action
func (b *FSMBuilder) AddTransitionFull(from State, event Event, to State, condition TransitionCondition, action TransitionAction) Builder {
	transition := Transition{
		From:      from,
		Event:     event,
		To:        to,
		Condition: condition,
		Action:    action,
	}
	
	// Automatically add states and events if they don't exist
	b.machine.AddState(from)
	b.machine.AddState(to)
	b.machine.AddEvent(event)
	
	b.machine.AddTransition(transition)
	return b
}

// SetInitialState sets the initial state for the FSM
func (b *FSMBuilder) SetInitialState(state State) Builder {
	b.initialState = state
	b.machine.AddState(state)
	return b
}

// Build creates and validates the FSM, returning it ready for use
func (b *FSMBuilder) Build() (Machine, error) {
	// Validate the machine configuration
	if err := b.machine.Validate(); err != nil {
		return nil, err
	}
	
	// Set initial state if specified
	if b.initialState != "" {
		if err := b.machine.Start(b.initialState); err != nil {
			return nil, err
		}
	}
	
	return b.machine, nil
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