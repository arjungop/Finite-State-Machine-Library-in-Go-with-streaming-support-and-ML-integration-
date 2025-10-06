package fsm

import (
	"errors"
	"testing"
	"time"
)

// TestBasicStateMachine tests basic FSM functionality
func TestBasicStateMachine(t *testing.T) {
	// Create a simple FSM
	machine, err := NewBuilder().
		AddStates("idle", "running", "stopped").
		AddEvents("start", "stop", "reset").
		AddTransition("idle", "start", "running").
		AddTransition("running", "stop", "stopped").
		AddTransition("stopped", "reset", "idle").
		SetInitialState("idle").
		Build()

	if err != nil {
		t.Fatalf("Failed to build FSM: %v", err)
	}

	// Test initial state
	if machine.CurrentState() != "idle" {
		t.Errorf("Expected initial state 'idle', got '%s'", machine.CurrentState())
	}

	// Test valid transition
	result, err := machine.SendEvent("start")
	if err != nil {
		t.Fatalf("Failed to send 'start' event: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected successful transition, got failure")
	}
	if machine.CurrentState() != "running" {
		t.Errorf("Expected state 'running', got '%s'", machine.CurrentState())
	}

	// Test invalid transition
	_, err = machine.SendEvent("start")
	if err == nil {
		t.Errorf("Expected error for invalid transition, got nil")
	}
}

// TestTransitionConditions tests guard conditions
func TestTransitionConditions(t *testing.T) {
	conditionMet := false
	condition := func(context Context) bool {
		return conditionMet
	}

	machine, err := NewBuilder().
		AddStates("waiting", "ready").
		AddEvents("check").
		AddTransitionWithCondition("waiting", "check", "ready", condition).
		SetInitialState("waiting").
		Build()

	if err != nil {
		t.Fatalf("Failed to build FSM: %v", err)
	}

	// Test condition not met
	_, err = machine.SendEvent("check")
	if err == nil {
		t.Errorf("Expected error when condition not met, got nil")
	}
	if machine.CurrentState() != "waiting" {
		t.Errorf("Expected state 'waiting', got '%s'", machine.CurrentState())
	}

	// Test condition met
	conditionMet = true
	result, err := machine.SendEvent("check")
	if err != nil {
		t.Fatalf("Failed to send event when condition met: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected successful transition when condition met")
	}
	if machine.CurrentState() != "ready" {
		t.Errorf("Expected state 'ready', got '%s'", machine.CurrentState())
	}
}

// TestTransitionActions tests transition actions
func TestTransitionActions(t *testing.T) {
	actionExecuted := false
	action := func(from, to State, event Event, context Context) error {
		actionExecuted = true
		context.Set("action_executed", true)
		return nil
	}

	machine, err := NewBuilder().
		AddStates("start", "end").
		AddEvents("go").
		AddTransitionWithAction("start", "go", "end", action).
		SetInitialState("start").
		Build()

	if err != nil {
		t.Fatalf("Failed to build FSM: %v", err)
	}

	// Test action execution
	_, err = machine.SendEvent("go")
	if err != nil {
		t.Fatalf("Failed to send event: %v", err)
	}

	if !actionExecuted {
		t.Errorf("Expected action to be executed")
	}

	if machine.GetContext().Get("action_executed") != true {
		t.Errorf("Expected context to be updated by action")
	}
}

// TestTransitionActionError tests error handling in actions
func TestTransitionActionError(t *testing.T) {
	actionError := errors.New("action failed")
	action := func(from, to State, event Event, context Context) error {
		return actionError
	}

	machine, err := NewBuilder().
		AddStates("start", "end").
		AddEvents("go").
		AddTransitionWithAction("start", "go", "end", action).
		SetInitialState("start").
		Build()

	if err != nil {
		t.Fatalf("Failed to build FSM: %v", err)
	}

	// Test action error
	result, err := machine.SendEvent("go")
	if err == nil {
		t.Errorf("Expected error from action, got nil")
	}
	if result.Success {
		t.Errorf("Expected failed transition due to action error")
	}
	if machine.CurrentState() != "start" {
		t.Errorf("Expected state to remain 'start' after action error, got '%s'", machine.CurrentState())
	}
}

