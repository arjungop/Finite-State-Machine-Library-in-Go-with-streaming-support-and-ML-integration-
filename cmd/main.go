package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fla/self-programming-ai/pkg/fsm"
	"github.com/fla/self-programming-ai/pkg/ml"
	"github.com/fla/self-programming-ai/pkg/web"
)

// FormalLanguageAI represents the complete self-programming AI system
type FormalLanguageAI struct {
	configLoader     *fsm.ConfigLoader
	nlParser         *fsm.NaturalLanguageParser
	eventStreamer    *fsm.EventStreamer
	mlAssistant      *ml.MLAssistant
	visualizationSvr *web.AdvancedVisualizationServer
	machines         map[string]fsm.Machine
	distributedFSMs  map[string]*fsm.DistributedFSM
	ctx              context.Context
	cancel           context.CancelFunc
}

// NewFormalLanguageAI creates the complete AI system
func NewFormalLanguageAI() *FormalLanguageAI {
	ctx, cancel := context.WithCancel(context.Background())
	
	streamConfig := fsm.StreamConfig{
		BufferSize:    200,
		RetryAttempts: 5,
		RetryDelay:    time.Second,
		Timeout:       30 * time.Second,
	}
	
	return &FormalLanguageAI{
		configLoader:     fsm.NewConfigLoader(),
		nlParser:         fsm.NewNaturalLanguageParser(),
		eventStreamer:    fsm.NewEventStreamer(streamConfig),
		mlAssistant:      ml.NewMLAssistant(),
		visualizationSvr: web.NewAdvancedVisualizationServer(8080),
		machines:         make(map[string]fsm.Machine),
		distributedFSMs:  make(map[string]*fsm.DistributedFSM),
		ctx:              ctx,
		cancel:           cancel,
	}
}

// Start initializes and starts the complete AI system
func (fla *FormalLanguageAI) Start() error {
	log.Println("Starting Formal Language-Based Self-Programming AI Framework")
	
	// Start web visualization server
	go func() {
	log.Println("Starting web visualization server on port 8080...")
		if err := fla.visualizationSvr.Start(); err != nil {
			log.Printf("Visualization server error: %v", err)
		}
	}()
	
	// Wait for server to start
	time.Sleep(2 * time.Second)
	
	// Create core system FSMs
	if err := fla.createCoreSystems(); err != nil {
		return fmt.Errorf("failed to create core systems: %w", err)
	}
	
	// Load configuration-based FSMs
	if err := fla.loadConfiguredFSMs(); err != nil {
		log.Printf("Warning: failed to load some configured FSMs: %v", err)
	}
	
	// Create example FSMs from natural language
	if err := fla.createNaturalLanguageFSMs(); err != nil {
		log.Printf("Warning: failed to create some NL FSMs: %v", err)
	}
	
	// Start autonomous operations
	fla.startAutonomousOperations()
	
	// Start ML training
	fla.startMLTraining()
	
	log.Println("Formal Language AI system fully operational")
	log.Println("Web dashboard: http://localhost:8080")
	log.Println("ML-assisted optimization: Active")
	log.Println("Event streaming: Active")
	log.Println("Dynamic configuration: Ready")
	
	return nil
}

