package ml

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/fla/self-programming-ai/pkg/fsm"
)

// TransitionProbability represents learned probabilities for state transitions
type TransitionProbability struct {
	From        fsm.State  `json:"from"`
	Event       fsm.Event  `json:"event"`
	To          fsm.State  `json:"to"`
	Probability float64    `json:"probability"`
	Count       int        `json:"count"`
	LastSeen    time.Time  `json:"last_seen"`
}

// StateMetrics holds performance metrics for each state
type StateMetrics struct {
	State              fsm.State     `json:"state"`
	VisitCount         int           `json:"visit_count"`
	AverageResidenceTime time.Duration `json:"average_residence_time"`
	ErrorRate          float64       `json:"error_rate"`
	SuccessfulExits    int           `json:"successful_exits"`
	TotalExits         int           `json:"total_exits"`
	LastVisit          time.Time     `json:"last_visit"`
}

// MLAssistant provides machine learning capabilities for FSM optimization
type MLAssistant struct {
	probabilities          map[string]*TransitionProbability
	stateMetrics          map[fsm.State]*StateMetrics
	transitions           []TransitionHistory
	contextPatterns       map[string]*ContextPattern
	performanceHistory    []PerformanceSnapshot
	learningRate          float64
	decayFactor           float64
	mu                    sync.RWMutex
	neuralNetwork         *NeuralNetwork
	adaptiveBehavior      *AdaptiveBehavior
}

// TransitionHistory records transition events for learning
type TransitionHistory struct {
	Timestamp     time.Time         `json:"timestamp"`
	FromState     fsm.State         `json:"from_state"`
	ToState       fsm.State         `json:"to_state"`
	Event         fsm.Event         `json:"event"`
	Success       bool              `json:"success"`
	ResidenceTime time.Duration     `json:"residence_time"`
	Context       map[string]interface{} `json:"context"`
}

// PredictionResult represents ML predictions for state transitions
type PredictionResult struct {
	RecommendedEvent fsm.Event `json:"recommended_event"`
	Confidence       float64   `json:"confidence"`
	AlternativeEvents []EventPrediction `json:"alternatives"`
	Reasoning        string    `json:"reasoning"`
}

// EventPrediction represents a predicted event with its probability
type EventPrediction struct {
	Event       fsm.Event `json:"event"`
	Probability float64   `json:"probability"`
	ExpectedOutcome string `json:"expected_outcome"`
}

// NewMLAssistant creates a new machine learning assistant
func NewMLAssistant() *MLAssistant {
	return &MLAssistant{
		probabilities:       make(map[string]*TransitionProbability),
		stateMetrics:       make(map[fsm.State]*StateMetrics),
		transitions:        make([]TransitionHistory, 0),
		contextPatterns:    make(map[string]*ContextPattern),
		performanceHistory: make([]PerformanceSnapshot, 0),
		learningRate:       0.1,
		decayFactor:        0.95,
		neuralNetwork:      NewNeuralNetwork(),
		adaptiveBehavior:   NewAdaptiveBehavior(),
	}
}

// AttachToMachine attaches the ML assistant to an FSM for learning
func (ml *MLAssistant) AttachToMachine(machine fsm.Machine) {
	// Add learning hooks
	machine.AddHook(fsm.AfterTransition, ml.createLearningHook())
	machine.AddHook(fsm.OnStateEnter, ml.createStateMetricsHook())
	machine.AddHook(fsm.OnTransitionError, ml.createErrorLearningHook())
}

// createLearningHook creates a hook that learns from successful transitions
func (ml *MLAssistant) createLearningHook() fsm.Hook {
	return func(result fsm.TransitionResult, context fsm.Context) {
		if result.Success {
			ml.recordTransition(TransitionHistory{
				Timestamp: result.Timestamp,
				FromState: result.FromState,
				ToState:   result.ToState,
				Event:     result.Event,
				Success:   true,
				Context:   context.GetAll(),
			})
		}
	}
}

