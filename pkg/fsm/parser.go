package fsm

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// NaturalLanguageParser converts natural language descriptions to FSM configurations
type NaturalLanguageParser struct {
	statePatterns      map[string]*regexp.Regexp
	eventPatterns      map[string]*regexp.Regexp
	transitionPatterns map[string]*regexp.Regexp
}

// NewNaturalLanguageParser creates a new parser instance
func NewNaturalLanguageParser() *NaturalLanguageParser {
	return &NaturalLanguageParser{
		statePatterns: map[string]*regexp.Regexp{
			"state_list": regexp.MustCompile(`(?i)states?:?\s*([a-zA-Z0-9_,\s]+)`),
			"states":     regexp.MustCompile(`(?i)(?:in|at|state)\s+([a-zA-Z0-9_]+)`),
		},
		eventPatterns: map[string]*regexp.Regexp{
			"event_list": regexp.MustCompile(`(?i)events?:?\s*([a-zA-Z0-9_,\s]+)`),
			"events":     regexp.MustCompile(`(?i)(?:when|on|event)\s+([a-zA-Z0-9_]+)`),
		},
		transitionPatterns: map[string]*regexp.Regexp{
			"transition": regexp.MustCompile(`(?i)from\s+([a-zA-Z0-9_]+)\s+(?:to|→)\s+([a-zA-Z0-9_]+)\s+(?:when|on)\s+([a-zA-Z0-9_]+)`),
			"simple":     regexp.MustCompile(`(?i)([a-zA-Z0-9_]+)\s+(?:→|->|to)\s+([a-zA-Z0-9_]+)`),
		},
	}
}

// ParseDescription converts natural language to FSM configuration
func (nlp *NaturalLanguageParser) ParseDescription(description string) (*ConfigMachine, error) {
	config := &ConfigMachine{
		Name:        "parsed_fsm",
		Description: "Auto-generated from natural language",
		States:      []StateConfig{},
		Events:      []EventConfig{},
		Transitions: []TransitionConfig{},
		Context:     make(map[string]interface{}),
	}

	// Parse states
	states, err := nlp.extractStates(description)
	if err != nil {
		return nil, fmt.Errorf("failed to parse states: %w", err)
	}
	config.States = states

	// Parse events
	events, err := nlp.extractEvents(description)
	if err != nil {
		return nil, fmt.Errorf("failed to parse events: %w", err)
	}
	config.Events = events

	// Parse transitions
	transitions, err := nlp.extractTransitions(description)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transitions: %w", err)
	}
	config.Transitions = transitions

	// Set initial state (first state found)
	if len(config.States) > 0 {
		config.InitialState = config.States[0].Name
	}

	return config, nil
}

// extractStates finds all states mentioned in the description
func (nlp *NaturalLanguageParser) extractStates(description string) ([]StateConfig, error) {
	stateSet := make(map[string]bool)
	var states []StateConfig

	// Look for explicit state lists
	for _, pattern := range nlp.statePatterns {
		matches := pattern.FindAllStringSubmatch(description, -1)
		for _, match := range matches {
			if len(match) > 1 {
				stateNames := strings.Split(match[1], ",")
				for _, name := range stateNames {
					name = strings.TrimSpace(name)
					if name != "" && !stateSet[name] {
						stateSet[name] = true
						states = append(states, StateConfig{
							Name:        name,
							Description: fmt.Sprintf("State: %s", name),
						})
					}
				}
			}
		}
	}

	// If no explicit states found, try to infer from transitions
	if len(states) == 0 {
		return nlp.inferStatesFromTransitions(description)
	}

	return states, nil
}

// extractEvents finds all events mentioned in the description
func (nlp *NaturalLanguageParser) extractEvents(description string) ([]EventConfig, error) {
	eventSet := make(map[string]bool)
	var events []EventConfig

	// Look for explicit event lists
	for _, pattern := range nlp.eventPatterns {
		matches := pattern.FindAllStringSubmatch(description, -1)
		for _, match := range matches {
			if len(match) > 1 {
				eventNames := strings.Split(match[1], ",")
				for _, name := range eventNames {
					name = strings.TrimSpace(name)
					if name != "" && !eventSet[name] {
						eventSet[name] = true
						events = append(events, EventConfig{
							Name:        name,
							Description: fmt.Sprintf("Event: %s", name),
						})
					}
				}
			}
		}
	}

	// If no explicit events found, try to infer from transitions
	if len(events) == 0 {
		return nlp.inferEventsFromTransitions(description)
	}

	return events, nil
}

// extractTransitions finds all transitions mentioned in the description
func (nlp *NaturalLanguageParser) extractTransitions(description string) ([]TransitionConfig, error) {
	var transitions []TransitionConfig

	// Look for transition patterns
	for _, pattern := range nlp.transitionPatterns {
		matches := pattern.FindAllStringSubmatch(description, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				var from, to, event string

				if len(match) >= 4 { // from X to Y when Z
					from, to, event = match[1], match[2], match[3]
				} else { // X -> Y
					from, to = match[1], match[2]
					event = "trigger" // default event
				}

				transitions = append(transitions, TransitionConfig{
					From:  from,
					Event: event,
					To:    to,
				})
			}
		}
	}

	return transitions, nil
}