// createCoreSystems creates the fundamental system FSMs
func (fla *FormalLanguageAI) createCoreSystems() error {
	log.Println("Creating core system FSMs...")
	
	// System Monitor FSM
	systemMonitor, err := fsm.NewBuilderWithHooks().
		AddStates("initializing", "monitoring", "analyzing", "optimizing", "alerting", "maintenance").
		AddEvents("start_monitoring", "analyze", "optimize", "alert", "maintain", "reset").
		
		AddTransitionWithAction("initializing", "start_monitoring", "monitoring",
			func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
				log.Printf("Transition: %s -> %s on %s", from, to, event)
				return nil
			}).
		AddTransitionWithCondition("monitoring", "analyze", "analyzing",
			func(context fsm.Context) bool {
				return context.Get("performance_data") != nil
			}).
		AddTransitionWithAction("analyzing", "optimize", "optimizing",
			fla.createOptimizationAction()).
		AddTransitionWithCondition("optimizing", "alert", "alerting",
			fla.createAlertCondition()).
		AddTransition("alerting", "maintain", "maintenance").
		AddTransition("maintenance", "reset", "monitoring").
		AddTransition("optimizing", "reset", "monitoring").
		
		SetInitialState("initializing").
		Build()
	
	if err != nil {
		return fmt.Errorf("failed to create system monitor: %w", err)
	}
	
	// Add hooks after creation
	systemMonitor.AddHook(fsm.OnStateEnter, fla.createSystemMonitorHook())
	
	fla.registerMachine("system_monitor", systemMonitor)
	
	// Resource Manager FSM
	resourceManager, err := fsm.NewBuilderWithHooks().
		AddStates("idle", "allocating", "monitoring_usage", "rebalancing", "cleanup").
		AddEvents("allocate", "monitor", "rebalance", "cleanup", "reset").
		
		AddTransitionWithAction("idle", "allocate", "allocating",
			fla.createResourceAllocationAction()).
		AddTransition("allocating", "monitor", "monitoring_usage").
		AddTransitionWithCondition("monitoring_usage", "rebalance", "rebalancing",
			fla.createRebalanceCondition()).
		AddTransition("rebalancing", "cleanup", "cleanup").
		AddTransition("cleanup", "reset", "idle").
		AddTransition("monitoring_usage", "reset", "idle").
		
		SetInitialState("idle").
		Build()
	
	if err != nil {
		return fmt.Errorf("failed to create resource manager: %w", err)
	}
	
	fla.registerMachine("resource_manager", resourceManager)
	
	// Event Coordinator FSM
	eventCoordinator, err := fsm.NewBuilderWithHooks().
		AddStates("listening", "processing", "routing", "aggregating", "responding").
		AddEvents("receive_event", "process", "route", "aggregate", "respond", "reset").
		
		AddTransition("listening", "receive_event", "processing").
		AddTransitionWithAction("processing", "process", "routing",
			fla.createEventProcessingAction()).
		AddTransition("routing", "route", "aggregating").
		AddTransitionWithAction("aggregating", "aggregate", "responding",
			fla.createEventAggregationAction()).
		AddTransition("responding", "respond", "listening").
		
		SetInitialState("listening").
		Build()
	
	if err != nil {
		return fmt.Errorf("failed to create event coordinator: %w", err)
	}
	
	// Add hooks after creation
	eventCoordinator.AddHook(fsm.AfterTransition, fla.createEventCoordinatorHook())
	
	fla.registerMachine("event_coordinator", eventCoordinator)
	
	log.Println("✅ Core systems created successfully")
	return nil
}

// loadConfiguredFSMs loads FSMs from configuration files
func (fla *FormalLanguageAI) loadConfiguredFSMs() error {
	log.Println("📁 Loading configured FSMs...")
	
	configFiles := []struct {
		name string
		file string
	}{
		{"smart_device", "configs/smart_device.yaml"},
		{"autonomous_vehicle", "configs/autonomous_vehicle.json"},
	}
	
	for _, cfg := range configFiles {
		if _, err := os.Stat(cfg.file); os.IsNotExist(err) {
			log.Printf("⚠️  Configuration file not found: %s", cfg.file)
			continue
		}
		
		var config *fsm.ConfigMachine
		var err error
		
		if strings.HasSuffix(cfg.file, ".yaml") {
			config, err = fla.configLoader.LoadFromYAML(cfg.file)
		} else {
			config, err = fla.configLoader.LoadFromJSON(cfg.file)
		}
		
		if err != nil {
			log.Printf("⚠️  Failed to load %s: %v", cfg.file, err)
			continue
		}
		
		machine, err := fla.configLoader.BuildMachine(config)
		if err != nil {
			log.Printf("⚠️  Failed to build %s: %v", cfg.name, err)
			continue
		}
		
		fla.registerMachine(cfg.name, machine)
		log.Printf("✅ Loaded %s from %s", cfg.name, cfg.file)
	}
	
	return nil
}