// createStateMetricsHook creates a hook that tracks state metrics
func (ml *MLAssistant) createStateMetricsHook() fsm.Hook {
	return func(result fsm.TransitionResult, context fsm.Context) {
		ml.updateStateMetrics(result.ToState, true, 0) // Successful entry
	}
}

// createErrorLearningHook creates a hook that learns from errors
func (ml *MLAssistant) createErrorLearningHook() fsm.Hook {
	return func(result fsm.TransitionResult, context fsm.Context) {
		ml.recordTransition(TransitionHistory{
			Timestamp: result.Timestamp,
			FromState: result.FromState,
			ToState:   result.ToState,
			Event:     result.Event,
			Success:   false,
			Context:   context.GetAll(),
		})
		
		// Update error metrics
		ml.updateStateMetrics(result.FromState, false, 0)
	}
}

// recordTransition records a transition for learning
func (ml *MLAssistant) recordTransition(transition TransitionHistory) {
	ml.transitions = append(ml.transitions, transition)
	
	// Update transition probabilities
	key := ml.transitionKey(transition.FromState, transition.Event, transition.ToState)
	
	if prob, exists := ml.probabilities[key]; exists {
		// Update existing probability using exponential moving average
		prob.Count++
		prob.Probability = (1-ml.learningRate)*prob.Probability + ml.learningRate*1.0
		prob.LastSeen = transition.Timestamp
	} else {
		// Create new probability entry
		ml.probabilities[key] = &TransitionProbability{
			From:        transition.FromState,
			Event:       transition.Event,
			To:          transition.ToState,
			Probability: 1.0,
			Count:       1,
			LastSeen:    transition.Timestamp,
		}
	}
	
	// Decay old probabilities
	ml.decayProbabilities()
	
	// Keep only recent transitions (sliding window)
	if len(ml.transitions) > 1000 {
		ml.transitions = ml.transitions[len(ml.transitions)-1000:]
	}
}

// updateStateMetrics updates metrics for a specific state
func (ml *MLAssistant) updateStateMetrics(state fsm.State, success bool, residenceTime time.Duration) {
	metrics, exists := ml.stateMetrics[state]
	if !exists {
		metrics = &StateMetrics{
			State:     state,
			LastVisit: time.Now(),
		}
		ml.stateMetrics[state] = metrics
	}
	
	metrics.VisitCount++
	metrics.LastVisit = time.Now()
	
	if residenceTime > 0 {
		// Update average residence time
		if metrics.AverageResidenceTime == 0 {
			metrics.AverageResidenceTime = residenceTime
		} else {
			metrics.AverageResidenceTime = time.Duration(
				float64(metrics.AverageResidenceTime)*0.9 + float64(residenceTime)*0.1,
			)
		}
	}
	
	// Update success/error rates
	if success {
		metrics.SuccessfulExits++
	}
	metrics.TotalExits++
	
	if metrics.TotalExits > 0 {
		metrics.ErrorRate = 1.0 - (float64(metrics.SuccessfulExits) / float64(metrics.TotalExits))
	}
}

// PredictNextTransition predicts the best next transition from current state
func (ml *MLAssistant) PredictNextTransition(currentState fsm.State, availableEvents []fsm.Event, context fsm.Context) *PredictionResult {
	eventPredictions := make([]EventPrediction, 0)
	
	for _, event := range availableEvents {
		probability := ml.calculateEventProbability(currentState, event, context)
		outcome := ml.predictOutcome(currentState, event)
		
		eventPredictions = append(eventPredictions, EventPrediction{
			Event:           event,
			Probability:     probability,
			ExpectedOutcome: outcome,
		})
	}
	
	// Sort by probability (descending)
	sort.Slice(eventPredictions, func(i, j int) bool {
		return eventPredictions[i].Probability > eventPredictions[j].Probability
	})
	
	if len(eventPredictions) == 0 {
		return &PredictionResult{
			Confidence: 0.0,
			Reasoning:  "No available events to predict",
		}
	}
	
	bestEvent := eventPredictions[0]
	alternatives := eventPredictions[1:]
	if len(alternatives) > 3 {
		alternatives = alternatives[:3] // Top 3 alternatives
	}
	
	reasoning := ml.generateReasoning(currentState, bestEvent, context)
	
	return &PredictionResult{
		RecommendedEvent:  bestEvent.Event,
		Confidence:        bestEvent.Probability,
		AlternativeEvents: alternatives,
		Reasoning:         reasoning,
	}
}

