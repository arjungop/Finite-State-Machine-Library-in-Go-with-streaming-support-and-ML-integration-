package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fla/self-programming-ai/pkg/fsm"
)

// AdvancedVisualizationServer provides comprehensive FSM visualization and design tools
// Simplified version without ML dependencies for clean FSM management
type AdvancedVisualizationServer struct {
	port           int                            // HTTP server port number
	machines       map[string]fsm.Machine         // Collection of FSM instances by name
	history        map[string][]TransitionHistory // Transition history for each machine
	mu             sync.RWMutex                   // Thread-safe access to server state
	streamer       *fsm.EventStreamer             // Optional event streamer for distributed events
	designSessions map[string]*DesignSession      // Active FSM design sessions
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

// NewAdvancedVisualizationServer creates a simplified visualization server
// Returns a server instance without ML dependencies for lightweight operation
func NewAdvancedVisualizationServer(port int) *AdvancedVisualizationServer {
	return &AdvancedVisualizationServer{
		port:     port,                                 // Set HTTP server port
		machines: make(map[string]fsm.Machine),         // Initialize empty machine collection
		history:  make(map[string][]TransitionHistory), // Initialize empty history tracking
		// Initialize a default in-process event streamer with sane defaults
		streamer:       fsm.NewEventStreamer(fsm.StreamConfig{}),
		designSessions: make(map[string]*DesignSession), // Initialize empty design sessions
	}
}

// Start starts the advanced visualization server
func (avs *AdvancedVisualizationServer) Start() error {
	mux := http.NewServeMux()

	// Static assets
	mux.HandleFunc("/", avs.handleDashboard)
	mux.HandleFunc("/designer", avs.handleDesigner)
	mux.HandleFunc("/analyzer", avs.handleAnalyzer)

	// API endpoints - simplified without ML dependencies
	mux.HandleFunc("/api/machines", avs.handleMachinesAPI)              // Machine management API
	mux.HandleFunc("/api/machines/", avs.handleMachineAPI)              // Individual machine operations
	mux.HandleFunc("/api/design/sessions", avs.handleDesignSessionsAPI) // Design session management
	mux.HandleFunc("/api/design/sessions/", avs.handleDesignSessionAPI) // Individual session operations
	mux.HandleFunc("/api/metrics", avs.handleMetricsAPI)                // Basic performance metrics

	log.Printf("Simplified visualization server starting on port %d", avs.port) // Log server startup
	return http.ListenAndServe(fmt.Sprintf(":%d", avs.port), mux)               // Start HTTP server
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
                    const title = document.createElement('h3');
                    title.textContent = machine.name;
                    const state = document.createElement('div');
                    state.className = 'state';
                    state.textContent = 'State: ' + machine.current_state;
                    const eventsEl = document.createElement('div');
                    eventsEl.className = 'events';
                    eventsEl.textContent = 'Valid Events: ' + (machine.valid_events || []).join(', ');
                    const runningEl = document.createElement('div');
                    runningEl.textContent = 'Running: ' + (machine.is_running ? 'Yes' : 'No');
                    const updatedEl = document.createElement('div');
                    try { updatedEl.textContent = 'Last Update: ' + new Date(machine.last_update).toLocaleString(); } catch(e) { updatedEl.textContent = 'Last Update: -'; }

                    div.appendChild(title);
                    div.appendChild(state);
                    div.appendChild(eventsEl);
                    div.appendChild(runningEl);
                    div.appendChild(updatedEl);
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
                <button onclick="designManually()">Design Manually</button>
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
        
        function designManually() {
            // Parse natural language description
            const description = document.getElementById('nl-description').value;
            if (!description.trim()) {
                alert('Please enter a description of your FSM');
                return;
            }
            
            // Simple parser for the format shown in placeholder
            const lines = description.split('\n').map(l => l.trim()).filter(l => l);
            const states = [];
            const events = [];
            const transitions = [];
            
            lines.forEach(line => {
                if (line.toLowerCase().startsWith('states:')) {
                    const stateList = line.substring(7).split(',').map(s => s.trim());
                    stateList.forEach((state, i) => {
                        states.push({
                            name: state,
                            description: state,
                            x: 100 + (i % 3) * 200,
                            y: 100 + Math.floor(i / 3) * 150,
                            color: '#3498db',
                            is_initial: i === 0
                        });
                    });
                } else if (line.toLowerCase().startsWith('events:')) {
                    const eventList = line.substring(7).split(',').map(e => e.trim());
                    eventList.forEach(event => {
                        events.push({
                            name: event,
                            description: event,
                            color: '#e74c3c'
                        });
                    });
                } else if (line.toLowerCase().includes('to') && line.toLowerCase().includes('when')) {
                    // Parse "From X to Y when Z" format
                    const match = line.match(/from\s+(\w+)\s+to\s+(\w+)\s+when\s+(\w+)/i);
                    if (match) {
                        transitions.push({
                            from: match[1],
                            to: match[2],
                            event: match[3],
                            description: match[1] + " -> " + match[2] + " on " + match[3],
                            curved: false
                        });
                    }
                }
            });
            
            currentDesign = { states, events, transitions };
            visualizeDesign();
            updateDesignInfo();
        }
        
        function generateFromNL() {
            designManually(); // Use the same logic for now
        }
        
        function addState() {
            const name = prompt('Enter state name:');
            if (name && !currentDesign.states.find(s => s.name === name)) {
                currentDesign.states.push({
                    name: name,
                    description: name,
                    x: 100 + (currentDesign.states.length % 3) * 200,
                    y: 100 + Math.floor(currentDesign.states.length / 3) * 150,
                    color: '#3498db',
                    is_initial: currentDesign.states.length === 0
                });
                visualizeDesign();
                updateDesignInfo();
            }
        }
        
        function addEvent() {
            const name = prompt('Enter event name:');
            if (name && !currentDesign.events.find(e => e.name === name)) {
                currentDesign.events.push({
                    name: name,
                    description: name,
                    color: '#e74c3c'
                });
                updateDesignInfo();
            }
        }
        
        function addTransition() {
            if (currentDesign.states.length < 2) {
                alert('You need at least 2 states to create a transition');
                return;
            }
            if (currentDesign.events.length === 0) {
                alert('You need at least 1 event to create a transition');
                return;
            }
            
            const from = prompt('From state:', currentDesign.states[0].name);
            const to = prompt('To state:', currentDesign.states[1].name);
            const event = prompt('On event:', currentDesign.events[0].name);
            
            if (from && to && event) {
                const fromExists = currentDesign.states.find(s => s.name === from);
                const toExists = currentDesign.states.find(s => s.name === to);
                const eventExists = currentDesign.events.find(e => e.name === event);
                
                if (fromExists && toExists && eventExists) {
                    currentDesign.transitions.push({
                        from: from,
                        to: to,
                        event: event,
                        description: from + " -> " + to + " on " + event,
                        curved: false
                    });
                    visualizeDesign();
                    updateDesignInfo();
                } else {
                    alert('Invalid state or event name');
                }
            }
        }
        
        function deployFSM() {
            if (currentDesign.states.length === 0) {
                alert('Design is empty - nothing to deploy');
                return;
            }
            
            const name = document.getElementById('design-name').value || ('fsm_' + Date.now());
            
            // Convert design to FSM config format
            const config = {
                name: name,
                initial_state: currentDesign.states.find(s => s.is_initial)?.name || currentDesign.states[0].name,
                states: currentDesign.states.map(s => ({
                    name: s.name,
                    description: s.description
                })),
                events: currentDesign.events.map(e => ({
                    name: e.name,
                    description: e.description
                })),
                transitions: currentDesign.transitions.map(t => ({
                    from: t.from,
                    to: t.to,
                    event: t.event,
                    action: t.description
                }))
            };
            
            // Deploy via API
            fetch('/api/machines', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config)
            })
            .then(response => response.json())
            .then(result => {
                alert('FSM "' + name + '" deployed successfully!');
                console.log('Deployed FSM:', result);
            })
            .catch(error => {
                console.error('Deploy error:', error);
                alert('Deployment failed - check console for details');
            });
        }
        
        function loadDesign() {
            fetch('/api/design/sessions')
            .then(response => response.json())
            .then(sessions => {
                if (sessions.length === 0) {
                    alert('No saved designs found');
                    return;
                }
                
                let options = sessions.map((s, i) => i + ': ' + s.name).join('\n');
                const choice = prompt('Choose design to load:\n' + options);
                const index = parseInt(choice);
                
                if (index >= 0 && index < sessions.length) {
                    const session = sessions[index];
                    currentDesign = {
                        states: session.states || [],
                        events: session.events || [],
                        transitions: session.transitions || []
                    };
                    document.getElementById('design-name').value = session.name;
                    visualizeDesign();
                    updateDesignInfo();
                }
            })
            .catch(error => console.error('Load error:', error));
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

// handleDesignSessionsAPI manages design sessions
// Provides CRUD operations for FSM design sessions without ML dependencies
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

// handleMachinesAPI provides machine information and creates new machines
func (avs *AdvancedVisualizationServer) handleMachinesAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
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
		
	case "POST":
		// Create new FSM from design
		var config struct {
			Name         string `json:"name"`
			InitialState string `json:"initial_state"`
			States       []struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"states"`
			Events []struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"events"`
			Transitions []struct {
				From   string `json:"from"`
				To     string `json:"to"`
				Event  string `json:"event"`
				Action string `json:"action"`
			} `json:"transitions"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		// Create FSM using builder
		builder := fsm.NewBuilderWithHooks()
		
		// Add states
		for _, state := range config.States {
			builder.AddState(fsm.State(state.Name))
		}
		
		// Add events  
		for _, event := range config.Events {
			builder.AddEvent(fsm.Event(event.Name))
		}
		
		// Add transitions
		for _, transition := range config.Transitions {
			builder.AddTransitionWithAction(
				fsm.State(transition.From),
				fsm.Event(transition.Event),
				fsm.State(transition.To),
				func(from, to fsm.State, event fsm.Event, ctx fsm.Context) error {
					fmt.Printf("Executing transition: %s -> %s on %s\n", 
						transition.From, transition.To, transition.Event)
					return nil
				},
			)
		}
		
		// Set initial state and build machine
		builder.SetInitialState(fsm.State(config.InitialState))
		machine, err := builder.Build()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to build FSM: %v", err), http.StatusBadRequest)
			return
		}
		
		avs.RegisterMachine(config.Name, machine)
		
		// Return success response
		response := map[string]interface{}{
			"name": config.Name,
			"status": "deployed",
			"initial_state": config.InitialState,
			"states_count": len(config.States),
			"events_count": len(config.Events),
			"transitions_count": len(config.Transitions),
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
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
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>FSM Analytics Dashboard</title>
    <style>
        :root {
            --bg: #f5f7fb;
            --text: #1f2937;
            --muted: #6b7280;
            --primary: #2563eb;
            --card: #ffffff;
            --border: #e5e7eb;
            --success: #10b981;
            --error: #ef4444;
        }
        
        * { box-sizing: border-box; }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Arial, sans-serif;
            margin: 0;
            padding: 24px;
            background: var(--bg);
            color: var(--text);
        }
        
        .header {
            text-align: center;
            margin-bottom: 2rem;
        }
        
        .header h1 {
            margin: 0;
            font-size: 2.5rem;
            font-weight: 700;
            color: var(--text);
        }
        
        .header p {
            margin: 0.5rem 0 0;
            color: var(--muted);
            font-size: 1.125rem;
        }
        
        .nav {
            display: flex;
            gap: 1rem;
            justify-content: center;
            margin-bottom: 2rem;
        }
        
        .nav a {
            text-decoration: none;
            color: var(--muted);
            padding: 0.5rem 1rem;
            border-radius: 8px;
            transition: all 0.2s;
        }
        
        .nav a:hover {
            background: var(--card);
            color: var(--primary);
        }
        
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 1.5rem;
            margin-bottom: 2rem;
        }
        
        .card {
            background: var(--card);
            padding: 1.5rem;
            border-radius: 12px;
            box-shadow: 0 4px 16px rgba(0,0,0,.06);
            border: 1px solid var(--border);
        }
        
        .card h2 {
            margin: 0 0 1rem;
            font-size: 1.25rem;
            font-weight: 600;
            color: var(--text);
        }
        
        .metrics {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
            gap: 1rem;
        }
        
        .metric {
            text-align: center;
            padding: 1rem;
            background: var(--bg);
            border-radius: 8px;
        }
        
        .metric-value {
            font-size: 2rem;
            font-weight: 700;
            color: var(--primary);
            margin-bottom: 0.25rem;
        }
        
        .metric-label {
            font-size: 0.875rem;
            color: var(--muted);
            font-weight: 500;
        }
        
        .chart-container {
            height: 300px;
            margin-top: 1rem;
        }
        
        .machine-list {
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
        }
        
        .machine-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 0.75rem;
            background: var(--bg);
            border-radius: 8px;
            border: 1px solid var(--border);
        }
        
        .machine-name {
            font-weight: 600;
            color: var(--text);
        }
        
        .machine-state {
            font-size: 0.875rem;
            padding: 0.25rem 0.5rem;
            border-radius: 4px;
            background: var(--success);
            color: white;
        }
        
        .history-item {
            display: flex;
            justify-content: space-between;
            padding: 0.5rem;
            margin-bottom: 0.5rem;
            background: var(--bg);
            border-radius: 4px;
            font-size: 0.875rem;
        }
        
        .success { color: var(--success); }
        .error { color: var(--error); }
        
        button {
            background: var(--primary);
            color: white;
            border: none;
            padding: 0.5rem 1rem;
            border-radius: 6px;
            cursor: pointer;
            font-weight: 500;
            transition: background-color 0.2s;
        }
        
        button:hover {
            background: var(--primary-600);
        }
    </style>
    <script src="https://d3js.org/d3.v7.min.js"></script>
</head>
<body>
    <div class="header">
        <h1>📊 FSM Analytics Dashboard</h1>
        <p>Real-time performance metrics and insights</p>
    </div>
    
    <div class="nav">
        <a href="/">Dashboard</a>
        <a href="/designer">Visual Designer</a>
        <a href="/analyzer" style="color: var(--primary);">Performance Analyzer</a>
        <a href="/streaming">Event Streaming</a>
    </div>
    
    <div class="grid">
        <div class="card">
            <h2>📈 Performance Metrics</h2>
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
                    <div class="metric-value" id="total-transitions">0</div>
                    <div class="metric-label">Total Transitions</div>
                </div>
                <div class="metric">
                    <div class="metric-value" id="success-rate">100%</div>
                    <div class="metric-label">Success Rate</div>
                </div>
            </div>
        </div>
        
        <div class="card">
            <h2>🎯 Current Machines</h2>
            <div id="machine-list" class="machine-list">
                <p class="metric-label">Loading machines...</p>
            </div>
            <button onclick="refreshMachines()">Refresh</button>
        </div>
        
        <div class="card">
            <h2>📊 State Distribution</h2>
            <div id="state-chart" class="chart-container"></div>
        </div>
        
        <div class="card">
            <h2>🔄 Recent Transitions</h2>
            <div id="transition-history" style="max-height: 250px; overflow-y: auto;">
                <p class="metric-label">Loading transition history...</p>
            </div>
            <button onclick="refreshHistory()">Refresh History</button>
        </div>
    </div>

    <script>
        let machines = [];
        let transitionHistory = [];
        
        function loadAnalytics() {
            // Load machines
            fetch('/api/machines')
                .then(response => response.json())
                .then(data => {
                    machines = data;
                    updateMachineMetrics();
                    updateMachineList();
                    updateStateChart();
                })
                .catch(error => console.error('Error loading machines:', error));
            
            // Load transition history from all machines
            refreshHistory();
        }
        
        function updateMachineMetrics() {
            document.getElementById('total-machines').textContent = machines.length;
            document.getElementById('active-machines').textContent = machines.filter(m => m.is_running).length;
            
            // Calculate total transitions from history
            document.getElementById('total-transitions').textContent = transitionHistory.length;
            
            // Calculate success rate
            const successful = transitionHistory.filter(t => t.success).length;
            const rate = transitionHistory.length > 0 ? Math.round((successful / transitionHistory.length) * 100) : 100;
            document.getElementById('success-rate').textContent = rate + '%';
        }
        
        function updateMachineList() {
            const container = document.getElementById('machine-list');
            if (machines.length === 0) {
                container.innerHTML = '<p class="metric-label">No machines found</p>';
                return;
            }
            
            container.innerHTML = machines.map(machine => 
                '<div class="machine-item">' +
                    '<div>' +
                        '<div class="machine-name">' + machine.name + '</div>' +
                        '<div class="metric-label">Events: ' + machine.valid_events.join(', ') + '</div>' +
                    '</div>' +
                    '<div class="machine-state">' + machine.current_state + '</div>' +
                '</div>'
            ).join('');
        }
        
        function updateStateChart() {
            // Count states
            const stateCounts = {};
            machines.forEach(machine => {
                const state = machine.current_state;
                stateCounts[state] = (stateCounts[state] || 0) + 1;
            });
            
            // Simple bar chart using D3
            const container = d3.select('#state-chart');
            container.selectAll('*').remove();
            
            if (Object.keys(stateCounts).length === 0) {
                container.append('p').text('No data to display').style('color', '#6b7280');
                return;
            }
            
            const data = Object.entries(stateCounts);
            const maxCount = d3.max(data, d => d[1]);
            
            const svg = container.append('svg')
                .attr('width', '100%')
                .attr('height', '100%')
                .attr('viewBox', '0 0 400 250');
            
            const barHeight = 30;
            const spacing = 10;
            
            svg.selectAll('rect')
                .data(data)
                .enter()
                .append('rect')
                .attr('x', 80)
                .attr('y', (d, i) => 20 + i * (barHeight + spacing))
                .attr('width', d => (d[1] / maxCount) * 250)
                .attr('height', barHeight)
                .attr('fill', '#2563eb')
                .attr('rx', 4);
            
            svg.selectAll('.label')
                .data(data)
                .enter()
                .append('text')
                .attr('class', 'label')
                .attr('x', 75)
                .attr('y', (d, i) => 20 + i * (barHeight + spacing) + barHeight/2 + 5)
                .attr('text-anchor', 'end')
                .style('font-size', '12px')
                .style('fill', '#374151')
                .text(d => d[0]);
            
            svg.selectAll('.count')
                .data(data)
                .enter()
                .append('text')
                .attr('class', 'count')
                .attr('x', d => 85 + (d[1] / maxCount) * 250)
                .attr('y', (d, i) => 20 + i * (barHeight + spacing) + barHeight/2 + 5)
                .style('font-size', '12px')
                .style('fill', 'white')
                .text(d => d[1]);
        }
        
        function refreshHistory() {
            // Fetch history for all machines
            const historyPromises = machines.map(machine => 
                fetch('/api/machines/' + machine.name + '/history')
                    .then(response => response.ok ? response.json() : [])
                    .catch(() => [])
            );
            
            Promise.all(historyPromises)
                .then(histories => {
                    transitionHistory = histories.flat()
                        .sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp))
                        .slice(0, 20); // Keep only recent 20
                    
                    updateTransitionHistory();
                    updateMachineMetrics();
                })
                .catch(error => console.error('Error loading history:', error));
        }
        
        function updateTransitionHistory() {
            const container = document.getElementById('transition-history');
            if (transitionHistory.length === 0) {
                container.innerHTML = '<p class="metric-label">No transitions recorded yet</p>';
                return;
            }
            
            container.innerHTML = transitionHistory.map(t => {
                const time = new Date(t.timestamp).toLocaleTimeString();
                const statusClass = t.success ? 'success' : 'error';
                return '<div class="history-item">' +
                    '<span>' + t.from_state + ' → ' + t.to_state + ' (' + t.event + ')</span>' +
                    '<span class="' + statusClass + '">' + time + '</span>' +
                '</div>';
            }).join('');
        }
        
        function refreshMachines() {
            loadAnalytics();
        }
        
        // Load data on page load
        loadAnalytics();
        
        // Auto-refresh every 5 seconds
        setInterval(loadAnalytics, 5000);
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, _ := template.New("analyzer").Parse(tmpl)
	t.Execute(w, nil)
}

func (avs *AdvancedVisualizationServer) handleStreaming(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<!DOCTYPE html><html lang=\"en\"><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"><title>Event Streaming</title><style>body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Arial,sans-serif;margin:0;padding:24px;background:#f5f7fb;color:#1f2937}.card{background:#fff;padding:20px;border-radius:12px;box-shadow:0 4px 16px rgba(0,0,0,.06);border:1px solid #e5e7eb}</style></head><body><div class=\"card\"><h1>Event Streaming</h1><p>Real-time event streaming interface coming soon...</p></div></body></html>")
}

func (avs *AdvancedVisualizationServer) handleMachineAPI(w http.ResponseWriter, r *http.Request) {
	// Extract machine name from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	machineName := pathParts[3]
	
	// Check if this is a history request
	if len(pathParts) >= 5 && pathParts[4] == "history" {
		avs.handleMachineHistoryAPI(w, r, machineName)
		return
	}
	
	avs.mu.Lock()
	machine, exists := avs.machines[machineName]
	avs.mu.Unlock()
	
	if !exists {
		http.Error(w, "Machine not found", http.StatusNotFound)
		return
	}
	
	switch r.Method {
	case "GET":
		// Get machine status
		status := MachineStatus{
			Name:         machineName,
			CurrentState: string(machine.CurrentState()),
			IsRunning:    machine.IsRunning(),
			ValidEvents:  make([]string, 0),
			LastUpdate:   time.Now(),
		}
		
		for _, event := range machine.GetValidEvents() {
			status.ValidEvents = append(status.ValidEvents, string(event))
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
		
	case "POST":
		// Trigger event
		var request struct {
			Event string `json:"event"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		// Record transition attempt
		fromState := string(machine.CurrentState())
		
		// Send event
		result, err := machine.SendEvent(fsm.Event(request.Event))
		
		// Record result
		toState := string(machine.CurrentState())
		success := err == nil
		
		// Use result if available
		if result != nil {
			toState = string(result.ToState)
		}
		
		avs.mu.Lock()
		avs.history[machineName] = append(avs.history[machineName], TransitionHistory{
			Timestamp:   time.Now(),
			FromState:   fromState,
			ToState:     toState,
			Event:       request.Event,
			Success:     success,
			Error:       func() string { if err != nil { return err.Error() }; return "" }(),
			ExecutionID: fmt.Sprintf("%s_%d", machineName, time.Now().UnixNano()),
		})
		avs.mu.Unlock()
		
		response := map[string]interface{}{
			"machine": machineName,
			"event": request.Event,
			"from_state": fromState,
			"to_state": toState,
			"success": success,
		}
		
		if err != nil {
			response["error"] = err.Error()
			w.WriteHeader(http.StatusBadRequest)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (avs *AdvancedVisualizationServer) handleMachineHistoryAPI(w http.ResponseWriter, r *http.Request, machineName string) {
	avs.mu.RLock()
	history, exists := avs.history[machineName]
	avs.mu.RUnlock()
	
	if !exists {
		// Return empty history if machine doesn't exist or has no history
		history = make([]TransitionHistory, 0)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
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
		EventFrequency:       make(map[string]int), // Initialize empty event frequency map
		LastUpdated:          time.Now(),           // Set current timestamp
	}

	w.Header().Set("Content-Type", "application/json") // Set JSON response header
	json.NewEncoder(w).Encode(metrics)                 // Send metrics as JSON response
}