// createNaturalLanguageFSMs creates FSMs from natural language descriptions
func (fla *FormalLanguageAI) createNaturalLanguageFSMs() error {
	log.Println("🗣️  Creating FSMs from natural language...")
	
	examples := []struct {
		name        string
		description string
	}{
		{
			"task_scheduler",
			`States: idle, scheduling, executing, waiting, completed, failed
			 Events: schedule_task, execute, wait, complete, fail, retry, reset
			 From idle to scheduling when schedule_task
			 From scheduling to executing when execute
			 From executing to waiting when wait
			 From waiting to executing when execute
			 From executing to completed when complete
			 From executing to failed when fail
			 From failed to scheduling when retry
			 From completed to idle when reset`,
		},
		{
			"communication_hub",
			`States: offline, connecting, connected, transmitting, receiving, error
			 Events: connect, disconnect, send, receive, error_occur, recover
			 From offline to connecting when connect
			 From connecting to connected when connect
			 From connected to transmitting when send
			 From connected to receiving when receive
			 From transmitting to connected when send
			 From receiving to connected when receive
			 From connected to error when error_occur
			 From error to connecting when recover
			 From connected to offline when disconnect`,
		},
	}
	
	for _, example := range examples {
		config, err := fla.nlParser.ParseDescription(example.description)
		if err != nil {
			log.Printf("⚠️  Failed to parse %s: %v", example.name, err)
			continue
		}
		
		machine, err := fla.configLoader.BuildMachine(config)
		if err != nil {
			log.Printf("⚠️  Failed to build %s: %v", example.name, err)
			continue
		}
		
		fla.registerMachine(example.name, machine)
		log.Printf("✅ Created %s from natural language", example.name)
	}
	
	return nil
}

// registerMachine registers a machine with all components
func (fla *FormalLanguageAI) registerMachine(name string, machine fsm.Machine) {
	fla.machines[name] = machine
	fla.visualizationSvr.RegisterMachine(name, machine)
	fla.mlAssistant.AttachToMachine(machine)
	
	// Create distributed FSM wrapper
	dfsm := fsm.NewDistributedFSM(name, machine, fla.eventStreamer)
	fla.distributedFSMs[name] = dfsm
}

// startAutonomousOperations begins autonomous system operations
func (fla *FormalLanguageAI) startAutonomousOperations() {
	log.Println("🤖 Starting autonomous operations...")
	
	for name, machine := range fla.machines {
		go fla.runAutonomousMachine(name, machine)
	}
	
	// Start distributed event coordination
	go fla.runDistributedCoordination()
}

// runAutonomousMachine runs a machine autonomously with ML guidance
func (fla *FormalLanguageAI) runAutonomousMachine(name string, machine fsm.Machine) {
	ticker := time.NewTicker(time.Duration(1000+rand.Intn(2000)) * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-fla.ctx.Done():
			return
		case <-ticker.C:
			if !machine.IsRunning() {
				continue
			}
			
			validEvents := machine.GetValidEvents()
			if len(validEvents) == 0 {
				continue
			}
			
			// Get ML prediction
			prediction := fla.mlAssistant.PredictNextTransition(
				machine.CurrentState(),
				validEvents,
				machine.GetContext(),
			)
			
			// Use ML recommendation with some randomness for exploration
			var eventToSend fsm.Event
			if prediction.Confidence > 0.6 && rand.Float64() > 0.2 {
				eventToSend = prediction.RecommendedEvent
			} else {
				eventToSend = validEvents[rand.Intn(len(validEvents))]
			}
			
			// Send event
			if _, err := machine.SendEvent(eventToSend); err != nil {
				log.Printf("Error sending event %s to %s: %v", eventToSend, name, err)
			}
			
			// Occasionally trigger distributed events
			if rand.Float64() < 0.1 {
				fla.triggerDistributedEvent(name)
			}
		}
	}
}

