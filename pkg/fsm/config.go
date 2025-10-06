package fsm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// ConfigMachine represents a machine configuration that can be loaded from files
type ConfigMachine struct {
	Name         string                    `json:"name" yaml:"name"`
	Description  string                    `json:"description" yaml:"description"`
	InitialState string                    `json:"initial_state" yaml:"initial_state"`
	States       []StateConfig             `json:"states" yaml:"states"`
	Events       []EventConfig             `json:"events" yaml:"events"`
	Transitions  []TransitionConfig        `json:"transitions" yaml:"transitions"`
	Context      map[string]interface{}    `json:"context" yaml:"context"`
	Hooks        map[string][]HookConfig   `json:"hooks" yaml:"hooks"`
}

// StateConfig represents a state configuration
type StateConfig struct {
	Name        string      `json:"name" yaml:"name"`
	Description string      `json:"description" yaml:"description"`
	Properties  interface{} `json:"properties" yaml:"properties"`
}

// EventConfig represents an event configuration
type EventConfig struct {
	Name        string      `json:"name" yaml:"name"`
	Description string      `json:"description" yaml:"description"`
	Properties  interface{} `json:"properties" yaml:"properties"`
}

// TransitionConfig represents a transition configuration
type TransitionConfig struct {
	From        string            `json:"from" yaml:"from"`
	Event       string            `json:"event" yaml:"event"`
	To          string            `json:"to" yaml:"to"`
	Condition   string            `json:"condition" yaml:"condition"`
	Action      string            `json:"action" yaml:"action"`
	Properties  map[string]string `json:"properties" yaml:"properties"`
}

// HookConfig represents a hook configuration
type HookConfig struct {
	Type        string            `json:"type" yaml:"type"`
	Action      string            `json:"action" yaml:"action"`
	Properties  map[string]string `json:"properties" yaml:"properties"`
}

// ConditionRegistry holds registered condition functions
type ConditionRegistry map[string]func(props map[string]string) TransitionCondition

// ActionRegistry holds registered action functions  
type ActionRegistry map[string]func(props map[string]string) TransitionAction

// HookRegistry holds registered hook functions
type HookRegistry map[string]func(props map[string]string) Hook

// ConfigLoader handles loading and building FSMs from configuration
type ConfigLoader struct {
	conditions ConditionRegistry
	actions    ActionRegistry
	hooks      HookRegistry
}

// NewConfigLoader creates a new configuration loader with default registries
func NewConfigLoader() *ConfigLoader {
	loader := &ConfigLoader{
		conditions: make(ConditionRegistry),
		actions:    make(ActionRegistry),
		hooks:      make(HookRegistry),
	}
	
	// Register default conditions
	loader.RegisterCondition("always_true", func(props map[string]string) TransitionCondition {
		return AlwaysTrue()
	})
	
	loader.RegisterCondition("always_false", func(props map[string]string) TransitionCondition {
		return AlwaysFalse()
	})
	
	loader.RegisterCondition("context_has_key", func(props map[string]string) TransitionCondition {
		key := props["key"]
		return ContextHasKey(key)
	})
	
	loader.RegisterCondition("context_equals", func(props map[string]string) TransitionCondition {
		key := props["key"]
		value := props["value"]
		return ContextEquals(key, value)
	})
	
	loader.RegisterCondition("context_greater_than", func(props map[string]string) TransitionCondition {
		key := props["key"]
		thresholdStr := props["threshold"]
		threshold, _ := strconv.ParseFloat(thresholdStr, 64)
		return ContextGreaterThan(key, threshold)
	})
	
	// Register default actions
	loader.RegisterAction("log", func(props map[string]string) TransitionAction {
		message := props["message"]
		return LogTransition(func(msg string) {
			fmt.Printf("[LOG] %s: %s\n", message, msg)
		})
	})
	
	loader.RegisterAction("set_context", func(props map[string]string) TransitionAction {
		key := props["key"]
		value := props["value"]
		return SetContextValue(key, value)
	})
	
	loader.RegisterAction("increment_counter", func(props map[string]string) TransitionAction {
		key := props["key"]
		return IncrementCounter(key)
	})
	
	// Register default hooks
	loader.RegisterHook("log_transition", func(props map[string]string) Hook {
		prefix := props["prefix"]
		if prefix == "" {
			prefix = "TRANSITION"
		}
		return func(result TransitionResult, context Context) {
			fmt.Printf("[%s] %s -> %s (Event: %s)\n", 
				prefix, result.FromState, result.ToState, result.Event)
		}
	})
	
	loader.RegisterHook("log_state_enter", func(props map[string]string) Hook {
		prefix := props["prefix"]
		if prefix == "" {
			prefix = "STATE_ENTER"
		}
		return func(result TransitionResult, context Context) {
			fmt.Printf("[%s] Entered state: %s\n", prefix, result.ToState)
		}
	})
	
	return loader
}