// calculateEventProbability calculates probability of an event being successful
func (ml *MLAssistant) calculateEventProbability(state fsm.State, event fsm.Event, context fsm.Context) float64 {
	// Base probability from historical data
	baseProbability := 0.5 // Default probability
	
	// Look for historical patterns
	for _, prob := range ml.probabilities {
		if prob.From == state && prob.Event == event {
			baseProbability = prob.Probability
			break
		}
	}
	
	// Context-based adjustments
	contextBonus := ml.calculateContextBonus(state, event, context)
	
	// Time-based decay for old data
	timeDecay := ml.calculateTimeDecay(state, event)
	
	// State health factor
	stateHealth := ml.calculateStateHealth(state)
	
	// Combine factors
	finalProbability := baseProbability * (1 + contextBonus) * timeDecay * stateHealth
	
	// Clamp between 0 and 1
	if finalProbability > 1.0 {
		finalProbability = 1.0
	}
	if finalProbability < 0.0 {
		finalProbability = 0.0
	}
	
	return finalProbability
}

// calculateContextBonus calculates bonus probability based on context
func (ml *MLAssistant) calculateContextBonus(state fsm.State, event fsm.Event, context fsm.Context) float64 {
	bonus := 0.0
	contextData := context.GetAll()
	
	// Look for patterns in historical transitions with similar context
	for _, transition := range ml.transitions {
		if transition.FromState == state && transition.Event == event && transition.Success {
			similarity := ml.calculateContextSimilarity(contextData, transition.Context)
			bonus += similarity * 0.1 // Small bonus for similar contexts
		}
	}
	
	return math.Min(bonus, 0.5) // Cap bonus at 50%
}

// calculateContextSimilarity calculates similarity between contexts
func (ml *MLAssistant) calculateContextSimilarity(context1, context2 map[string]interface{}) float64 {
	if len(context1) == 0 && len(context2) == 0 {
		return 1.0
	}
	
	matches := 0
	total := 0
	
	// Compare common keys
	for key, value1 := range context1 {
		if value2, exists := context2[key]; exists {
			total++
			if fmt.Sprintf("%v", value1) == fmt.Sprintf("%v", value2) {
				matches++
			}
		}
	}
	
	if total == 0 {
		return 0.0
	}
	
	return float64(matches) / float64(total)
}

// calculateTimeDecay calculates decay factor based on data age
func (ml *MLAssistant) calculateTimeDecay(state fsm.State, event fsm.Event) float64 {
	// Find most recent occurrence
	var mostRecent time.Time
	for _, prob := range ml.probabilities {
		if prob.From == state && prob.Event == event {
			if prob.LastSeen.After(mostRecent) {
				mostRecent = prob.LastSeen
			}
		}
	}
	
	if mostRecent.IsZero() {
		return 0.8 // Default decay for unknown patterns
	}
	
	age := time.Since(mostRecent)
	decayRate := math.Exp(-age.Hours() / 24.0) // Decay over days
	
	return math.Max(decayRate, 0.1) // Minimum 10% relevance
}

