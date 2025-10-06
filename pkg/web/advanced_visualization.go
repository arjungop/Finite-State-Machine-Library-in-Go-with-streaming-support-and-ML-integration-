package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/fla/self-programming-ai/pkg/fsm"
)

// AdvancedVisualizationServer provides comprehensive FSM visualization and design tools
type AdvancedVisualizationServer struct {
	port           int
	machines       map[string]fsm.Machine
	loader         *fsm.ConfigLoader
	parser         *fsm.NaturalLanguageParser
	streamer       *fsm.EventStreamer
	history        map[string][]TransitionHistory
	mu             sync.RWMutex
	designSessions map[string]*DesignSession
}

// DesignSession represents an FSM design session
type DesignSession struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	States      []StateDesign          `json:"states"`
	Events      []EventDesign          `json:"events"`
	Transitions []TransitionDesign     `json:"transitions"`
	Context     map[string]interface{} `json:"context"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// StateDesign represents a state in the designer
type StateDesign struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Color       string  `json:"color"`
	IsInitial   bool    `json:"is_initial"`
	IsFinal     bool    `json:"is_final"`
}

// EventDesign represents an event in the designer
type EventDesign struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

// TransitionDesign represents a transition in the designer
type TransitionDesign struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Event       string `json:"event"`
	Condition   string `json:"condition"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Curved      bool   `json:"curved"`
}

// RealTimeMetrics tracks performance metrics
type RealTimeMetrics struct {
	TransitionsPerSecond float64        `json:"transitions_per_second"`
	AverageResponseTime  time.Duration  `json:"average_response_time"`
	ErrorRate            float64        `json:"error_rate"`
	StateDistribution    map[string]int `json:"state_distribution"`
	EventFrequency       map[string]int `json:"event_frequency"`
	LastUpdated          time.Time      `json:"last_updated"`
}

// NewAdvancedVisualizationServer creates an enhanced visualization server
func NewAdvancedVisualizationServer(port int) *AdvancedVisualizationServer {
	return &AdvancedVisualizationServer{
		port:           port,
		machines:       make(map[string]fsm.Machine),
		loader:         fsm.NewConfigLoader(),
		parser:         fsm.NewNaturalLanguageParser(),
		streamer:       fsm.NewEventStreamer(fsm.StreamConfig{}),
		history:        make(map[string][]TransitionHistory),
		designSessions: make(map[string]*DesignSession),
	}
}