// RegisterCondition registers a condition function
func (cl *ConfigLoader) RegisterCondition(name string, factory func(props map[string]string) TransitionCondition) {
	cl.conditions[name] = factory
}

// RegisterAction registers an action function
func (cl *ConfigLoader) RegisterAction(name string, factory func(props map[string]string) TransitionAction) {
	cl.actions[name] = factory
}

// RegisterHook registers a hook function
func (cl *ConfigLoader) RegisterHook(name string, factory func(props map[string]string) Hook) {
	cl.hooks[name] = factory
}

// LoadFromJSON loads an FSM configuration from a JSON file
func (cl *ConfigLoader) LoadFromJSON(filename string) (*ConfigMachine, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %w", err)
	}
	
	var config ConfigMachine
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	return &config, nil
}

// LoadFromYAML loads an FSM configuration from a YAML file
func (cl *ConfigLoader) LoadFromYAML(filename string) (*ConfigMachine, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}
	
	var config ConfigMachine
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	
	return &config, nil
}

// BuildMachine builds an FSM from a configuration
func (cl *ConfigLoader) BuildMachine(config *ConfigMachine) (Machine, error) {
	builder := NewBuilderWithHooks()
	
	// Add states
	for _, stateConfig := range config.States {
		builder.AddState(State(stateConfig.Name))
	}
	
	// Add events
	for _, eventConfig := range config.Events {
		builder.AddEvent(Event(eventConfig.Name))
	}
	
	// Add transitions
	for _, transConfig := range config.Transitions {
		from := State(transConfig.From)
		event := Event(transConfig.Event)
		to := State(transConfig.To)
		
		// Handle condition
		var condition TransitionCondition
		if transConfig.Condition != "" {
			if conditionFactory, exists := cl.conditions[transConfig.Condition]; exists {
				condition = conditionFactory(transConfig.Properties)
			} else {
				return nil, fmt.Errorf("unknown condition: %s", transConfig.Condition)
			}
		}
		
		// Handle action
		var action TransitionAction
		if transConfig.Action != "" {
			if actionFactory, exists := cl.actions[transConfig.Action]; exists {
				action = actionFactory(transConfig.Properties)
			} else {
				return nil, fmt.Errorf("unknown action: %s", transConfig.Action)
			}
		}
		
		// Add transition based on what's configured
		if condition != nil && action != nil {
			builder.AddTransitionFull(from, event, to, condition, action)
		} else if condition != nil {
			builder.AddTransitionWithCondition(from, event, to, condition)
		} else if action != nil {
			builder.AddTransitionWithAction(from, event, to, action)
		} else {
			builder.AddTransition(from, event, to)
		}
	}
	
	// Add hooks
	for hookTypeStr, hookConfigs := range config.Hooks {
		hookType, err := cl.parseHookType(hookTypeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid hook type: %s", hookTypeStr)
		}
		
		for _, hookConfig := range hookConfigs {
			if hookFactory, exists := cl.hooks[hookConfig.Action]; exists {
				hook := hookFactory(hookConfig.Properties)
				switch hookType {
				case BeforeTransition:
					builder.AddBeforeTransitionHook(hook)
				case AfterTransition:
					builder.AddAfterTransitionHook(hook)
				case OnStateEnter:
					builder.AddOnStateEnterHook(hook)
				case OnStateExit:
					builder.AddOnStateExitHook(hook)
				case OnTransitionError:
					builder.AddOnTransitionErrorHook(hook)
				}
			} else {
				return nil, fmt.Errorf("unknown hook action: %s", hookConfig.Action)
			}
		}
	}
	
	// Set initial state
	if config.InitialState != "" {
		builder.SetInitialState(State(config.InitialState))
	}
	
	// Build the machine
	machine, err := builder.Build()
	if err != nil {
		return nil, err
	}
	
	// Set initial context
	if config.Context != nil {
		context := machine.GetContext()
		for key, value := range config.Context {
			context.Set(key, value)
		}
	}
	
	return machine, nil
}