// calculateStateHealth calculates overall health of a state
func (ml *MLAssistant) calculateStateHealth(state fsm.State) float64 {
	metrics, exists := ml.stateMetrics[state]
	if !exists {
		return 1.0 // Unknown states get neutral health
	}
	
	// Health based on error rate (inverted)
	healthScore := 1.0 - metrics.ErrorRate
	
	// Bonus for frequently used states
	if metrics.VisitCount > 10 {
		healthScore += 0.1
	}
	
	// Penalty for states not visited recently
	daysSinceVisit := time.Since(metrics.LastVisit).Hours() / 24.0
	if daysSinceVisit > 7 {
		healthScore -= 0.2
	}
	
	return math.Max(healthScore, 0.1)
}

// predictOutcome predicts the likely outcome of a transition
func (ml *MLAssistant) predictOutcome(state fsm.State, event fsm.Event) string {
	// Find the most common target state for this transition
	targetCounts := make(map[fsm.State]int)
	
	for _, prob := range ml.probabilities {
		if prob.From == state && prob.Event == event {
			targetCounts[prob.To] += prob.Count
		}
	}
	
	if len(targetCounts) == 0 {
		return "Unknown outcome"
	}
	
	// Find most frequent target
	var mostFrequentTarget fsm.State
	maxCount := 0
	
	for target, count := range targetCounts {
		if count > maxCount {
			maxCount = count
			mostFrequentTarget = target
		}
	}
	
	return fmt.Sprintf("Likely transition to '%s'", mostFrequentTarget)
}

// generateReasoning generates human-readable reasoning for predictions
func (ml *MLAssistant) generateReasoning(state fsm.State, prediction EventPrediction, context fsm.Context) string {
	reasoning := fmt.Sprintf("Recommended '%s' (%.1f%% confidence)", 
		prediction.Event, prediction.Probability*100)
	
	// Add context-based reasoning
	if prediction.Probability > 0.8 {
		reasoning += " - High confidence based on historical success patterns"
	} else if prediction.Probability > 0.6 {
		reasoning += " - Moderate confidence based on past performance"
	} else {
		reasoning += " - Low confidence, limited historical data"
	}
	
	// Add state-specific insights
	if metrics, exists := ml.stateMetrics[state]; exists {
		if metrics.ErrorRate < 0.1 {
			reasoning += ". State shows excellent stability"
		} else if metrics.ErrorRate > 0.3 {
			reasoning += ". Warning: State has elevated error rate"
		}
	}
	
	return reasoning
}

// OptimizeMachine suggests optimizations for an FSM based on learned patterns
func (ml *MLAssistant) OptimizeMachine(machine fsm.Machine) []OptimizationSuggestion {
	suggestions := make([]OptimizationSuggestion, 0)
	
	// Analyze state usage patterns
	suggestions = append(suggestions, ml.analyzeStateUsage()...)
	
	// Analyze transition patterns
	suggestions = append(suggestions, ml.analyzeTransitionPatterns()...)
	
	// Analyze error patterns
	suggestions = append(suggestions, ml.analyzeErrorPatterns()...)
	
	return suggestions
}

// OptimizationSuggestion represents a suggestion for FSM improvement
type OptimizationSuggestion struct {
	Type        string  `json:"type"`
	Priority    string  `json:"priority"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"`
	Confidence  float64 `json:"confidence"`
}

// analyzeStateUsage analyzes state usage patterns
func (ml *MLAssistant) analyzeStateUsage() []OptimizationSuggestion {
	suggestions := make([]OptimizationSuggestion, 0)
	
	for state, metrics := range ml.stateMetrics {
		// Detect underused states
		if metrics.VisitCount < 5 && time.Since(metrics.LastVisit).Hours() > 48 {
			suggestions = append(suggestions, OptimizationSuggestion{
				Type:        "State Usage",
				Priority:    "Low",
				Description: fmt.Sprintf("State '%s' is rarely used - consider removal or better integration", state),
				Impact:      "Simplification and maintenance reduction",
				Confidence:  0.7,
			})
		}
		
		// Detect high-error states
		if metrics.ErrorRate > 0.3 && metrics.VisitCount > 10 {
			suggestions = append(suggestions, OptimizationSuggestion{
				Type:        "State Reliability",
				Priority:    "High",
				Description: fmt.Sprintf("State '%s' has high error rate (%.1f%%) - review transitions and conditions", state, metrics.ErrorRate*100),
				Impact:      "Improved system reliability",
				Confidence:  0.9,
			})
		}
	}
	
	return suggestions
}

