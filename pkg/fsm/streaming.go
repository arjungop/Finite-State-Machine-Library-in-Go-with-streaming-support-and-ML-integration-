package fsm

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// EventStreamer handles distributed event processing for FSMs
type EventStreamer struct {
	machines    map[string]Machine
	subscribers map[string][]chan EventMessage
	publishers  map[string]chan EventMessage
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// EventMessage represents a distributed event
type EventMessage struct {
	ID          string                 `json:"id"`
	MachineID   string                 `json:"machine_id"`
	Event       string                 `json:"event"`
	Timestamp   time.Time              `json:"timestamp"`
	Context     map[string]interface{} `json:"context"`
	Source      string                 `json:"source"`
	Destination string                 `json:"destination"`
}

// EventHandler processes incoming events
type EventHandler func(msg EventMessage) error

// StreamConfig configures event streaming behavior
type StreamConfig struct {
	BufferSize    int
	RetryAttempts int
	RetryDelay    time.Duration
	Timeout       time.Duration
}

// NewEventStreamer creates a new event streaming system
func NewEventStreamer(config StreamConfig) *EventStreamer {
	ctx, cancel := context.WithCancel(context.Background())

	if config.BufferSize == 0 {
		config.BufferSize = 100
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = time.Second
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &EventStreamer{
		machines:    make(map[string]Machine),
		subscribers: make(map[string][]chan EventMessage),
		publishers:  make(map[string]chan EventMessage),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// RegisterMachine registers an FSM for event streaming
func (es *EventStreamer) RegisterMachine(id string, machine Machine) {
	es.mu.Lock()
	defer es.mu.Unlock()

	es.machines[id] = machine
	es.publishers[id] = make(chan EventMessage, 100)

	// Start event processor for this machine
	go es.processEvents(id)
}

// Subscribe to events from a specific machine
func (es *EventStreamer) Subscribe(machineID string, handler EventHandler) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	if _, exists := es.machines[machineID]; !exists {
		return fmt.Errorf("machine %s not registered", machineID)
	}

	subscriber := make(chan EventMessage, 100)
	es.subscribers[machineID] = append(es.subscribers[machineID], subscriber)

	// Start subscriber processor
	go es.handleSubscription(subscriber, handler)

	return nil
}

// PublishEvent publishes an event to the stream
func (es *EventStreamer) PublishEvent(msg EventMessage) error {
	es.mu.RLock()
	defer es.mu.RUnlock()

	if msg.ID == "" {
		msg.ID = generateEventID()
	}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	// Send to machine's publisher channel
	if publisher, exists := es.publishers[msg.MachineID]; exists {
		select {
		case publisher <- msg:
			return nil
		case <-time.After(5 * time.Second):
			return fmt.Errorf("timeout publishing event to machine %s", msg.MachineID)
		}
	}

	return fmt.Errorf("machine %s not found", msg.MachineID)
}

// BroadcastEvent sends an event to all registered machines
func (es *EventStreamer) BroadcastEvent(event string, context map[string]interface{}) error {
	es.mu.RLock()
	machines := make([]string, 0, len(es.machines))
	for id := range es.machines {
		machines = append(machines, id)
	}
	es.mu.RUnlock()

	var errors []error
	for _, machineID := range machines {
		msg := EventMessage{
			ID:        generateEventID(),
			MachineID: machineID,
			Event:     event,
			Timestamp: time.Now(),
			Context:   context,
			Source:    "broadcast",
		}

		if err := es.PublishEvent(msg); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("broadcast failed for %d machines", len(errors))
	}

	return nil
}

// processEvents handles events for a specific machine
func (es *EventStreamer) processEvents(machineID string) {
	es.mu.RLock()
	publisher := es.publishers[machineID]
	machine := es.machines[machineID]
	subscribers := es.subscribers[machineID]
	es.mu.RUnlock()

	for {
		select {
		case <-es.ctx.Done():
			return
		case msg := <-publisher:
			// Process event on the machine
			if err := es.processEventOnMachine(machine, msg); err != nil {
				// Handle error - could log or retry
				continue
			}

			// Notify subscribers
			for _, subscriber := range subscribers {
				select {
				case subscriber <- msg:
				default:
					// Subscriber channel full, skip
				}
			}
		}
	}
}

// processEventOnMachine applies an event to a specific machine
func (es *EventStreamer) processEventOnMachine(machine Machine, msg EventMessage) error {
	// Update machine context with event context
	if msg.Context != nil {
		machineContext := machine.GetContext()
		for key, value := range msg.Context {
			machineContext.Set(key, value)
		}
	}

	// Send event to machine
	machine.SendEvent(Event(msg.Event))

	return nil
}

// handleSubscription processes subscription events
func (es *EventStreamer) handleSubscription(subscriber chan EventMessage, handler EventHandler) {
	for {
		select {
		case <-es.ctx.Done():
			return
		case msg := <-subscriber:
			if err := handler(msg); err != nil {
				// Log error or handle as appropriate
			}
		}
	}
}

// Close stops the event streamer
func (es *EventStreamer) Close() error {
	es.cancel()

	es.mu.Lock()
	defer es.mu.Unlock()

	// Close all channels
	for _, publisher := range es.publishers {
		close(publisher)
	}

	for _, subscriberList := range es.subscribers {
		for _, subscriber := range subscriberList {
			close(subscriber)
		}
	}

	return nil
}

// EventSourcing provides event sourcing capabilities
type EventSourcing struct {
	events []EventMessage
	mu     sync.RWMutex
}

// NewEventSourcing creates a new event sourcing system
func NewEventSourcing() *EventSourcing {
	return &EventSourcing{
		events: make([]EventMessage, 0),
	}
}

// AppendEvent adds an event to the event store
func (es *EventSourcing) AppendEvent(event EventMessage) {
	es.mu.Lock()
	defer es.mu.Unlock()

	es.events = append(es.events, event)
}

// GetEvents retrieves events for a specific machine
func (es *EventSourcing) GetEvents(machineID string) []EventMessage {
	es.mu.RLock()
	defer es.mu.RUnlock()

	var machineEvents []EventMessage
	for _, event := range es.events {
		if event.MachineID == machineID {
			machineEvents = append(machineEvents, event)
		}
	}

	return machineEvents
}

// GetEventsAfter retrieves events after a specific timestamp
func (es *EventSourcing) GetEventsAfter(timestamp time.Time) []EventMessage {
	es.mu.RLock()
	defer es.mu.RUnlock()

	var recentEvents []EventMessage
	for _, event := range es.events {
		if event.Timestamp.After(timestamp) {
			recentEvents = append(recentEvents, event)
		}
	}

	return recentEvents
}

// ReplayEvents replays events on a machine to reconstruct state
func (es *EventSourcing) ReplayEvents(machine Machine, machineID string) error {
	events := es.GetEvents(machineID)

	for _, event := range events {
		// Apply context
		if event.Context != nil {
			context := machine.GetContext()
			for key, value := range event.Context {
				context.Set(key, value)
			}
		}

		// Send event
		machine.SendEvent(Event(event.Event))
	}

	return nil
}

// SerializeEvents converts events to JSON
func (es *EventSourcing) SerializeEvents() ([]byte, error) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	return json.Marshal(es.events)
}

// DeserializeEvents loads events from JSON
func (es *EventSourcing) DeserializeEvents(data []byte) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	return json.Unmarshal(data, &es.events)
}

// generateEventID creates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

// DistributedFSM represents a distributed finite state machine
type DistributedFSM struct {
	LocalMachine Machine
	Streamer     *EventStreamer
	Sourcing     *EventSourcing
	ID           string
}

// NewDistributedFSM creates a distributed FSM
func NewDistributedFSM(id string, machine Machine, streamer *EventStreamer) *DistributedFSM {
	dfsm := &DistributedFSM{
		LocalMachine: machine,
		Streamer:     streamer,
		Sourcing:     NewEventSourcing(),
		ID:           id,
	}

	// Register with streamer
	streamer.RegisterMachine(id, machine)

	return dfsm
}

// SendDistributedEvent sends an event that can trigger other machines
func (dfsm *DistributedFSM) SendDistributedEvent(event string, targetMachine string, context map[string]interface{}) error {
	msg := EventMessage{
		ID:          generateEventID(),
		MachineID:   targetMachine,
		Event:       event,
		Timestamp:   time.Now(),
		Context:     context,
		Source:      dfsm.ID,
		Destination: targetMachine,
	}

	// Store in event sourcing
	dfsm.Sourcing.AppendEvent(msg)

	// Publish to stream
	return dfsm.Streamer.PublishEvent(msg)
}

// CurrentState returns the current state of the local machine
func (dfsm *DistributedFSM) CurrentState() State {
	return dfsm.LocalMachine.CurrentState()
}

// GetContext returns the context of the local machine
func (dfsm *DistributedFSM) GetContext() Context {
	return dfsm.LocalMachine.GetContext()
}