// TestHooks tests the hook system
func TestHooks(t *testing.T) {
	var hookResults []string

	beforeHook := func(result TransitionResult, context Context) {
		hookResults = append(hookResults, "before")
	}

	afterHook := func(result TransitionResult, context Context) {
		hookResults = append(hookResults, "after")
	}

	enterHook := func(result TransitionResult, context Context) {
		hookResults = append(hookResults, "enter")
	}

	exitHook := func(result TransitionResult, context Context) {
		hookResults = append(hookResults, "exit")
	}

	machine := NewStateMachine()
	machine.AddState("a")
	machine.AddState("b")
	machine.AddEvent("move")
	machine.AddTransition(Transition{From: "a", Event: "move", To: "b"})

	machine.AddHook(BeforeTransition, beforeHook)
	machine.AddHook(AfterTransition, afterHook)
	machine.AddHook(OnStateEnter, enterHook)
	machine.AddHook(OnStateExit, exitHook)

	machine.Start("a")
	hookResults = nil // Reset after start hooks

	machine.SendEvent("move")

	expectedHooks := []string{"before", "exit", "enter", "after"}
	if len(hookResults) != len(expectedHooks) {
		t.Fatalf("Expected %d hooks, got %d: %v", len(expectedHooks), len(hookResults), hookResults)
	}

	for i, expected := range expectedHooks {
		if hookResults[i] != expected {
			t.Errorf("Expected hook %d to be '%s', got '%s'", i, expected, hookResults[i])
		}
	}
}