// analyzeTransitionPatterns analyzes transition patterns for optimization
func (ml *MLAssistant) analyzeTransitionPatterns() []OptimizationSuggestion {
	suggestions := make([]OptimizationSuggestion, 0)
	
	// Group transitions by source state
	stateTransitions := make(map[fsm.State][]TransitionProbability)
	for _, prob := range ml.probabilities {
		stateTransitions[prob.From] = append(stateTransitions[prob.From], *prob)
	}
	
	for state, transitions := range stateTransitions {
		// Detect states with too many outgoing transitions (complexity)
		if len(transitions) > 5 {
			suggestions = append(suggestions, OptimizationSuggestion{
				Type:        "Complexity",
				Priority:    "Medium",
				Description: fmt.Sprintf("State '%s' has many outgoing transitions (%d) - consider refactoring", state, len(transitions)),
				Impact:      "Reduced complexity and improved maintainability",
				Confidence:  0.6,
			})
		}
		
		// Detect dominant transition patterns
		if len(transitions) > 1 {
			sort.Slice(transitions, func(i, j int) bool {
				return transitions[i].Count > transitions[j].Count
			})
			
			if transitions[0].Count > transitions[1].Count*3 {
				suggestions = append(suggestions, OptimizationSuggestion{
					Type:        "Flow Optimization",
					Priority:    "Medium",
					Description: fmt.Sprintf("State '%s' has dominant transition to '%s' - consider direct optimization", state, transitions[0].To),
					Impact:      "Performance improvement through optimized flow",
					Confidence:  0.8,
				})
			}
		}
	}
	
	return suggestions
}

// analyzeErrorPatterns analyzes error patterns for improvement suggestions
func (ml *MLAssistant) analyzeErrorPatterns() []OptimizationSuggestion {
	suggestions := make([]OptimizationSuggestion, 0)
	
	errorCounts := make(map[string]int)
	
	for _, transition := range ml.transitions {
		if !transition.Success {
			key := fmt.Sprintf("%s->%s", transition.FromState, transition.Event)
			errorCounts[key]++
		}
	}
	
	for pattern, count := range errorCounts {
		if count > 5 {
			suggestions = append(suggestions, OptimizationSuggestion{
				Type:        "Error Prevention",
				Priority:    "High",
				Description: fmt.Sprintf("Frequent failures in pattern '%s' (%d occurrences) - add guard conditions or improve validation", pattern, count),
				Impact:      "Reduced error rate and improved reliability",
				Confidence:  0.85,
			})
		}
	}
	
	return suggestions
}

// decayProbabilities applies decay to old probability data
func (ml *MLAssistant) decayProbabilities() {
	for _, prob := range ml.probabilities {
		age := time.Since(prob.LastSeen).Hours()
		if age > 24 { // Decay data older than 24 hours
			decayFactor := math.Exp(-age / 168) // Decay over weeks
			prob.Probability *= decayFactor
		}
	}
}

// transitionKey creates a unique key for transition probability storage
func (ml *MLAssistant) transitionKey(from fsm.State, event fsm.Event, to fsm.State) string {
	return fmt.Sprintf("%s:%s:%s", from, event, to)
}

// GetLearningStats returns statistics about the learning process
func (ml *MLAssistant) GetLearningStats() map[string]interface{} {
	return map[string]interface{}{
		"total_transitions":      len(ml.transitions),
		"learned_probabilities":  len(ml.probabilities),
		"monitored_states":       len(ml.stateMetrics),
		"learning_rate":          ml.learningRate,
		"decay_factor":           ml.decayFactor,
	}
}