// parseHookType converts string to HookType
func (cl *ConfigLoader) parseHookType(hookTypeStr string) (HookType, error) {
	switch strings.ToLower(hookTypeStr) {
	case "before_transition":
		return BeforeTransition, nil
	case "after_transition":
		return AfterTransition, nil
	case "on_state_enter":
		return OnStateEnter, nil
	case "on_state_exit":
		return OnStateExit, nil
	case "on_transition_error":
		return OnTransitionError, nil
	default:
		return 0, fmt.Errorf("unknown hook type: %s", hookTypeStr)
	}
}

// SaveToJSON saves a machine configuration to JSON
func (cl *ConfigLoader) SaveToJSON(config *ConfigMachine, filename string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}
	
	return nil
}

// SaveToYAML saves a machine configuration to YAML
func (cl *ConfigLoader) SaveToYAML(config *ConfigMachine, filename string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}
	
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}
	
	return nil
}

// ExtractConfig extracts configuration from an existing machine (reverse engineering)
func (cl *ConfigLoader) ExtractConfig(machine Machine, name, description string) *ConfigMachine {
	config := &ConfigMachine{
		Name:         name,
		Description:  description,
		InitialState: string(machine.CurrentState()),
		Context:      make(map[string]interface{}),
	}
	
	// Extract context
	contextData := machine.GetContext().GetAll()
	for key, value := range contextData {
		config.Context[key] = value
	}
	
	// Extract transitions (this would need additional machine introspection methods)
	transitions := machine.GetTransitions()
	for _, transition := range transitions {
		transConfig := TransitionConfig{
			From:  string(transition.From),
			Event: string(transition.Event),
			To:    string(transition.To),
		}
		
		// Note: We can't easily extract condition/action details without additional metadata
		// This would require the machine to store configuration metadata
		
		config.Transitions = append(config.Transitions, transConfig)
	}
	
	return config
}

// RuntimeReconfigurator allows dynamic reconfiguration of running machines
type RuntimeReconfigurator struct {
	loader   *ConfigLoader
	machines map[string]Machine
}

// NewRuntimeReconfigurator creates a new runtime reconfigurator
func NewRuntimeReconfigurator() *RuntimeReconfigurator {
	return &RuntimeReconfigurator{
		loader:   NewConfigLoader(),
		machines: make(map[string]Machine),
	}
}

// RegisterMachine registers a machine for runtime reconfiguration
func (rr *RuntimeReconfigurator) RegisterMachine(name string, machine Machine) {
	rr.machines[name] = machine
}

// ReconfigureFromFile reconfigures a machine from a configuration file
func (rr *RuntimeReconfigurator) ReconfigureFromFile(machineName, configFile string) error {
	_, exists := rr.machines[machineName]
	if !exists {
		return fmt.Errorf("machine not found: %s", machineName)
	}
	
	// Load new configuration
	var config *ConfigMachine
	var err error
	
	if strings.HasSuffix(configFile, ".json") {
		config, err = rr.loader.LoadFromJSON(configFile)
	} else if strings.HasSuffix(configFile, ".yaml") || strings.HasSuffix(configFile, ".yml") {
		config, err = rr.loader.LoadFromYAML(configFile)
	} else {
		return fmt.Errorf("unsupported config file format: %s", configFile)
	}
	
	if err != nil {
		return err
	}
	
	// Build new machine
	newMachine, err := rr.loader.BuildMachine(config)
	if err != nil {
		return err
	}
	
	// Replace the machine (in a real implementation, this might involve more sophisticated merging)
	rr.machines[machineName] = newMachine
	
	return nil
}

// GetMachine retrieves a registered machine
func (rr *RuntimeReconfigurator) GetMachine(name string) (Machine, bool) {
	machine, exists := rr.machines[name]
	return machine, exists
}