// TestConcurrency tests thread safety
func TestConcurrency(t *testing.T) {
	machine, err := NewBuilder().
		AddStates("idle", "busy").
		AddEvents("work", "done").
		AddTransition("idle", "work", "busy").
		AddTransition("busy", "done", "idle").
		SetInitialState("idle").
		Build()

	if err != nil {
		t.Fatalf("Failed to build FSM: %v", err)
	}

	// Run concurrent operations
	done := make(chan bool, 2)

	// Goroutine 1: Send events
	go func() {
		for i := 0; i < 100; i++ {
			if machine.CurrentState() == "idle" {
				machine.SendEvent("work")
			} else {
				machine.SendEvent("done")
			}
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Goroutine 2: Check state
	go func() {
		for i := 0; i < 100; i++ {
			state := machine.CurrentState()
			if state != "idle" && state != "busy" {
				t.Errorf("Invalid state detected: %s", state)
			}
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done
}

// TestContext tests context operations
func TestContext(t *testing.T) {
	context := NewContext()

	// Test Set and Get
	context.Set("key1", "value1")
	context.Set("key2", 42)

	if context.Get("key1") != "value1" {
		t.Errorf("Expected 'value1', got '%v'", context.Get("key1"))
	}

	if context.Get("key2") != 42 {
		t.Errorf("Expected 42, got %v", context.Get("key2"))
	}

	if context.Get("nonexistent") != nil {
		t.Errorf("Expected nil for nonexistent key, got %v", context.Get("nonexistent"))
	}

	// Test GetAll
	allData := context.GetAll()
	if len(allData) != 2 {
		t.Errorf("Expected 2 items in context, got %d", len(allData))
	}

	if allData["key1"] != "value1" {
		t.Errorf("Expected 'value1' in GetAll, got '%v'", allData["key1"])
	}
}

// TestValidation tests FSM validation
func TestValidation(t *testing.T) {
	// Test empty FSM
	machine := NewStateMachine()
	err := machine.Validate()
	if err == nil {
		t.Errorf("Expected validation error for empty FSM")
	}

	// Test FSM with states but no events
	machine.AddState("state1")
	err = machine.Validate()
	if err == nil {
		t.Errorf("Expected validation error for FSM with no events")
	}

	// Test valid FSM
	machine.AddEvent("event1")
	machine.AddTransition(Transition{From: "state1", Event: "event1", To: "state1"})
	err = machine.Validate()
	if err != nil {
		t.Errorf("Expected no validation error for valid FSM, got: %v", err)
	}
}

// TestMachineLifecycle tests start, stop, and reset operations
func TestMachineLifecycle(t *testing.T) {
	machine, err := NewBuilder().
		AddStates("init", "active", "inactive").
		AddEvents("activate", "deactivate").
		AddTransition("init", "activate", "active").
		AddTransition("active", "deactivate", "inactive").
		Build()

	if err != nil {
		t.Fatalf("Failed to build FSM: %v", err)
	}

	// Test start
	if machine.IsRunning() {
		t.Errorf("Expected machine to not be running initially")
	}

	err = machine.Start("init")
	if err != nil {
		t.Fatalf("Failed to start machine: %v", err)
	}

	if !machine.IsRunning() {
		t.Errorf("Expected machine to be running after start")
	}

	if machine.CurrentState() != "init" {
		t.Errorf("Expected current state 'init', got '%s'", machine.CurrentState())
	}

	// Test transition
	machine.SendEvent("activate")
	if machine.CurrentState() != "active" {
		t.Errorf("Expected current state 'active', got '%s'", machine.CurrentState())
	}

	// Test stop
	err = machine.Stop()
	if err != nil {
		t.Fatalf("Failed to stop machine: %v", err)
	}

	if machine.IsRunning() {
		t.Errorf("Expected machine to not be running after stop")
	}

	// Test sending event to stopped machine
	_, err = machine.SendEvent("deactivate")
	if err == nil {
		t.Errorf("Expected error when sending event to stopped machine")
	}

	// Test reset
	machine.Start("init")
	machine.SendEvent("activate")
	
	err = machine.Reset()
	if err != nil {
		t.Fatalf("Failed to reset machine: %v", err)
	}

	if machine.CurrentState() != "init" {
		t.Errorf("Expected current state 'init' after reset, got '%s'", machine.CurrentState())
	}
}

// TestBuilderValidation tests builder validation
func TestBuilderValidation(t *testing.T) {
	// Test building FSM without states
	_, err := NewBuilder().
		AddEvents("event1").
		Build()

	if err == nil {
		t.Errorf("Expected error when building FSM without states")
	}

	// Test building FSM without events
	_, err = NewBuilder().
		AddStates("state1").
		Build()

	if err == nil {
		t.Errorf("Expected error when building FSM without events")
	}

	// Test automatic state/event addition
	machine, err := NewBuilder().
		AddTransition("auto_state1", "auto_event1", "auto_state2").
		Build()

	if err != nil {
		t.Fatalf("Failed to build FSM with auto-added states/events: %v", err)
	}

	if !machine.IsValidState("auto_state1") {
		t.Errorf("Expected auto-added state 'auto_state1' to be valid")
	}

	if !machine.IsValidState("auto_state2") {
		t.Errorf("Expected auto-added state 'auto_state2' to be valid")
	}
}

// TestGetValidEvents tests the GetValidEvents method
func TestGetValidEvents(t *testing.T) {
	machine, err := NewBuilder().
		AddStates("idle", "working", "done").
		AddEvents("start", "finish", "reset").
		AddTransition("idle", "start", "working").
		AddTransition("working", "finish", "done").
		AddTransition("done", "reset", "idle").
		SetInitialState("idle").
		Build()

	if err != nil {
		t.Fatalf("Failed to build FSM: %v", err)
	}

	// Test valid events from idle state
	validEvents := machine.GetValidEvents()
	if len(validEvents) != 1 {
		t.Errorf("Expected 1 valid event from idle state, got %d", len(validEvents))
	}
	if validEvents[0] != "start" {
		t.Errorf("Expected valid event 'start', got '%s'", validEvents[0])
	}

	// Transition to working state
	machine.SendEvent("start")
	validEvents = machine.GetValidEvents()
	if len(validEvents) != 1 {
		t.Errorf("Expected 1 valid event from working state, got %d", len(validEvents))
	}
	if validEvents[0] != "finish" {
		t.Errorf("Expected valid event 'finish', got '%s'", validEvents[0])
	}
}

// TestCanTransition tests the CanTransition method
func TestCanTransition(t *testing.T) {
	machine, err := NewBuilder().
		AddStates("state1", "state2").
		AddEvents("event1", "event2").
		AddTransition("state1", "event1", "state2").
		SetInitialState("state1").
		Build()

	if err != nil {
		t.Fatalf("Failed to build FSM: %v", err)
	}

	// Test valid transition
	if !machine.CanTransition("event1") {
		t.Errorf("Expected CanTransition to return true for valid transition")
	}

	// Test invalid transition
	if machine.CanTransition("event2") {
		t.Errorf("Expected CanTransition to return false for invalid transition")
	}

	// Test nonexistent event
	if machine.CanTransition("nonexistent") {
		t.Errorf("Expected CanTransition to return false for nonexistent event")
	}
}