// ExportLearningData exports learned data for backup or analysis
func (ml *MLAssistant) ExportLearningData() ([]byte, error) {
	data := map[string]interface{}{
		"probabilities": ml.probabilities,
		"state_metrics": ml.stateMetrics,
		"transitions":   ml.transitions,
		"exported_at":   time.Now(),
	}
	
	return json.MarshalIndent(data, "", "  ")
}

// ImportLearningData imports previously exported learning data
func (ml *MLAssistant) ImportLearningData(data []byte) error {
	var importData map[string]interface{}
	if err := json.Unmarshal(data, &importData); err != nil {
		return err
	}
	
	// Import probabilities
	if probData, exists := importData["probabilities"]; exists {
		probBytes, _ := json.Marshal(probData)
		json.Unmarshal(probBytes, &ml.probabilities)
	}
	
	// Import state metrics
	if metricsData, exists := importData["state_metrics"]; exists {
		metricsBytes, _ := json.Marshal(metricsData)
		json.Unmarshal(metricsBytes, &ml.stateMetrics)
	}
	
	// Import transitions
	if transData, exists := importData["transitions"]; exists {
		transBytes, _ := json.Marshal(transData)
		json.Unmarshal(transBytes, &ml.transitions)
	}
	
	return nil
}

// ContextPattern represents learned patterns in context data
type ContextPattern struct {
	Pattern     map[string]interface{} `json:"pattern"`
	Frequency   int                   `json:"frequency"`
	Success     float64               `json:"success"`
	LastSeen    time.Time             `json:"last_seen"`
	Confidence  float64               `json:"confidence"`
}

// PerformanceSnapshot captures system performance at a point in time
type PerformanceSnapshot struct {
	Timestamp       time.Time         `json:"timestamp"`
	TransitionCount int               `json:"transition_count"`
	SuccessRate     float64           `json:"success_rate"`
	AverageLatency  time.Duration     `json:"average_latency"`
	StateCount      map[string]int    `json:"state_count"`
	EventCount      map[string]int    `json:"event_count"`
}

// NeuralNetwork provides basic neural network functionality for pattern recognition
type NeuralNetwork struct {
	weights      [][]float64
	biases       []float64
	inputSize    int
	hiddenSize   int
	outputSize   int
	learningRate float64
}

// NewNeuralNetwork creates a simple neural network
func NewNeuralNetwork() *NeuralNetwork {
	nn := &NeuralNetwork{
		inputSize:    10,
		hiddenSize:   20,
		outputSize:   5,  
		learningRate: 0.01,
	}
	
	// Initialize weights and biases
	nn.weights = make([][]float64, 2)
	nn.weights[0] = make([]float64, nn.inputSize*nn.hiddenSize)
	nn.weights[1] = make([]float64, nn.hiddenSize*nn.outputSize)
	nn.biases = make([]float64, nn.hiddenSize+nn.outputSize)
	
	// Random initialization
	for i := range nn.weights[0] {
		nn.weights[0][i] = (rand.Float64() - 0.5) * 2
	}
	for i := range nn.weights[1] {
		nn.weights[1][i] = (rand.Float64() - 0.5) * 2
	}
	for i := range nn.biases {
		nn.biases[i] = (rand.Float64() - 0.5) * 2
	}
	
	return nn
}

// Forward performs forward propagation
func (nn *NeuralNetwork) Forward(input []float64) []float64 {
	if len(input) != nn.inputSize {
		// Pad or truncate input to match expected size
		normalizedInput := make([]float64, nn.inputSize)
		for i := 0; i < nn.inputSize && i < len(input); i++ {
			normalizedInput[i] = input[i]
		}
		input = normalizedInput
	}
	
	// Hidden layer
	hidden := make([]float64, nn.hiddenSize)
	for i := 0; i < nn.hiddenSize; i++ {
		sum := nn.biases[i]
		for j := 0; j < nn.inputSize; j++ {
			sum += input[j] * nn.weights[0][i*nn.inputSize+j]
		}
		hidden[i] = sigmoid(sum)
	}
	
	// Output layer
	output := make([]float64, nn.outputSize)
	for i := 0; i < nn.outputSize; i++ {
		sum := nn.biases[nn.hiddenSize+i]
		for j := 0; j < nn.hiddenSize; j++ {
			sum += hidden[j] * nn.weights[1][i*nn.hiddenSize+j]
		}
		output[i] = sigmoid(sum)
	}
	
	return output
}