// runDistributedCoordination coordinates distributed FSM operations
func (fla *FormalLanguageAI) runDistributedCoordination() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-fla.ctx.Done():
			return
		case <-ticker.C:
			// Broadcast system status updates
			context := map[string]interface{}{
				"timestamp":     time.Now().Unix(),
				"active_fsms":   len(fla.machines),
				"system_status": "operational",
			}
			
			if err := fla.eventStreamer.BroadcastEvent("system_heartbeat", context); err != nil {
				log.Printf("Failed to broadcast heartbeat: %v", err)
			}
		}
	}
}

// triggerDistributedEvent triggers cross-machine events
func (fla *FormalLanguageAI) triggerDistributedEvent(sourceMachine string) {
	if len(fla.distributedFSMs) < 2 {
		return
	}
	
	// Pick a random target machine
	var targets []string
	for name := range fla.distributedFSMs {
		if name != sourceMachine {
			targets = append(targets, name)
		}
	}
	
	if len(targets) == 0 {
		return
	}
	
	targetMachine := targets[rand.Intn(len(targets))]
	dfsm := fla.distributedFSMs[sourceMachine]
	
	// Send a coordination event
	context := map[string]interface{}{
		"coordination_request": true,
		"source_state":        string(dfsm.CurrentState()),
		"timestamp":          time.Now().Unix(),
	}
	
	events := []string{"coordinate", "sync", "update", "notify"}
	event := events[rand.Intn(len(events))]
	
	if err := dfsm.SendDistributedEvent(event, targetMachine, context); err != nil {
		log.Printf("Failed to send distributed event: %v", err)
	}
}

// startMLTraining begins continuous ML training
func (fla *FormalLanguageAI) startMLTraining() {
	log.Println("🧠 Starting ML training and optimization...")
	
	// Train from any existing historical data
	fla.mlAssistant.TrainFromHistory()
	
	// Periodic optimization
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-fla.ctx.Done():
				return
			case <-ticker.C:
				fla.performMLOptimization()
			}
		}
	}()
}

// performMLOptimization performs periodic ML optimization
func (fla *FormalLanguageAI) performMLOptimization() {
	for name := range fla.machines {
		suggestions := fla.mlAssistant.OptimizeMachine(fla.machines[name])
		
		if len(suggestions) > 0 {
			log.Printf("🧠 ML suggestions for %s:", name)
			for i, suggestion := range suggestions {
				if i >= 3 { // Limit output
					break
				}
				log.Printf("  • [%s] %s (%.1f%% confidence)",
					suggestion.Priority, suggestion.Description, suggestion.Confidence*100)
			}
		}
	}
	
	// Retrain neural network periodically
	fla.mlAssistant.TrainFromHistory()
}

// Action and condition factories
func (fla *FormalLanguageAI) createOptimizationAction() fsm.TransitionAction {
	return func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
		log.Printf("🔧 System optimization triggered from %s", from)
		context.Set("optimization_timestamp", time.Now().Unix())
		if context.Get("optimization_level") == nil {
			context.Set("optimization_level", 0)
		}
		context.Set("optimization_level", context.Get("optimization_level").(int)+1)
		return nil
	}
}

func (fla *FormalLanguageAI) createAlertCondition() fsm.TransitionCondition {
	return func(context fsm.Context) bool {
		// Alert if optimization level is high
		level := context.Get("optimization_level")
		return level != nil && level.(int) > 3
	}
}