// inferStatesFromTransitions extracts states from transition descriptions
func (nlp *NaturalLanguageParser) inferStatesFromTransitions(description string) ([]StateConfig, error) {
	stateSet := make(map[string]bool)
	var states []StateConfig

	transitions, _ := nlp.extractTransitions(description)
	for _, transition := range transitions {
		if !stateSet[transition.From] {
			stateSet[transition.From] = true
			states = append(states, StateConfig{
				Name:        transition.From,
				Description: fmt.Sprintf("Inferred state: %s", transition.From),
			})
		}
		if !stateSet[transition.To] {
			stateSet[transition.To] = true
			states = append(states, StateConfig{
				Name:        transition.To,
				Description: fmt.Sprintf("Inferred state: %s", transition.To),
			})
		}
	}

	return states, nil
}

// inferEventsFromTransitions extracts events from transition descriptions
func (nlp *NaturalLanguageParser) inferEventsFromTransitions(description string) ([]EventConfig, error) {
	eventSet := make(map[string]bool)
	var events []EventConfig

	transitions, _ := nlp.extractTransitions(description)
	for _, transition := range transitions {
		if !eventSet[transition.Event] {
			eventSet[transition.Event] = true
			events = append(events, EventConfig{
				Name:        transition.Event,
				Description: fmt.Sprintf("Inferred event: %s", transition.Event),
			})
		}
	}

	return events, nil
}

// ParseJSON converts JSON-based FSM descriptions to configurations
func (nlp *NaturalLanguageParser) ParseJSON(jsonStr string) (*ConfigMachine, error) {
	var rawConfig map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawConfig); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Convert to structured configuration
	config := &ConfigMachine{
		Context: make(map[string]interface{}),
	}

	// Extract basic information
	if name, ok := rawConfig["name"].(string); ok {
		config.Name = name
	}
	if desc, ok := rawConfig["description"].(string); ok {
		config.Description = desc
	}
	if initial, ok := rawConfig["initial_state"].(string); ok {
		config.InitialState = initial
	}

	// Extract states
	if statesData, ok := rawConfig["states"].([]interface{}); ok {
		for _, stateData := range statesData {
			if stateMap, ok := stateData.(map[string]interface{}); ok {
				state := StateConfig{}
				if name, ok := stateMap["name"].(string); ok {
					state.Name = name
				}
				if desc, ok := stateMap["description"].(string); ok {
					state.Description = desc
				}
				config.States = append(config.States, state)
			}
		}
	}

	// Extract events
	if eventsData, ok := rawConfig["events"].([]interface{}); ok {
		for _, eventData := range eventsData {
			if eventMap, ok := eventData.(map[string]interface{}); ok {
				event := EventConfig{}
				if name, ok := eventMap["name"].(string); ok {
					event.Name = name
				}
				if desc, ok := eventMap["description"].(string); ok {
					event.Description = desc
				}
				config.Events = append(config.Events, event)
			}
		}
	}

	// Extract transitions
	if transitionsData, ok := rawConfig["transitions"].([]interface{}); ok {
		for _, transitionData := range transitionsData {
			if transitionMap, ok := transitionData.(map[string]interface{}); ok {
				transition := TransitionConfig{}
				if from, ok := transitionMap["from"].(string); ok {
					transition.From = from
				}
				if event, ok := transitionMap["event"].(string); ok {
					transition.Event = event
				}
				if to, ok := transitionMap["to"].(string); ok {
					transition.To = to
				}
				if action, ok := transitionMap["action"].(string); ok {
					transition.Action = action
				}
				config.Transitions = append(config.Transitions, transition)
			}
		}
	}

	return config, nil
}

// GenerateExample creates example FSMs for testing
func (nlp *NaturalLanguageParser) GenerateExample(exampleType string) (*ConfigMachine, error) {
	switch exampleType {
	case "traffic_light":
		config, err := nlp.ParseDescription(`
			States: red, yellow, green
			Events: timer, emergency_override
			From red to green when timer
			From green to yellow when timer
			From yellow to red when timer
		`)
		return config, err

	case "order_processing":
		config, err := nlp.ParseDescription(`
			States: pending, processing, shipped, delivered, cancelled
			Events: process, ship, deliver, cancel
			From pending to processing when process
			From processing to shipped when ship
			From shipped to delivered when deliver
			From pending to cancelled when cancel
			From processing to cancelled when cancel
		`)
		return config, err

	case "user_session":
		config, err := nlp.ParseDescription(`
			States: anonymous, authenticated, premium, locked
			Events: login, logout, upgrade, lock, unlock
			From anonymous to authenticated when login
			From authenticated to anonymous when logout
			From authenticated to premium when upgrade
			From authenticated to locked when lock
			From locked to authenticated when unlock
		`)
		return config, err

	default:
		return nil, fmt.Errorf("unknown example type: %s", exampleType)
	}
}