// Train performs basic training (simplified backpropagation)
func (nn *NeuralNetwork) Train(input []float64, target []float64) {
	output := nn.Forward(input)
	
	// Calculate error and update weights (simplified)
	for i := range output {
		if i < len(target) {
			error := target[i] - output[i]
			// Simple weight update (this is a very basic implementation)
			for j := range nn.weights[1] {
				if j < len(output) {
					nn.weights[1][j] += nn.learningRate * error * 0.1
				}
			}
		}
	}
}

// sigmoid activation function
func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

// AdaptiveBehavior manages dynamic behavior adaptation
type AdaptiveBehavior struct {
	behaviorRules    map[string]*BehaviorRule
	adaptationRate   float64
	performanceGoals map[string]float64
	mu               sync.RWMutex
}

// BehaviorRule defines adaptive behavior rules
type BehaviorRule struct {
	Name        string                 `json:"name"`
	Condition   map[string]interface{} `json:"condition"`
	Action      string                 `json:"action"`
	Priority    int                    `json:"priority"`
	Active      bool                   `json:"active"`
	Performance float64                `json:"performance"`
	LastUsed    time.Time              `json:"last_used"`
}

// NewAdaptiveBehavior creates adaptive behavior manager
func NewAdaptiveBehavior() *AdaptiveBehavior {
	return &AdaptiveBehavior{
		behaviorRules:    make(map[string]*BehaviorRule),
		adaptationRate:   0.1,
		performanceGoals: make(map[string]float64),
	}
}

// AddBehaviorRule adds a new behavior rule
func (ab *AdaptiveBehavior) AddBehaviorRule(rule *BehaviorRule) {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	
	ab.behaviorRules[rule.Name] = rule
}

// EvaluateRules evaluates behavior rules and returns recommendations
func (ab *AdaptiveBehavior) EvaluateRules(context map[string]interface{}) []string {
	ab.mu.RLock()
	defer ab.mu.RUnlock()
	
	var recommendations []string
	
	for _, rule := range ab.behaviorRules {
		if rule.Active && ab.matchesCondition(rule.Condition, context) {
			recommendations = append(recommendations, rule.Action)
			rule.LastUsed = time.Now()
		}
	}
	
	return recommendations
}

