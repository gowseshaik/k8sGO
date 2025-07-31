package server

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Server struct {
	version  string
	server   *http.Server
	upgrader websocket.Upgrader
}

type Resource struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Ready     string `json:"ready"`
	Age       string `json:"age"`
	CPU       string `json:"cpu"`
	Memory    string `json:"memory"`
}

type ApiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
}

func NewServer(version string) *Server {
	return &Server{
		version: version,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for local use
			},
		},
	}
}

func (s *Server) Start(port string) error {
	router := mux.NewRouter()

	// Static routes
	router.HandleFunc("/", s.handleHome).Methods("GET")
	router.HandleFunc("/api/contexts", s.handleGetContexts).Methods("GET")
	router.HandleFunc("/api/namespaces", s.handleGetNamespaces).Methods("GET")
	router.HandleFunc("/api/resources/{type}", s.handleGetResources).Methods("GET")
	router.HandleFunc("/api/describe/{type}/{name}", s.handleDescribeResource).Methods("GET")
	router.HandleFunc("/api/yaml/{type}/{name}", s.handleGetYaml).Methods("GET")
	router.HandleFunc("/api/mermaid/{type}/{name}", s.handleGetMermaid).Methods("GET")
	router.HandleFunc("/api/top", s.handleTopCommand).Methods("GET")
	router.HandleFunc("/api/switch-context", s.handleSwitchContext).Methods("POST")
	router.HandleFunc("/api/switch-namespace", s.handleSwitchNamespace).Methods("POST")
	
	// WebSocket for real-time updates
	router.HandleFunc("/ws", s.handleWebSocket)

	s.server = &http.Server{
		Addr:    port,
		Handler: router,
	}

	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>K8sGO Web GUI</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; 
            background: #f5f5f5; 
            color: #333; 
        }
        .header { 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); 
            color: white; 
            padding: 1rem; 
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .header h1 { display: inline-block; margin-right: 2rem; }
        .context-selector { 
            display: inline-block; 
            margin-left: 2rem; 
        }
        .context-selector select { 
            padding: 0.5rem; 
            border: none; 
            border-radius: 4px; 
            background: white;
        }
        .main-container { 
            display: flex; 
            height: calc(100vh - 80px); 
        }
        .sidebar { 
            width: 300px; 
            background: white; 
            border-right: 1px solid #ddd; 
            overflow-y: auto;
        }
        .content { 
            flex: 1; 
            background: white; 
            margin: 1rem; 
            border-radius: 8px; 
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .tab-container { 
            display: flex; 
            background: #f8f9fa; 
            border-bottom: 1px solid #ddd; 
        }
        .tab { 
            padding: 1rem 2rem; 
            cursor: pointer; 
            border-bottom: 3px solid transparent; 
            transition: all 0.3s ease;
        }
        .tab:hover { background: #e9ecef; }
        .tab.active { 
            background: white; 
            border-bottom-color: #667eea; 
            color: #667eea; 
        }
        .tab-content { 
            padding: 1rem; 
            height: calc(100% - 60px); 
            overflow-y: auto; 
        }
        .resource-list { 
            list-style: none; 
        }
        .resource-item { 
            padding: 0.75rem; 
            border-bottom: 1px solid #eee; 
            cursor: pointer; 
            transition: background 0.2s ease;
        }
        .resource-item:hover { background: #f8f9fa; }
        .resource-item.selected { 
            background: #e3f2fd; 
            border-left: 4px solid #2196f3; 
        }
        .output-area { 
            background: #1e1e1e; 
            color: #d4d4d4; 
            font-family: 'Consolas', 'Monaco', monospace; 
            padding: 1rem; 
            height: 100%; 
            overflow-y: auto; 
            white-space: pre-wrap; 
            user-select: text; 
            cursor: text;
        }
        .toolbar { 
            background: #f8f9fa; 
            padding: 0.5rem 1rem; 
            border-bottom: 1px solid #ddd; 
            display: flex; 
            gap: 1rem; 
            align-items: center;
        }
        .btn { 
            padding: 0.5rem 1rem; 
            background: #667eea; 
            color: white; 
            border: none; 
            border-radius: 4px; 
            cursor: pointer; 
            transition: background 0.2s ease;
        }
        .btn:hover { background: #5a6fd8; }
        .btn-success { background: #28a745; }
        .btn-success:hover { background: #218838; }
        .status-bar { 
            background: #343a40; 
            color: white; 
            padding: 0.5rem 1rem; 
            position: fixed; 
            bottom: 0; 
            left: 0; 
            right: 0; 
            font-size: 0.9rem;
        }
        .loading { opacity: 0.6; }
        .copy-success { 
            position: fixed; 
            top: 20px; 
            right: 20px; 
            background: #28a745; 
            color: white; 
            padding: 1rem; 
            border-radius: 4px; 
            box-shadow: 0 4px 12px rgba(0,0,0,0.3); 
            z-index: 1000; 
            animation: slideIn 0.3s ease;
        }
        @keyframes slideIn { 
            from { transform: translateX(100%); } 
            to { transform: translateX(0); } 
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🌐 K8sGO Web GUI</h1>
        <div class="context-selector">
            <label>Context:</label>
            <select id="contextSelect">
                <option>Loading...</option>
            </select>
        </div>
        <button class="btn" onclick="refreshData()">🔄 Refresh</button>
    </div>
    
    <div class="main-container">
        <div class="sidebar">
            <div class="tab-container" style="flex-direction: column;">
                <div class="tab active" onclick="switchTab('pods')">📦 Pods</div>
                <div class="tab" onclick="switchTab('services')">🌐 Services</div>
                <div class="tab" onclick="switchTab('deployments')">🚀 Deployments</div>
                <div class="tab" onclick="switchTab('top')">📊 Top</div>
            </div>
            <ul class="resource-list" id="resourceList">
                <li class="resource-item">Loading resources...</li>
            </ul>
        </div>
        
        <div class="content">
            <div class="toolbar">
                <span id="currentView">Pods</span>
                <button class="btn btn-success" onclick="copyContent()">📋 Copy All</button>
                <button class="btn" onclick="refreshContent()">🔄 Refresh</button>
            </div>
            <div class="tab-content">
                <div class="output-area" id="outputArea">
                    Select a resource to view details, or click "Top" tab to see cluster resource usage.
                    
                    🎯 Perfect Copy/Paste:
                    • Select any text with your mouse
                    • Copy with Ctrl+C (or right-click → Copy)
                    • Paste anywhere with Ctrl+V
                    
                    This web interface solves all clipboard issues!
                </div>
            </div>
        </div>
    </div>
    
    <div class="status-bar">
        <span id="statusText">Ready • Tool: kubectl • Web GUI v{{.Version}}</span>
    </div>

    <script>
        let currentTab = 'pods';
        let currentResource = null;
        
        // Initialize
        document.addEventListener('DOMContentLoaded', function() {
            loadContexts();
            loadResources('pods');
        });
        
        function switchTab(tab) {
            currentTab = tab;
            
            // Update tab styles
            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            event.target.classList.add('active');
            
            document.getElementById('currentView').textContent = tab.charAt(0).toUpperCase() + tab.slice(1);
            
            if (tab === 'top') {
                loadTopCommand();
            } else {
                loadResources(tab);
            }
        }
        
        function loadContexts() {
            fetch('/api/contexts')
                .then(response => response.json())
                .then(data => {
                    const select = document.getElementById('contextSelect');
                    if (data.success) {
                        select.innerHTML = data.data.map(ctx => 
                            '<option value="' + ctx + '">' + ctx + '</option>'
                        ).join('');
                    } else {
                        select.innerHTML = '<option>Error loading contexts</option>';
                    }
                })
                .catch(error => {
                    console.error('Error loading contexts:', error);
                    setStatus('❌ Error loading contexts');
                });
        }
        
        function loadResources(type) {
            setStatus('Loading ' + type + '...');
            document.getElementById('resourceList').innerHTML = '<li class="resource-item">Loading...</li>';
            
            fetch('/api/resources/' + type)
                .then(response => response.json())
                .then(data => {
                    const list = document.getElementById('resourceList');
                    if (data.success && data.data) {
                        list.innerHTML = data.data.map(resource => 
                            '<li class="resource-item" onclick="selectResource(\'' + type + '\', \'' + resource.name + '\')">' +
                            '<strong>' + resource.name + '</strong><br>' +
                            '<small>Status: ' + resource.status + ' | Ready: ' + resource.ready + '</small>' +
                            '</li>'
                        ).join('');
                        setStatus('✅ Loaded ' + data.data.length + ' ' + type);
                    } else {
                        list.innerHTML = '<li class="resource-item">No ' + type + ' found</li>';
                        setStatus('No ' + type + ' found');
                    }
                })
                .catch(error => {
                    console.error('Error loading resources:', error);
                    setStatus('❌ Error loading ' + type);
                });
        }
        
        function selectResource(type, name) {
            // Update selected item
            document.querySelectorAll('.resource-item').forEach(item => item.classList.remove('selected'));
            event.target.classList.add('selected');
            
            currentResource = { type, name };
            setStatus('Loading details for ' + name + '...');
            
            fetch('/api/describe/' + type + '/' + name)
                .then(response => response.json())
                .then(data => {
                    const output = document.getElementById('outputArea');
                    if (data.success) {
                        output.textContent = data.data;
                        setStatus('✅ Details loaded for ' + name + ' - text is selectable and copyable!');
                    } else {
                        output.textContent = 'Error loading details: ' + data.error;
                        setStatus('❌ Error loading details');
                    }
                })
                .catch(error => {
                    console.error('Error loading resource details:', error);
                    setStatus('❌ Error loading resource details');
                });
        }
        
        function loadTopCommand() {
            setStatus('Loading top resources...');
            document.getElementById('outputArea').textContent = 'Loading kubectl top output...';
            
            fetch('/api/top')
                .then(response => response.json())
                .then(data => {
                    const output = document.getElementById('outputArea');
                    if (data.success) {
                        output.textContent = data.data;
                        setStatus('✅ Top resources loaded - select text and copy with Ctrl+C!');
                    } else {
                        output.textContent = 'Error loading top command: ' + data.error;
                        setStatus('❌ Top command failed - check if metrics-server is installed');
                    }
                })
                .catch(error => {
                    console.error('Error loading top command:', error);
                    setStatus('❌ Error loading top command');
                });
        }
        
        function copyContent() {
            const output = document.getElementById('outputArea');
            const text = output.textContent;
            
            navigator.clipboard.writeText(text).then(function() {
                showCopySuccess();
                setStatus('✅ Copied ' + text.length + ' characters - paste with Ctrl+V anywhere!');
            }).catch(function(err) {
                console.error('Copy failed:', err);
                setStatus('❌ Copy failed - try selecting text and using Ctrl+C');
            });
        }
        
        function showCopySuccess() {
            const div = document.createElement('div');
            div.className = 'copy-success';
            div.textContent = '✅ Copied to clipboard!';
            document.body.appendChild(div);
            
            setTimeout(() => {
                document.body.removeChild(div);
            }, 3000);
        }
        
        function refreshData() {
            if (currentTab === 'top') {
                loadTopCommand();
            } else {
                loadResources(currentTab);
            }
        }
        
        function refreshContent() {
            if (currentTab === 'top') {
                loadTopCommand();
            } else if (currentResource) {
                selectResource(currentResource.type, currentResource.name);
            }
        }
        
        function setStatus(message) {
            document.getElementById('statusText').textContent = message + ' • Web GUI v{{.Version}}';
        }
        
        // Context switching
        document.getElementById('contextSelect').addEventListener('change', function() {
            const context = this.value;
            setStatus('Switching to context: ' + context);
            
            fetch('/api/switch-context', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ context: context })
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    setStatus('✅ Switched to context: ' + context);
                    refreshData();
                } else {
                    setStatus('❌ Failed to switch context: ' + data.error);
                }
            });
        });
    </script>
</body>
</html>`

	tmpl, err := template.New("home").Parse(html)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Version string
	}{
		Version: s.version,
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, data)
}

func (s *Server) handleGetContexts(w http.ResponseWriter, r *http.Request) {
	// Get contexts using kubectl
	cmd := exec.Command("kubectl", "config", "get-contexts", "-o", "name")
	output, err := cmd.Output()
	if err != nil {
		s.sendJSON(w, ApiResponse{Success: false, Error: err.Error()})
		return
	}

	contexts := strings.Split(strings.TrimSpace(string(output)), "\n")
	s.sendJSON(w, ApiResponse{Success: true, Data: contexts})
}

func (s *Server) handleGetResources(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resourceType := vars["type"]

	var cmd *exec.Cmd
	switch resourceType {
	case "pods":
		cmd = exec.Command("kubectl", "get", "pods", "-o", "json")
	case "services":
		cmd = exec.Command("kubectl", "get", "services", "-o", "json")
	case "deployments":
		cmd = exec.Command("kubectl", "get", "deployments", "-o", "json")
	default:
		s.sendJSON(w, ApiResponse{Success: false, Error: "Unknown resource type"})
		return
	}

	_, err := cmd.Output()
	if err != nil {
		s.sendJSON(w, ApiResponse{Success: false, Error: err.Error()})
		return
	}

	// Parse and simplify the JSON response
	var resources []Resource
	// This is simplified - in production you'd parse the full kubectl JSON response
	// For now, let's use a simple kubectl command to get formatted output
	simpleCmd := exec.Command("kubectl", "get", resourceType, "--no-headers")
	simpleOutput, _ := simpleCmd.Output()
	
	lines := strings.Split(strings.TrimSpace(string(simpleOutput)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			resource := Resource{
				Name:   fields[0],
				Ready:  fields[1],
				Status: fields[2],
			}
			if len(fields) > 3 {
				resource.Age = fields[len(fields)-1]
			}
			resources = append(resources, resource)
		}
	}

	s.sendJSON(w, ApiResponse{Success: true, Data: resources})
}

func (s *Server) handleDescribeResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resourceType := vars["type"]
	name := vars["name"]

	cmd := exec.Command("kubectl", "describe", resourceType, name)
	output, err := cmd.Output()
	if err != nil {
		s.sendJSON(w, ApiResponse{Success: false, Error: err.Error()})
		return
	}

	s.sendJSON(w, ApiResponse{Success: true, Data: string(output)})
}

func (s *Server) handleTopCommand(w http.ResponseWriter, r *http.Request) {
	// Execute: kubectl top nodes && echo && kubectl top pods -A
	cmd := exec.Command("bash", "-c", "kubectl top nodes && echo && kubectl top pods -A")
	output, err := cmd.Output()
	if err != nil {
		errorMsg := fmt.Sprintf("Error running top command: %s\n\nNote: The 'kubectl top' command requires the metrics-server to be installed in your cluster.", err.Error())
		s.sendJSON(w, ApiResponse{Success: false, Error: errorMsg})
		return
	}

	s.sendJSON(w, ApiResponse{Success: true, Data: string(output)})
}

func (s *Server) handleSwitchContext(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Context string `json:"context"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendJSON(w, ApiResponse{Success: false, Error: err.Error()})
		return
	}

	cmd := exec.Command("kubectl", "config", "use-context", req.Context)
	err := cmd.Run()
	if err != nil {
		s.sendJSON(w, ApiResponse{Success: false, Error: err.Error()})
		return
	}

	s.sendJSON(w, ApiResponse{Success: true, Data: "Context switched successfully"})
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Keep connection alive and handle real-time updates
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		// Echo back for now - could implement real-time resource updates
	}
}

func (s *Server) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}