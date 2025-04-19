package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"main/db"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Deployment tracks basic deployment metadata
type Deployment struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Language  string `json:"language"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	Port      string `json:"port,omitempty"` // Store the port if running
}

// runningCmds will track the active "func run" processes by function name
var runningCmds = make(map[string]*exec.Cmd)

// To avoid race conditions on runningCmds, wrap in a mutex
var cmdMutex sync.Mutex

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

// WebSocket clients
var clients = make(map[*websocket.Conn]bool)
var clientsMutex sync.Mutex

// Broadcast message to all connected clients
func broadcastMessage(message interface{}) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	msg, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Printf("Error sending message to client: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// WebSocket handler
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	defer conn.Close()

	// Add client to the list
	clientsMutex.Lock()
	clients[conn] = true
	clientsMutex.Unlock()

	// Remove client when they disconnect
	defer func() {
		clientsMutex.Lock()
		delete(clients, conn)
		clientsMutex.Unlock()
	}()

	// Keep the connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Initialize database
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create a new mux
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/ws", wsHandler)
	mux.HandleFunc("/create/", createHandler)
	mux.HandleFunc("/upload/", uploadHandler)
	mux.HandleFunc("/build/", buildHandler)
	mux.HandleFunc("/start/", startHandler)
	mux.HandleFunc("/stop/", stopHandler)
	mux.HandleFunc("/deployments/", handleDeployments)

	// Wrap the mux with CORS middleware
	handler := corsMiddleware(mux)

	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// handleDeployments multiplexes /deployments/ and /deployments/<name>/
func handleDeployments(w http.ResponseWriter, r *http.Request) {
	// If exact path is "/deployments/", list all deployments.
	if r.URL.Path == "/deployments/" || r.URL.Path == "/deployments" {
		deploymentsHandler(w, r)
		return
	}
	// Otherwise, treat it as a request for a named deployment's details.
	deploymentDetailHandler(w, r)
}

// createHandler - creates a new deployment via CLI func create -l <language> <name>
func createHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received Request at:", r.URL.Path)
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	language := parts[2]
	if language == "" {
		http.Error(w, "Language is required", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Check if deployment already exists
	existingDeployment, err := db.GetDeployment(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking existing deployment: %v", err), http.StatusInternalServerError)
		return
	}
	if existingDeployment != nil {
		http.Error(w, "Function with that name already exists", http.StatusBadRequest)
		return
	}

	dataDir := "./data"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		err := os.Mkdir(dataDir, 0755)
		if err != nil {
			log.Printf("Error creating directory: %v", err)
			http.Error(w, fmt.Sprintf("Error creating directory: %s", err), http.StatusInternalServerError)
			return
		}
	}

	functionDir := filepath.Join(dataDir, name)
	if _, err := os.Stat(functionDir); !os.IsNotExist(err) {
		http.Error(w, "Function directory already exists", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("func", "create", "-l", language, name)
	cmd.Dir = dataDir
	output, err := cmd.CombinedOutput()
	log.Printf("Command Output: %s", output)

	if err != nil {
		log.Printf("Error executing command: %v", err)
		http.Error(w, fmt.Sprintf("Error creating function: %s\nOutput: %s", err, output), http.StatusInternalServerError)
		return
	}

	// Generate a new UUID for the deployment ID
	deploymentID := uuid.New().String()
	deployment := db.Deployment{
		ID:        deploymentID,
		Name:      name,
		Language:  language,
		Status:    "Stopped",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	if err := db.CreateDeployment(deployment); err != nil {
		http.Error(w, fmt.Sprintf("Error saving deployment: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Function created successfully: %s", output)
}

// uploadHandler - handles uploading code and package files
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/upload/")
	name = strings.TrimSuffix(name, "/")

	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	codeFile, _, err := r.FormFile("code")
	if err != nil {
		http.Error(w, "Error retrieving code file", http.StatusBadRequest)
		return
	}
	defer codeFile.Close()

	packageFile, _, err := r.FormFile("package")
	if err != nil {
		http.Error(w, "Error retrieving package file", http.StatusBadRequest)
		return
	}
	defer packageFile.Close()

	deployment, err := db.GetDeployment(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving deployment: %v", err), http.StatusInternalServerError)
		return
	}
	if deployment == nil {
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	codeFileName, packageFileName := getLanguageSpecificFiles(deployment.Language)
	if err := saveFile(codeFile, filepath.Join("data", name, codeFileName)); err != nil {
		http.Error(w, "Error saving code file", http.StatusInternalServerError)
		return
	}
	if err := saveFile(packageFile, filepath.Join("data", name, packageFileName)); err != nil {
		http.Error(w, "Error saving package file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Files uploaded successfully")
}

func saveFile(file io.Reader, path string) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	return err
}

// buildHandler - updates the status to "Building", and runs build/deploy in a separate goroutine.
// When build+deploy completes successfully, sets the status to "Built."
func buildHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/build/")
	name = strings.TrimSuffix(name, "/")

	// Find the deployment
	deployment, err := db.GetDeployment(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving deployment: %v", err), http.StatusInternalServerError)
		return
	}
	if deployment == nil {
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	// Set status to "Building"
	deployment.Status = "Building"
	if err := db.UpdateDeployment(*deployment); err != nil {
		http.Error(w, fmt.Sprintf("Error updating deployment status: %v", err), http.StatusInternalServerError)
		return
	}

	// Run build and deploy in a separate goroutine to avoid blocking the request
	go func(d *db.Deployment, fnName string) {
		// Step 1: Build
		buildCmd := exec.Command("func", "build", fnName, "--registry", "localhost:5000")
		buildCmd.Dir = filepath.Join("./data", fnName)
		buildOutput, err := buildCmd.CombinedOutput()
		if err != nil {
			log.Printf("[ERROR] Build for %s failed: %v\nOutput:\n%s", fnName, err, string(buildOutput))
			d.Status = "Failed"
			if err := db.UpdateDeployment(*d); err != nil {
				log.Printf("Error updating deployment status: %v", err)
			}
			// Broadcast status update
			broadcastMessage(map[string]interface{}{
				"type": "status_update",
				"data": d,
			})
			return
		}
		log.Printf("[INFO] Build output for %s:\n%s", fnName, string(buildOutput))

		// Update status to "Built" after successful build
		d.Status = "Built"
		if err := db.UpdateDeployment(*d); err != nil {
			log.Printf("Error updating deployment status: %v", err)
		}
		// Broadcast status update
		broadcastMessage(map[string]interface{}{
			"type": "status_update",
			"data": d,
		})
	}(deployment, name)

	// Immediately return a quick message
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "Build process started for '%s'. Status set to Building.\n", name)
}

// startHandler - starts the function using "func run" in a separate goroutine
func startHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/start/")
	name = strings.TrimSuffix(name, "/")

	// Find the deployment
	deployment, err := db.GetDeployment(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving deployment: %v", err), http.StatusInternalServerError)
		return
	}
	if deployment == nil {
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	// If it's already running, return early
	if deployment.Status == "Running" {
		http.Error(w, "Function is already started", http.StatusBadRequest)
		return
	}

	// Prepare the command: e.g. func run <name>
	cmd := exec.Command("func", "run", "--registry", "localhost:5000")
	cmd.Dir = filepath.Join("./data", name)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating stdout pipe: %v", err), http.StatusInternalServerError)
		return
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating stderr pipe: %v", err), http.StatusInternalServerError)
		return
	}

	// Create a channel to receive the port number
	portChan := make(chan string, 1)

	go func() {
		reader := io.MultiReader(stdoutPipe, stderrPipe)
		buf := make([]byte, 1024)
		for {
			n, err := reader.Read(buf)
			if n > 0 {
				outputChunk := string(buf[:n])
				log.Printf("[func run output for %s]: %s", name, outputChunk)

				// Attempt to detect the port from the output
				if strings.Contains(outputChunk, "Running on host port ") {
					// Extract port number using regex
					re := regexp.MustCompile(`Running on host port (\d+)`)
					matches := re.FindStringSubmatch(outputChunk)
					if len(matches) > 1 {
						foundPort := matches[1]
						portChan <- foundPort
						cmdMutex.Lock()
						deployment.Port = foundPort
						if err := db.UpdateDeployment(*deployment); err != nil {
							log.Printf("Error updating deployment port: %v", err)
						}
						cmdMutex.Unlock()
						log.Printf("[%s] Detected port: %s", name, foundPort)
					}
				}
			}
			if err != nil {
				if err != io.EOF {
					log.Printf("[%s run] read error: %v", name, err)
				}
				break
			}
		}
	}()

	if err := cmd.Start(); err != nil {
		http.Error(w, fmt.Sprintf("Error starting function: %v", err), http.StatusInternalServerError)
		return
	}

	// Store the cmd in our runningCmds map so we can stop it later
	cmdMutex.Lock()
	runningCmds[name] = cmd
	cmdMutex.Unlock()

	// Update the deployment status
	deployment.Status = "Running"
	if err := db.UpdateDeployment(*deployment); err != nil {
		http.Error(w, fmt.Sprintf("Error updating deployment status: %v", err), http.StatusInternalServerError)
		return
	}

	// Broadcast status update
	broadcastMessage(map[string]interface{}{
		"type": "status_update",
		"data": deployment,
	})

	// Wait for the port to be detected with a timeout
	var port string
	select {
	case port = <-portChan:
		// Port detected, update deployment
		deployment.Port = port
		if err := db.UpdateDeployment(*deployment); err != nil {
			log.Printf("Error updating deployment port: %v", err)
		}
	case <-time.After(10 * time.Second):
		// Timeout after 10 seconds
		log.Printf("Timeout waiting for port detection for %s", name)
		// Update status to indicate timeout
		deployment.Status = "Failed"
		deployment.Port = ""
		if err := db.UpdateDeployment(*deployment); err != nil {
			log.Printf("Error updating deployment status after timeout: %v", err)
		}
		// Broadcast status update
		broadcastMessage(map[string]interface{}{
			"type": "status_update",
			"data": deployment,
		})
	}

	// Wait in a separate goroutine
	go func() {
		err := cmd.Wait()
		if err != nil {
			log.Printf("Function [%s] exited with error: %v", name, err)
		} else {
			log.Printf("Function [%s] has stopped naturally.", name)
		}
		// Update status to Stopped once it ends
		cmdMutex.Lock()
		delete(runningCmds, name)
		cmdMutex.Unlock()

		deployment.Status = "Stopped"
		deployment.Port = ""
		if err := db.UpdateDeployment(*deployment); err != nil {
			log.Printf("Error updating deployment status: %v", err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"url": "http://localhost:" + deployment.Port,
	})
}

// stopHandler - stops the function if it's running
func stopHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/stop/")
	name = strings.TrimSuffix(name, "/")

	// Find the deployment
	deployment, err := db.GetDeployment(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving deployment: %v", err), http.StatusInternalServerError)
		return
	}
	if deployment == nil {
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	cmdMutex.Lock()
	cmd, exists := runningCmds[name]
	cmdMutex.Unlock()

	if !exists {
		http.Error(w, "Function is not running", http.StatusBadRequest)
		return
	}

	// Kill the process
	if err := cmd.Process.Kill(); err != nil {
		http.Error(w, fmt.Sprintf("Error stopping function: %v", err), http.StatusInternalServerError)
		return
	}

	// Clean up
	cmdMutex.Lock()
	delete(runningCmds, name)
	cmdMutex.Unlock()

	// Update status
	deployment.Status = "Stopped"
	deployment.Port = ""
	if err := db.UpdateDeployment(*deployment); err != nil {
		http.Error(w, fmt.Sprintf("Error updating deployment status: %v", err), http.StatusInternalServerError)
		return
	}

	// Broadcast status update
	broadcastMessage(map[string]interface{}{
		"type": "status_update",
		"data": deployment,
	})

	fmt.Fprintf(w, "Function %s stopped.", name)
}

// deploymentsHandler - returns the list of all deployments
func deploymentsHandler(w http.ResponseWriter, r *http.Request) {
	deployments, err := db.GetAllDeployments()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving deployments: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployments)
}

func getLanguageSpecificFiles(lang string) (string, string) {
	switch lang {
	case "python":
		return "func.py", "requirements.txt"
	case "go":
		return "handle.go", "go.mod"
	default:
		return "index.js", "package.json"
	}
}

// deploymentDetailHandler - returns deployment details (metadata + code + package content)
func deploymentDetailHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/deployments/")
	name = strings.TrimSuffix(name, "/")

	deployment, err := db.GetDeployment(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving deployment: %v", err), http.StatusInternalServerError)
		return
	}
	if deployment == nil {
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	// Attempt to read language-specific code/package files
	codeFile, pkgFile := getLanguageSpecificFiles(deployment.Language)
	codePath := filepath.Join("data", deployment.Name, codeFile)
	pkgPath := filepath.Join("data", deployment.Name, pkgFile)

	codeContent, _ := os.ReadFile(codePath)
	pkgContent, _ := os.ReadFile(pkgPath)

	// We can wrap deployment + file contents into a single response struct
	type deploymentDetail struct {
		db.Deployment
		Code    string `json:"code,omitempty"`
		Package string `json:"package,omitempty"`
	}

	detail := deploymentDetail{
		Deployment: *deployment,
		Code:       string(codeContent),
		Package:    string(pkgContent),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(detail); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}