// matchesCondition checks if context matches rule condition
func (ab *AdaptiveBehavior) matchesCondition(condition map[string]interface{}, context map[string]interface{}) bool {
	for key, expectedValue := range condition {
		if contextValue, exists := context[key]; exists {
			if contextValue != expectedValue {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

// UpdatePerformance updates rule performance based on outcomes
func (ab *AdaptiveBehavior) UpdatePerformance(ruleName string, success bool) {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	
	if rule, exists := ab.behaviorRules[ruleName]; exists {
		if success {
			rule.Performance += ab.adaptationRate
		} else {
			rule.Performance -= ab.adaptationRate
		}
		
		// Keep performance in reasonable bounds
		if rule.Performance > 1.0 {
			rule.Performance = 1.0
		} else if rule.Performance < 0.0 {
			rule.Performance = 0.0
		}
		
		// Deactivate poorly performing rules
		rule.Active = rule.Performance > 0.3
	}
}

// GetTopPerformingRules returns the best performing behavior rules
func (ab *AdaptiveBehavior) GetTopPerformingRules() []*BehaviorRule {
	ab.mu.RLock()
	defer ab.mu.RUnlock()
	
	var rules []*BehaviorRule
	for _, rule := range ab.behaviorRules {
		if rule.Active {
			rules = append(rules, rule)
		}
	}
	
	// Sort by performance (simple bubble sort for small datasets)
	for i := 0; i < len(rules)-1; i++ {
		for j := 0; j < len(rules)-i-1; j++ {
			if rules[j].Performance < rules[j+1].Performance {
				rules[j], rules[j+1] = rules[j+1], rules[j]
			}
		}
	}
	
	// Return top 5 rules
	if len(rules) > 5 {
		rules = rules[:5]
	}
	
	return rules
}

// PredictOptimalTransition uses neural network to predict optimal transitions
func (ml *MLAssistant) PredictOptimalTransition(state fsm.State, availableEvents []fsm.Event, context fsm.Context) PredictionResult {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	
	if ml.neuralNetwork == nil {
		// Fallback to probabilistic prediction
		prediction := ml.PredictNextTransition(state, availableEvents, context)
		return PredictionResult{
			RecommendedEvent: prediction.RecommendedEvent,
			Confidence:       prediction.Confidence,
			Reasoning:        prediction.Reasoning,
		}
	}
	
	// Prepare input for neural network
	input := ml.prepareNeuralInput(state, availableEvents, context)
	
	// Get prediction from neural network
	output := ml.neuralNetwork.Forward(input)
	
	// Find best event based on neural network output
	bestEvent := availableEvents[0]
	bestScore := 0.0
	
	for i, event := range availableEvents {
		if i < len(output) && output[i] > bestScore {
			bestScore = output[i]
			bestEvent = event
		}
	}
	
	return PredictionResult{
		RecommendedEvent: bestEvent,
		Confidence:       bestScore,
		Reasoning:        "Neural network prediction",
	}
}

// prepareNeuralInput converts FSM state to neural network input
func (ml *MLAssistant) prepareNeuralInput(state fsm.State, events []fsm.Event, context fsm.Context) []float64 {
	input := make([]float64, 10) // Fixed size input
	
	// Encode state (simple hash-based encoding)
	stateHash := float64(len(string(state))) / 100.0
	input[0] = stateHash
	
	// Encode number of available events
	input[1] = float64(len(events)) / 10.0
	
	// Encode context features
	contextData := context.GetAll()
	input[2] = float64(len(contextData)) / 20.0
	
	// Add some random features based on context
	i := 3
	for _, value := range contextData {
		if i >= len(input) {
			break
		}
		
		// Convert value to float (simplified)
		if strVal, ok := value.(string); ok {
			input[i] = float64(len(strVal)) / 50.0
		} else if floatVal, ok := value.(float64); ok {
			input[i] = floatVal / 100.0 // Normalize
		} else if intVal, ok := value.(int); ok {
			input[i] = float64(intVal) / 100.0
		}
		
		i++
	}
	
	return input
}

// TrainFromHistory trains the neural network from historical data
func (ml *MLAssistant) TrainFromHistory() {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	
	if ml.neuralNetwork == nil {
		return
	}
	
	// Train on historical transitions
	for _, transition := range ml.transitions {
		if len(transition.Context) == 0 {
			continue
		}
		
		// Prepare input
		input := make([]float64, 10)
		input[0] = float64(len(string(transition.FromState))) / 100.0
		input[1] = 1.0 // Single event
		
		// Add context features
		i := 2
		for _, value := range transition.Context {
			if i >= len(input) {
				break
			}
			
			if floatVal, ok := value.(float64); ok {
				input[i] = floatVal / 100.0
			} else if intVal, ok := value.(int); ok {
				input[i] = float64(intVal) / 100.0
			}
			i++
		}
		
		// Prepare target (success/failure)
		target := make([]float64, 5)
		if transition.Success {
			target[0] = 1.0 // Success
		} else {
			target[1] = 1.0 // Failure
		}
		
		// Train
		ml.neuralNetwork.Train(input, target)
	}
}