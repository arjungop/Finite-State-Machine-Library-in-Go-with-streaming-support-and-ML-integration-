package fsm

import (
	"testing"
	"time"
)

// BenchmarkStateMachine benchmarks basic state machine operations
func BenchmarkStateMachine(b *testing.B) {
	machine, err := NewBuilder().
		AddStates("idle", "active").
		AddEvents("activate", "deactivate").
		AddTransition("idle", "activate", "active").
		AddTransition("active", "deactivate", "idle").
		SetInitialState("idle").
		Build()

	if err != nil {
		b.Fatalf("Failed to build FSM: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if machine.CurrentState() == "idle" {
			machine.SendEvent("activate")
		} else {
			machine.SendEvent("deactivate")
		}
	}
}

// BenchmarkConcurrentAccess benchmarks concurrent access to state machine
func BenchmarkConcurrentAccess(b *testing.B) {
	machine, err := NewBuilder().
		AddStates("idle", "active").
		AddEvents("activate", "deactivate").
		AddTransition("idle", "activate", "active").
		AddTransition("active", "deactivate", "idle").
		SetInitialState("idle").
		Build()

	if err != nil {
		b.Fatalf("Failed to build FSM: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Mix of read and write operations
			machine.CurrentState()
			if machine.CanTransition("activate") {
				machine.SendEvent("activate")
			} else if machine.CanTransition("deactivate") {
				machine.SendEvent("deactivate")
			}
		}
	})
}

// BenchmarkBuilder benchmarks FSM construction
func BenchmarkBuilder(b *testing.B) {
	states := []State{"s1", "s2", "s3", "s4", "s5"}
	events := []Event{"e1", "e2", "e3", "e4", "e5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := NewBuilder()
		for _, state := range states {
			builder.AddState(state)
		}
		for _, event := range events {
			builder.AddEvent(event)
		}
		for j, state := range states[:len(states)-1] {
			builder.AddTransition(state, events[j], states[j+1])
		}
		builder.SetInitialState(states[0])
		builder.Build()
	}
}

// BenchmarkContext benchmarks context operations
func BenchmarkContext(b *testing.B) {
	context := NewContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key" + string(rune(i%10))
		context.Set(key, i)
		context.Get(key)
	}
}

// BenchmarkHooks benchmarks hook execution
func BenchmarkHooks(b *testing.B) {
	hookExecuted := 0
	hook := func(result TransitionResult, context Context) {
		hookExecuted++
	}

	machine, err := NewBuilder().
		AddStates("idle", "active").
		AddEvents("activate").
		AddTransition("idle", "activate", "active").
		SetInitialState("idle").
		Build()

	if err != nil {
		b.Fatalf("Failed to build FSM: %v", err)
	}

	machine.AddHook(AfterTransition, hook)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		machine.Reset()
		machine.SendEvent("activate")
	}
}

// TestPerformanceUnderLoad tests performance under sustained load
func TestPerformanceUnderLoad(t *testing.T) {
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

	// Test sustained load
	start := time.Now()
	iterations := 10000

	for i := 0; i < iterations; i++ {
		machine.SendEvent("start")
		machine.SendEvent("finish")
		machine.SendEvent("reset")
	}

	duration := time.Since(start)
	transitionsPerSecond := float64(iterations*3) / duration.Seconds()

	t.Logf("Processed %d transitions in %v (%.2f transitions/second)", 
		iterations*3, duration, transitionsPerSecond)

	// Performance should be reasonable (at least 1000 transitions/second)
	if transitionsPerSecond < 1000 {
		t.Errorf("Performance too slow: %.2f transitions/second", transitionsPerSecond)
	}
}

// TestMemoryUsage tests memory usage under load
func TestMemoryUsage(t *testing.T) {
	// Create many machines to test memory usage
	machines := make([]Machine, 1000)

	for i := 0; i < len(machines); i++ {
		machine, err := NewBuilder().
			AddStates("idle", "active").
			AddEvents("activate", "deactivate").
			AddTransition("idle", "activate", "active").
			AddTransition("active", "deactivate", "idle").
			SetInitialState("idle").
			Build()

		if err != nil {
			t.Fatalf("Failed to build FSM %d: %v", i, err)
		}

		machines[i] = machine
	}

	// Exercise all machines
	for _, machine := range machines {
		machine.SendEvent("activate")
		machine.SendEvent("deactivate")
	}

	// Machines should still be functional
	for i, machine := range machines {
		if machine.CurrentState() != "idle" {
			t.Errorf("Machine %d not in expected state after exercise", i)
		}
	}
}

// TestLargeStateMachine tests performance with a large state machine
func TestLargeStateMachine(t *testing.T) {
	builder := NewBuilder()

	// Create a large FSM with many states and transitions
	numStates := 100
	for i := 0; i < numStates; i++ {
		state := State(string(rune('A' + i%26)) + string(rune('A' + (i/26)%26)))
		builder.AddState(state)
	}

	// Add events
	events := []Event{"next", "prev", "jump"}
	for _, event := range events {
		builder.AddEvent(event)
	}

	// Add transitions (create a cycle)
	for i := 0; i < numStates-1; i++ {
		fromState := State(string(rune('A' + i%26)) + string(rune('A' + (i/26)%26)))
		toState := State(string(rune('A' + (i+1)%26)) + string(rune('A' + ((i+1)/26)%26)))
		builder.AddTransition(fromState, "next", toState)
	}

	// Close the cycle
	lastState := State(string(rune('A' + (numStates-1)%26)) + string(rune('A' + ((numStates-1)/26)%26)))
	firstState := State("AA")
	builder.AddTransition(lastState, "next", firstState)

	machine, err := builder.SetInitialState(firstState).Build()
	if err != nil {
		t.Fatalf("Failed to build large FSM: %v", err)
	}

	// Test performance with large FSM
	start := time.Now()
	for i := 0; i < 1000; i++ {
		machine.SendEvent("next")
	}
	duration := time.Since(start)

	t.Logf("Large FSM: 1000 transitions in %v", duration)

	// Should still be performant
	if duration > time.Second {
		t.Errorf("Large FSM too slow: %v for 1000 transitions", duration)
	}
}