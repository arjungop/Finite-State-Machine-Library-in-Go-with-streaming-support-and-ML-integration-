# Formal Language-Based Self-Programming AI Framework

A foundational autonomous software framework that bridges theoretical computer science with practical applications through Go-based finite state machines (FSM) and formal language principles.

## Core Philosophy

This framework implements symbolic AI through:
- Structured logical reasoning
- Deterministic state transitions  
- Formal grammars governing events and states
- Rule-based decision making

## Architecture

### Core Components
- **State Manager**: Maintains system state and prevents illegal transitions
- **Event Dispatcher**: Routes external triggers to the transition engine
- **Transition Engine**: Evaluates and executes state changes based on predefined rules
- **Hook System**: Enables side effects like logging and notifications
- **Builder Interface**: Provides fluent API for defining machine grammar
- **Concurrency Control**: Ensures thread-safe operation in distributed environments

### Key Innovations
- **Declarative Self-Programming**: Systems define their behavior through configuration
- **Autonomous Agents**: Software entities operating within formally defined boundaries
- **Runtime Reconfiguration**: Modify system behavior by altering state-event grammars


## Quick Start

```go
import "github.com/fla/self-programming-ai/fsm"

// Define states and events
states := []string{"idle", "processing", "completed"}
events := []string{"start", "finish", "reset"}

// Create FSM
machine := fsm.NewBuilder().
    AddStates(states...).
    AddEvents(events...).
    AddTransition("idle", "start", "processing").
    AddTransition("processing", "finish", "completed").
    AddTransition("completed", "reset", "idle").
    Build()

// Use the machine
machine.Start("idle")
machine.SendEvent("start") // transitions to "processing"
```

## How to Run

### 1. Build the project

```sh
go build ./...
```

### 2. Run all tests

```sh
go test ./...
```

### 3. Run an example FSM

You can run the example order processor or vending machine FSMs:

```sh
go run examples/order_processor.go
go run examples/vending_machine.go
```

### 4. Run benchmarks (optional)

```sh
go test -bench=. ./pkg/fsm
```

### 5. Run your own FSM

Create a new Go file, import the `fsm` package, and use the builder as shown above.

---

## Applications

- Order processing lifecycle systems
- Vending machine controllers  
- Device control protocols
- Workflow automation engines
- Robotics and autonomous vehicles
- Distributed microservices
- IoT device management

## Development Status

This is an active research and development project implementing autonomous software principles through formal methods.