// Start starts the advanced visualization server
func (avs *AdvancedVisualizationServer) Start() error {
	mux := http.NewServeMux()

	// Static assets
	mux.HandleFunc("/", avs.handleDashboard)
	mux.HandleFunc("/designer", avs.handleDesigner)
	mux.HandleFunc("/analyzer", avs.handleAnalyzer)
	mux.HandleFunc("/streaming", avs.handleStreaming)

	// API endpoints
	mux.HandleFunc("/api/machines", avs.handleMachinesAPI)
	mux.HandleFunc("/api/machines/", avs.handleMachineAPI)
	mux.HandleFunc("/api/design/sessions", avs.handleDesignSessionsAPI)
	mux.HandleFunc("/api/design/sessions/", avs.handleDesignSessionAPI)
	mux.HandleFunc("/api/parse/natural", avs.handleNaturalLanguageAPI)
	mux.HandleFunc("/api/metrics", avs.handleMetricsAPI)
	mux.HandleFunc("/api/streaming/events", avs.handleStreamingEventsAPI)

	log.Printf("Advanced visualization server starting on port %d", avs.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", avs.port), mux)
}

// handleDashboard serves the main dashboard
func (avs *AdvancedVisualizationServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>FSM Advanced Dashboard</title>
    <style>
        :root {
            --bg: #f5f7fb;
            --text: #1f2937;
            --muted: #6b7280;
            --primary: #2563eb;
            --primary-600: #1d4ed8;
            --card: #ffffff;
            --border: #e5e7eb;
            --ok: #10b981;
            --stop: #ef4444;
        }
        * { box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Helvetica Neue', Arial, 'Noto Sans', 'Apple Color Emoji', 'Segoe UI Emoji', sans-serif; margin: 0; padding: 24px; background: var(--bg); color: var(--text); }
        .header { background: #111827; color: white; padding: 24px; border-radius: 12px; margin-bottom: 20px; }
        .header h1 { margin: 0 0 8px 0; font-weight: 700; letter-spacing: 0.2px; }
        .header p { margin: 0; color: #d1d5db; }
        .nav { display: flex; flex-wrap: wrap; gap: 12px; margin: 18px 0 22px; }
        .nav a { background: var(--primary); color: white; padding: 10px 16px; text-decoration: none; border-radius: 8px; font-weight: 600; border: 1px solid rgba(255,255,255,0.1); }
        .nav a:hover { background: var(--primary-600); }
        .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
        @media (max-width: 960px) { .grid { grid-template-columns: 1fr; } }
        .card { background: var(--card); padding: 20px; border-radius: 12px; box-shadow: 0 4px 16px rgba(0,0,0,0.06); border: 1px solid var(--border); }
        .card h2 { margin-top: 0; }
        .machine { border: 1px solid var(--border); padding: 15px; margin: 10px 0; border-radius: 8px; background: linear-gradient(180deg, #ffffff, #fbfbfd); }
        .running { border-left: 4px solid #27ae60; }
        .stopped { border-left: 4px solid #e74c3c; }
        .state { font-weight: 700; color: #111827; }
        .events { color: var(--muted); font-size: 0.9em; }
        .metrics { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; }
        .metric { text-align: center; padding: 12px; background: #f9fafb; border-radius: 10px; border: 1px solid var(--border); }
        .metric-value { font-size: 2.2em; font-weight: 800; color: var(--primary); }
        .metric-label { color: var(--muted); font-size: 0.9em; }
        .btn { background: var(--primary); color: #fff; padding: 10px 14px; border: none; border-radius: 8px; cursor: pointer; font-weight: 600; }
        .btn.secondary { background: #111827; }
        .btn.ghost { background: transparent; color: var(--primary); border: 1px solid var(--primary); }
        .btn + .btn { margin-left: 8px; }
    </style>
    <script>
        function refreshData() {
            Promise.all([
                fetch('/api/machines').then(r => r.json()),
                fetch('/api/metrics').then(r => r.json()).catch(() => null)
            ])
            .then(([machines, metrics]) => {
                updateMachines(machines);
                updateMetrics(machines, metrics);
            })
            .catch(error => console.error('Error:', error));
        }
        
        function updateMachines(machines) {
            const container = document.getElementById('machines');
            container.innerHTML = '';
            machines.forEach(machine => {
                const div = document.createElement('div');
                div.className = 'machine ' + (machine.is_running ? 'running' : 'stopped');
                div.innerHTML = ` + "`" + `
                    <h3>${machine.name}</h3>
                    <div class="state">State: ${machine.current_state}</div>
                    <div class="events">Valid Events: ${machine.valid_events.join(', ')}</div>
                    <div>Running: ${machine.is_running ? 'Yes' : 'No'}</div>
                    <div>Last Update: ${new Date(machine.last_update).toLocaleString()}</div>
                ` + "`" + `;
                container.appendChild(div);
            });
        }
        function updateMetrics(machines, metrics){
            try {
                document.getElementById('total-machines').textContent = machines.length;
                document.getElementById('active-machines').textContent = machines.filter(m => m.is_running).length;
                const tps = metrics && metrics.transitions_per_second ? Number(metrics.transitions_per_second) : 0;
                document.getElementById('transitions-sec').textContent = (isFinite(tps) ? tps.toFixed(1) : '0');
            } catch (e) { /* ignore */ }
        }
        
        setInterval(refreshData, 2000);
        window.onload = refreshData;
    </script>
</head>
<body>
    <div class="header">
        <h1>FSM Advanced Dashboard</h1>
        <p>Real-time monitoring, design, and analysis tools</p>
    </div>
    
    <div class="nav">
        <a href="/">Dashboard</a>
        <a href="/designer">Visual Designer</a>
        <a href="/analyzer">Performance Analyzer</a>
        <a href="/streaming">Event Streaming</a>
    </div>
    
    <div class="grid">
        <div class="card">
            <h2>System Metrics</h2>
            <div class="metrics">
                <div class="metric">
                    <div class="metric-value" id="total-machines">0</div>
                    <div class="metric-label">Total Machines</div>
                </div>
                <div class="metric">
                    <div class="metric-value" id="active-machines">0</div>
                    <div class="metric-label">Active Machines</div>
                </div>
                <div class="metric">
                    <div class="metric-value" id="transitions-sec">0</div>
                    <div class="metric-label">Transitions/sec</div>
                </div>
            </div>
        </div>
        
        <div class="card">
            <h2>Quick Actions</h2>
            <button onclick="location.href='/designer'">Create New FSM</button>
            <button onclick="refreshData()">Refresh Data</button>
            <button onclick="location.href='/analyzer'">View Analytics</button>
        </div>
    </div>
    
    <div class="card">
        <h2>Active Machines</h2>
        <div id="machines"></div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, _ := template.New("dashboard").Parse(tmpl)
	t.Execute(w, nil)
}

// handleDesigner serves the visual FSM designer
func (avs *AdvancedVisualizationServer) handleDesigner(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>FSM Visual Designer</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 0; }
        .toolbar { background: #34495e; color: white; padding: 15px; display: flex; gap: 10px; align-items: center; }
        .toolbar button { background: #3498db; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer; }
        .toolbar button:hover { background: #2980b9; }
        .toolbar input, .toolbar textarea { padding: 8px; border: 1px solid #ddd; border-radius: 4px; }
        .main { display: flex; height: calc(100vh - 60px); }
        .sidebar { width: 300px; background: #ecf0f1; padding: 20px; border-right: 1px solid #bdc3c7; overflow-y: auto; }
        .canvas-container { flex: 1; position: relative; }
        .canvas { width: 100%; height: 100%; border: none; }
        .properties { background: white; padding: 15px; margin: 10px 0; border-radius: 4px; }
        .nl-input { width: 100%; height: 100px; margin: 10px 0; }
    </style>
    <script src="https://d3js.org/d3.v7.min.js"></script>
</head>
<body>
    <div class="toolbar">
        <button onclick="newDesign()">New</button>
        <button onclick="saveDesign()">Save</button>
        <button onclick="loadDesign()">Load</button>
        <button onclick="generateFromNL()">Generate from Text</button>
        <button onclick="deployFSM()">Deploy</button>
        <input type="text" id="design-name" placeholder="Design name" />
    </div>
    
    <div class="main">
        <div class="sidebar">
            <div class="properties">
                <h3>📝 Natural Language Input</h3>
                <textarea class="nl-input" id="nl-description" placeholder="Describe your FSM in natural language...
Example:
States: idle, working, done
Events: start, finish
From idle to working when start
From working to done when finish"></textarea>
                <button onclick="parseNaturalLanguage()">Parse & Visualize</button>
            </div>
            
            <div class="properties">
                <h3>🎨 Tools</h3>
                <button onclick="addState()">Add State</button>
                <button onclick="addEvent()">Add Event</button>
                <button onclick="addTransition()">Add Transition</button>
            </div>
            
            <div class="properties">
                <h3>Current Design</h3>
                <div id="design-info">
                    <div>States: <span id="state-count">0</span></div>
                    <div>Events: <span id="event-count">0</span></div>
                    <div>Transitions: <span id="transition-count">0</span></div>
                </div>
            </div>
        </div>
        
        <div class="canvas-container">
            <svg class="canvas" id="design-canvas"></svg>
        </div>
    </div>

    <script>
        let currentDesign = { states: [], events: [], transitions: [] };
        let svg = d3.select("#design-canvas");
        
        function parseNaturalLanguage() {
            const description = document.getElementById('nl-description').value;
            fetch('/api/parse/natural', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ description: description })
            })
            .then(response => response.json())
            .then(config => {
                currentDesign = convertConfigToDesign(config);
                visualizeDesign();
                updateDesignInfo();
            })
            .catch(error => console.error('Error:', error));
        }
        
        function convertConfigToDesign(config) {
            const design = { states: [], events: [], transitions: [] };
            
            config.states.forEach((state, i) => {
                design.states.push({
                    name: state.name,
                    description: state.description,
                    x: 100 + (i % 3) * 200,
                    y: 100 + Math.floor(i / 3) * 150,
                    color: '#3498db',
                    is_initial: state.name === config.initial_state
                });
            });
            
            config.events.forEach(event => {
                design.events.push({
                    name: event.name,
                    description: event.description,
                    color: '#e74c3c'
                });
            });
            
            config.transitions.forEach(transition => {
                design.transitions.push({
                    from: transition.from,
                    to: transition.to,
                    event: transition.event,
                    description: transition.action || '',
                    curved: false
                });
            });
            
            return design;
        }
        
        function visualizeDesign() {
            svg.selectAll("*").remove();
            
            // Draw transitions first (so they appear behind states)
            const transitions = svg.selectAll(".transition")
                .data(currentDesign.transitions)
                .enter().append("g")
                .attr("class", "transition");
            
            transitions.each(function(d) {
                const fromState = currentDesign.states.find(s => s.name === d.from);
                const toState = currentDesign.states.find(s => s.name === d.to);
                
                if (fromState && toState) {
                    d3.select(this).append("line")
                        .attr("x1", fromState.x + 30)
                        .attr("y1", fromState.y + 30)
                        .attr("x2", toState.x + 30)
                        .attr("y2", toState.y + 30)
                        .attr("stroke", "#7f8c8d")
                        .attr("stroke-width", 2)
                        .attr("marker-end", "url(#arrowhead)");
                    
                    // Add event label
                    d3.select(this).append("text")
                        .attr("x", (fromState.x + toState.x) / 2 + 30)
                        .attr("y", (fromState.y + toState.y) / 2 + 25)
                        .attr("text-anchor", "middle")
                        .attr("font-size", "12px")
                        .attr("fill", "#2c3e50")
                        .text(d.event);
                }
            });
            
            // Add arrowhead marker
            svg.append("defs").append("marker")
                .attr("id", "arrowhead")
                .attr("viewBox", "0 -5 10 10")
                .attr("refX", 8)
                .attr("refY", 0)
                .attr("markerWidth", 6)
                .attr("markerHeight", 6)
                .attr("orient", "auto")
                .append("path")
                .attr("d", "M0,-5L10,0L0,5")
                .attr("fill", "#7f8c8d");
            
            // Draw states
            const states = svg.selectAll(".state")
                .data(currentDesign.states)
                .enter().append("g")
                .attr("class", "state")
                .attr("transform", d => ` + "`" + `translate(${d.x}, ${d.y})` + "`" + `);
            
            states.append("circle")
                .attr("cx", 30)
                .attr("cy", 30)
                .attr("r", 25)
                .attr("fill", d => d.is_initial ? "#27ae60" : d.color)
                .attr("stroke", "#2c3e50")
                .attr("stroke-width", 2);
            
            states.append("text")
                .attr("x", 30)
                .attr("y", 35)
                .attr("text-anchor", "middle")
                .attr("font-size", "12px")
                .attr("fill", "white")
                .text(d => d.name);
        }
        
        function updateDesignInfo() {
            document.getElementById('state-count').textContent = currentDesign.states.length;
            document.getElementById('event-count').textContent = currentDesign.events.length;
            document.getElementById('transition-count').textContent = currentDesign.transitions.length;
        }
        
        function newDesign() {
            currentDesign = { states: [], events: [], transitions: [] };
            visualizeDesign();
            updateDesignInfo();
        }
        
        function saveDesign() {
            const name = document.getElementById('design-name').value || 'Untitled';
            fetch('/api/design/sessions', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    name: name,
                    description: 'Visual design session',
                    ...currentDesign
                })
            })
            .then(response => response.json())
            .then(result => alert('Design saved successfully!'))
            .catch(error => console.error('Error:', error));
        }
        
        // Initialize empty design
        visualizeDesign();
        updateDesignInfo();
    </script>

</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, _ := template.New("designer").Parse(tmpl)
	t.Execute(w, nil)
}

// handleNaturalLanguageAPI processes natural language FSM descriptions
func (avs *AdvancedVisualizationServer) handleNaturalLanguageAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	config, err := avs.parser.ParseDescription(request.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// handleDesignSessionsAPI manages design sessions
func (avs *AdvancedVisualizationServer) handleDesignSessionsAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		avs.mu.RLock()
		sessions := make([]*DesignSession, 0, len(avs.designSessions))
		for _, session := range avs.designSessions {
			sessions = append(sessions, session)
		}
		avs.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sessions)

	case "POST":
		var session DesignSession
		if err := json.NewDecoder(r.Body).Decode(&session); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		session.ID = fmt.Sprintf("session_%d", time.Now().UnixNano())
		session.CreatedAt = time.Now()
		session.UpdatedAt = time.Now()

		avs.mu.Lock()
		avs.designSessions[session.ID] = &session
		avs.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// RegisterMachine registers a machine for visualization
func (avs *AdvancedVisualizationServer) RegisterMachine(name string, machine fsm.Machine) {
	avs.mu.Lock()
	defer avs.mu.Unlock()

	avs.machines[name] = machine
	avs.history[name] = make([]TransitionHistory, 0)

	// Register with streamer
	avs.streamer.RegisterMachine(name, machine)
}

// handleMachinesAPI provides machine information
func (avs *AdvancedVisualizationServer) handleMachinesAPI(w http.ResponseWriter, r *http.Request) {
	avs.mu.RLock()
	defer avs.mu.RUnlock()

	var machines []MachineStatus
	for name, machine := range avs.machines {
		status := MachineStatus{
			Name:         name,
			CurrentState: string(machine.CurrentState()),
			IsRunning:    machine.IsRunning(),
			ValidEvents:  make([]string, 0),
			LastUpdate:   time.Now(),
		}

		// Get valid events
		for _, event := range machine.GetValidEvents() {
			status.ValidEvents = append(status.ValidEvents, string(event))
		}

		machines = append(machines, status)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(machines)
}

// TransitionHistory represents historical transition data
type TransitionHistory struct {
	Timestamp   time.Time `json:"timestamp"`
	FromState   string    `json:"from_state"`
	ToState     string    `json:"to_state"`
	Event       string    `json:"event"`
	Success     bool      `json:"success"`
	Error       string    `json:"error,omitempty"`
	ExecutionID string    `json:"execution_id"`
}

// MachineStatus represents machine status for API
type MachineStatus struct {
	Name         string    `json:"name"`
	CurrentState string    `json:"current_state"`
	IsRunning    bool      `json:"is_running"`
	ValidEvents  []string  `json:"valid_events"`
	LastUpdate   time.Time `json:"last_update"`
}

// Placeholder handlers for other endpoints
func (avs *AdvancedVisualizationServer) handleAnalyzer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<!DOCTYPE html><html lang=\"en\"><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"><title>Performance Analyzer</title><style>body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Arial,sans-serif;margin:0;padding:24px;background:#f5f7fb;color:#1f2937}.card{background:#fff;padding:20px;border-radius:12px;box-shadow:0 4px 16px rgba(0,0,0,.06);border:1px solid #e5e7eb}</style></head><body><div class=\"card\"><h1>Performance Analyzer</h1><p>Analytics and performance metrics coming soon...</p></div></body></html>")
}

func (avs *AdvancedVisualizationServer) handleStreaming(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<!DOCTYPE html><html lang=\"en\"><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"><title>Event Streaming</title><style>body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Arial,sans-serif;margin:0;padding:24px;background:#f5f7fb;color:#1f2937}.card{background:#fff;padding:20px;border-radius:12px;box-shadow:0 4px 16px rgba(0,0,0,.06);border:1px solid #e5e7eb}</style></head><body><div class=\"card\"><h1>Event Streaming</h1><p>Real-time event streaming interface coming soon...</p></div></body></html>")
}

func (avs *AdvancedVisualizationServer) handleMachineAPI(w http.ResponseWriter, r *http.Request) {
	// Handle individual machine operations
	fmt.Fprintf(w, "Machine API endpoint")
}

func (avs *AdvancedVisualizationServer) handleDesignSessionAPI(w http.ResponseWriter, r *http.Request) {
	// Handle individual design session operations
	fmt.Fprintf(w, "Design session API endpoint")
}

func (avs *AdvancedVisualizationServer) handleMetricsAPI(w http.ResponseWriter, r *http.Request) {
	// Return real-time metrics
	metrics := RealTimeMetrics{
		TransitionsPerSecond: 0.0,
		AverageResponseTime:  0,
		ErrorRate:            0.0,
		StateDistribution:    make(map[string]int),
		EventFrequency:       make(map[string]int),
		LastUpdated:          time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (avs *AdvancedVisualizationServer) handleStreamingEventsAPI(w http.ResponseWriter, r *http.Request) {
	// Handle streaming events API
	fmt.Fprintf(w, "Streaming events API endpoint")
}