func (fla *FormalLanguageAI) createResourceAllocationAction() fsm.TransitionAction {
	return func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
		log.Printf("💾 Resource allocation: %s -> %s", from, to)
		context.Set("allocated_resources", rand.Intn(100)+50)
		return nil
	}
}

func (fla *FormalLanguageAI) createRebalanceCondition() fsm.TransitionCondition {
	return func(context fsm.Context) bool {
		resources := context.Get("allocated_resources")
		return resources != nil && resources.(int) > 120
	}
}

func (fla *FormalLanguageAI) createEventProcessingAction() fsm.TransitionAction {
	return func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
		log.Printf("📡 Event processing: %s", event)
		if context.Get("processed_events") == nil {
			context.Set("processed_events", 0)
		}
		context.Set("processed_events", context.Get("processed_events").(int)+1)
		return nil
	}
}

func (fla *FormalLanguageAI) createEventAggregationAction() fsm.TransitionAction {
	return func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
		count := 0
		if context.Get("processed_events") != nil {
			count = context.Get("processed_events").(int)
		}
		log.Printf("📊 Event aggregation: %d events processed", count)
		return nil
	}
}

func (fla *FormalLanguageAI) createSystemMonitorHook() fsm.Hook {
	return func(result fsm.TransitionResult, context fsm.Context) {
		log.Printf("📊 System Monitor: %s -> %s", result.FromState, result.ToState)
		
		// Auto-trigger next operations
		switch string(result.ToState) {
		case "monitoring":
			go func() {
				time.Sleep(2 * time.Second)
				context.Set("performance_data", map[string]interface{}{
					"cpu_usage":    rand.Float64(),
					"memory_usage": rand.Float64(),
					"response_time": rand.Intn(1000),
				})
				if machine, ok := fla.machines["system_monitor"]; ok {
					machine.SendEvent("analyze")
				}
			}()
		case "analyzing":
			go func() {
				time.Sleep(1 * time.Second)
				if machine, ok := fla.machines["system_monitor"]; ok {
					machine.SendEvent("optimize")
				}
			}()
		}
	}
}

func (fla *FormalLanguageAI) createEventCoordinatorHook() fsm.Hook {
	return func(result fsm.TransitionResult, context fsm.Context) {
		if string(result.ToState) == "processing" {
			// Initialize processed events counter
			if context.Get("processed_events") == nil {
				context.Set("processed_events", 0)
			}
		}
	}
}

// Stop gracefully shuts down the AI system
func (fla *FormalLanguageAI) Stop() error {
	log.Println("🛑 Shutting down Formal Language AI system...")
	
	fla.cancel()
	
	// Stop event streamer
	if err := fla.eventStreamer.Close(); err != nil {
		log.Printf("Error closing event streamer: %v", err)
	}
	
	// Stop all machines
	for name, machine := range fla.machines {
		log.Printf("Stopping machine: %s", name)
		machine.Stop()
	}
	
	log.Println("System shutdown complete")
	return nil
}

// ShowStatus displays current system status
func (fla *FormalLanguageAI) ShowStatus() {
	fmt.Println("\nFORMAL LANGUAGE AI SYSTEM STATUS")
	fmt.Println("================================")
    
	fmt.Printf("Active Machines: %d\n", len(fla.machines))
	fmt.Printf("Distributed FSMs: %d\n", len(fla.distributedFSMs))
	fmt.Printf("ML Assistant: Active\n")
	fmt.Printf("Event Streaming: Active\n")
	fmt.Printf("Web Dashboard: http://localhost:8080\n")
    
	fmt.Println("\nMachine States:")
	for name, machine := range fla.machines {
		status := "Running"
		if !machine.IsRunning() {
			status = "Stopped"
		}
		fmt.Printf("  %s: %s (State: %s)\n", name, status, machine.CurrentState())
	}
	
	// ML Statistics
	stats := fla.mlAssistant.GetLearningStats()
	fmt.Println("\nML Learning Statistics:")
	for key, value := range stats {
		fmt.Printf("  %s: %v\n", key, value)
	}
	
	fmt.Println()
}

