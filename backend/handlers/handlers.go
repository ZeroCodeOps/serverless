package handlers

import (
	"database/sql"
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

	"main/config"
	"main/db"
	"main/types"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Handlers struct {
	config      *config.Config
	db          *sql.DB
	upgrader    websocket.Upgrader
	clients     map[*websocket.Conn]bool
	clientsMux  sync.Mutex
	runningCmds map[string]*exec.Cmd
	cmdMux      sync.Mutex
}

func NewHandlers(cfg *config.Config, db *sql.DB) *Handlers {
	return &Handlers{
		config: cfg,
		db:     db,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients:     make(map[*websocket.Conn]bool),
		runningCmds: make(map[string]*exec.Cmd),
	}
}

func (h *Handlers) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/ws", h.wsHandler)
	mux.HandleFunc("/create/", h.createHandler)
	mux.HandleFunc("/upload/", h.uploadHandler)
	mux.HandleFunc("/build/", h.buildHandler)
	mux.HandleFunc("/start/", h.startHandler)
	mux.HandleFunc("/stop/", h.stopHandler)
	mux.HandleFunc("/deployments/", h.handleDeployments)
}

func (h *Handlers) broadcastMessage(message interface{}) {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	msg, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	for client := range h.clients {
		err := client.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Printf("Error sending message to client: %v", err)
			client.Close()
			delete(h.clients, client)
		}
	}
}

func (h *Handlers) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	defer conn.Close()

	h.clientsMux.Lock()
	h.clients[conn] = true
	h.clientsMux.Unlock()

	defer func() {
		h.clientsMux.Lock()
		delete(h.clients, conn)
		h.clientsMux.Unlock()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *Handlers) createHandler(w http.ResponseWriter, r *http.Request) {
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

	dataDir := h.config.Function.DataDir
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
	deployment := types.Deployment{
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

func (h *Handlers) uploadHandler(w http.ResponseWriter, r *http.Request) {
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
	if err := saveFile(codeFile, filepath.Join(h.config.Function.DataDir, name, codeFileName)); err != nil {
		http.Error(w, "Error saving code file", http.StatusInternalServerError)
		return
	}
	if err := saveFile(packageFile, filepath.Join(h.config.Function.DataDir, name, packageFileName)); err != nil {
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

func (h *Handlers) buildHandler(w http.ResponseWriter, r *http.Request) {
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
	go func(d *types.Deployment, fnName string) {
		// Step 1: Build
		buildCmd := exec.Command("func", "build", fnName, "--registry", h.config.Registry.Address)
		buildCmd.Dir = filepath.Join(h.config.Function.DataDir, fnName)
		buildOutput, err := buildCmd.CombinedOutput()
		if err != nil {
			log.Printf("[ERROR] Build for %s failed: %v\nOutput:\n%s", fnName, err, string(buildOutput))
			d.Status = "Failed"
			if err := db.UpdateDeployment(*d); err != nil {
				log.Printf("Error updating deployment status: %v", err)
			}
			// Broadcast status update
			h.broadcastMessage(map[string]interface{}{
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
		h.broadcastMessage(map[string]interface{}{
			"type": "status_update",
			"data": d,
		})
	}(deployment, name)

	// Immediately return a quick message
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "Build process started for '%s'. Status set to Building.\n", name)
}

func (h *Handlers) startHandler(w http.ResponseWriter, r *http.Request) {
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
	cmd := exec.Command("func", "run", "--registry", h.config.Registry.Address)
	cmd.Dir = filepath.Join(h.config.Function.DataDir, name)

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
						h.cmdMux.Lock()
						deployment.Port = foundPort
						if err := db.UpdateDeployment(*deployment); err != nil {
							log.Printf("Error updating deployment port: %v", err)
						}
						h.cmdMux.Unlock()
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
	h.cmdMux.Lock()
	h.runningCmds[name] = cmd
	h.cmdMux.Unlock()

	// Update the deployment status
	deployment.Status = "Running"
	if err := db.UpdateDeployment(*deployment); err != nil {
		http.Error(w, fmt.Sprintf("Error updating deployment status: %v", err), http.StatusInternalServerError)
		return
	}

	// Broadcast status update
	h.broadcastMessage(map[string]interface{}{
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
	case <-time.After(h.config.Function.PortDetectionTimeout):
		// Timeout after configured duration
		log.Printf("Timeout waiting for port detection for %s", name)
		// Update status to indicate timeout
		deployment.Status = "Failed"
		deployment.Port = ""
		if err := db.UpdateDeployment(*deployment); err != nil {
			log.Printf("Error updating deployment status after timeout: %v", err)
		}
		// Broadcast status update
		h.broadcastMessage(map[string]interface{}{
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
		h.cmdMux.Lock()
		delete(h.runningCmds, name)
		h.cmdMux.Unlock()

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

func (h *Handlers) stopHandler(w http.ResponseWriter, r *http.Request) {
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

	h.cmdMux.Lock()
	cmd, exists := h.runningCmds[name]
	h.cmdMux.Unlock()

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
	h.cmdMux.Lock()
	delete(h.runningCmds, name)
	h.cmdMux.Unlock()

	// Update status
	deployment.Status = "Stopped"
	deployment.Port = ""
	if err := db.UpdateDeployment(*deployment); err != nil {
		http.Error(w, fmt.Sprintf("Error updating deployment status: %v", err), http.StatusInternalServerError)
		return
	}

	// Broadcast status update
	h.broadcastMessage(map[string]interface{}{
		"type": "status_update",
		"data": deployment,
	})

	fmt.Fprintf(w, "Function %s stopped.", name)
}

func (h *Handlers) handleDeployments(w http.ResponseWriter, r *http.Request) {
	// If exact path is "/deployments/", list all deployments.
	if r.URL.Path == "/deployments/" || r.URL.Path == "/deployments" {
		h.deploymentsHandler(w, r)
		return
	}
	// Otherwise, treat it as a request for a named deployment's details.
	h.deploymentDetailHandler(w, r)
}

func (h *Handlers) deploymentsHandler(w http.ResponseWriter, r *http.Request) {
	deployments, err := db.GetAllDeployments()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving deployments: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployments)
}

func (h *Handlers) deploymentDetailHandler(w http.ResponseWriter, r *http.Request) {
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
	codePath := filepath.Join(h.config.Function.DataDir, deployment.Name, codeFile)
	pkgPath := filepath.Join(h.config.Function.DataDir, deployment.Name, pkgFile)

	codeContent, _ := os.ReadFile(codePath)
	pkgContent, _ := os.ReadFile(pkgPath)

	detail := types.DeploymentDetail{
		Deployment: *deployment,
		Code:       string(codeContent),
		Package:    string(pkgContent),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(detail); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
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