// Interactive CLI
func (fla *FormalLanguageAI) RunInteractiveCLI() {
	scanner := bufio.NewScanner(os.Stdin)
	
	fmt.Println("\nFORMAL LANGUAGE AI - INTERACTIVE MODE")
	fmt.Println("Commands: status, optimize, create <name>, start <name>, stop <name>, help, quit")
	
	for {
		fmt.Print("\nFLA> ")
		if !scanner.Scan() {
			break
		}
		
		command := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(command)
		
		if len(parts) == 0 {
			continue
		}
		
		switch parts[0] {
		case "status":
			fla.ShowStatus()
			
		case "optimize":
			log.Println("Running manual optimization...")
			fla.performMLOptimization()
			
		case "create":
			if len(parts) < 2 {
				fmt.Println("Usage: create <fsm_name>")
				continue
			}
			fla.createInteractiveFSM(parts[1])
			
		case "start":
			if len(parts) < 2 {
				fmt.Println("Usage: start <machine_name>")
				continue
			}
			if machine, exists := fla.machines[parts[1]]; exists {
				if err := machine.Start(machine.CurrentState()); err != nil {
					fmt.Printf("Failed to start machine: %v\n", err)
				} else {
					fmt.Printf("Started machine: %s\n", parts[1])
				}
			} else {
				fmt.Printf("Machine not found: %s\n", parts[1])
			}
			
		case "stop":
			if len(parts) < 2 {
				fmt.Println("Usage: stop <machine_name>")
				continue
			}
			if machine, exists := fla.machines[parts[1]]; exists {
				machine.Stop()
				fmt.Printf("Stopped machine: %s\n", parts[1])
			} else {
				fmt.Printf("Machine not found: %s\n", parts[1])
			}
			
		case "help":
			fmt.Println("Available commands:")
			fmt.Println("  status           - Show system status")
			fmt.Println("  optimize         - Run ML optimization")
			fmt.Println("  create <name>    - Create new FSM interactively")
			fmt.Println("  start <name>     - Start a machine")
			fmt.Println("  stop <name>      - Stop a machine")
			fmt.Println("  help             - Show this help")
			fmt.Println("  quit             - Exit interactive mode")
			
		case "quit", "exit":
			return
			
		default:
			fmt.Printf("Unknown command: %s (type 'help' for commands)\n", parts[0])
		}
	}
}

// createInteractiveFSM creates an FSM interactively
func (fla *FormalLanguageAI) createInteractiveFSM(name string) {
	fmt.Printf("Creating FSM: %s\n", name)
	fmt.Print("Enter natural language description: ")
	
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}
	
	description := scanner.Text()
	config, err := fla.nlParser.ParseDescription(description)
	if err != nil {
		fmt.Printf("❌ Failed to parse description: %v\n", err)
		return
	}
	
	machine, err := fla.configLoader.BuildMachine(config)
	if err != nil {
		fmt.Printf("❌ Failed to build machine: %v\n", err)
		return
	}
	
	fla.registerMachine(name, machine)
	fmt.Printf("✅ Created and registered FSM: %s\n", name)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	fmt.Println("FORMAL LANGUAGE-BASED SELF-PROGRAMMING AI FRAMEWORK")
	fmt.Println("    Complete 100% Implementation")
	fmt.Println("====================================================")
	
	// Create the AI system
	fla := NewFormalLanguageAI()
	
	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
	fmt.Println("\nReceived shutdown signal...")
		fla.Stop()
		os.Exit(0)
	}()
	
	// Start the system
	if err := fla.Start(); err != nil {
		log.Fatalf("Failed to start AI system: %v", err)
	}
	
	// Show initial status
	time.Sleep(3 * time.Second)
	fla.ShowStatus()
	
	// Run interactive CLI
	fla.RunInteractiveCLI()
}