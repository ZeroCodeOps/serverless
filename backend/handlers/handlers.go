package handlers

import (
	"bytes"
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

	"github.com/creack/pty"
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
	mux.HandleFunc("/delete/", h.deleteHandler)
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

	// Generate a new UUID for the deployment ID
	deploymentID := uuid.New().String()
	deployment := types.Deployment{
		ID:        deploymentID,
		Name:      name,
		Language:  language,
		Status:    "Creating",
		CreatedAt: time.Now().Format(time.RFC3339),
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

	// Update status to Stopped after creation
	deployment.Status = "Stopped"
	if err := db.CreateDeployment(deployment); err != nil {
		http.Error(w, fmt.Sprintf("Error saving deployment: %v", err), http.StatusInternalServerError)
		return
	}

	// Broadcast final status
	h.broadcastMessage(map[string]interface{}{
		"type": "create_deployment",
		"data": deployment,
	})

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
	deployment.Built = false
	if err := db.UpdateDeployment(*deployment); err != nil {
		http.Error(w, fmt.Sprintf("Error updating deployment status: %v", err), http.StatusInternalServerError)
		return
	}
	h.broadcastMessage(map[string]interface{}{
		"type": "status_update",
		"data": deployment,
	})
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

	// Broadcast building status
	h.broadcastMessage(map[string]interface{}{
		"type": "status_update",
		"data": deployment,
	})

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
		d.Status = "Stopped"
		d.Built = true
		if err := db.UpdateDeployment(*d); err != nil {
			log.Printf("Error updating deployment status: %v", err)
		}
		// Broadcast status update
		h.broadcastMessage(map[string]interface{}{
			"type": "build_complete",
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

	// Check if the function is built
	if !deployment.Built {
		http.Error(w, "Function needs to be built first", http.StatusBadRequest)
		return
	}

	// Update status to Starting
	deployment.Status = "Starting"
	if err := db.UpdateDeployment(*deployment); err != nil {
		http.Error(w, fmt.Sprintf("Error updating deployment status: %v", err), http.StatusInternalServerError)
		return
	}

	// Broadcast starting status
	h.broadcastMessage(map[string]interface{}{
		"type": "status_update",
		"data": deployment,
	})

	// Prepare the command: e.g. func run <name>
	cmd := exec.Command("func", "run", name, "--registry", h.config.Registry.Address)
	cmd.Dir = filepath.Join(h.config.Function.DataDir, name)

	// Create a pty
	ptmx, err := pty.Start(cmd)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error starting function with pty: %v", err), http.StatusInternalServerError)
		return
	}
	// defer ptmx.Close()

	// Store the cmd in our runningCmds map so we can stop it later
	h.cmdMux.Lock()
	h.runningCmds[name] = cmd
	h.cmdMux.Unlock()

	// Return 200 immediately to prevent timeout
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Function %s is starting...", name)

	// Start reading output in a goroutine
	go func() {
		buf := make([]byte, 4096)
		startTime := time.Now()
		timeout := 30 * time.Second
		errorBuffer := bytes.NewBuffer(nil)
		port := ""

		for {
			n, err := ptmx.Read(buf)
			if n > 0 {
				outputChunk := string(buf[:n])
				log.Printf("[func run output for %s]: %s", name, outputChunk)
				errorBuffer.Write(buf[:n])

				// Try to extract port from the output
				if port == "" {
					patterns := []string{
						`Running on host port (\d+)`,
						`port (\d+)`,
						`listening on port (\d+)`,
						`started on port (\d+)`,
					}

					for _, pattern := range patterns {
						re := regexp.MustCompile(pattern)
						matches := re.FindStringSubmatch(outputChunk)
						if len(matches) > 1 {
							port = matches[1]
							log.Printf("[DEBUG] Found port using pattern '%s': %s", pattern, port)
							h.cmdMux.Lock()
							deployment.Port = port
							deployment.Status = "Running"
							if err := db.UpdateDeployment(*deployment); err != nil {
								log.Printf("Error updating deployment port: %v", err)
							}
							h.cmdMux.Unlock()
							// Broadcast status update with port
							h.broadcastMessage(map[string]interface{}{
								"type": "status_update",
								"data": deployment,
							})
							break
						}
					}
				}
			}
			if err != nil {
				if err != io.EOF {
					log.Printf("[%s run] read error: %v", name, err)
				}
				break
			}

			// Check timeout
			if time.Since(startTime) > timeout {
				log.Printf("[%s] Function startup timeout after %v", name, timeout)
				if port == "" {
					log.Printf("[%s] Warning: No port detected within timeout period", name)
					// Update status to indicate timeout
					h.cmdMux.Lock()
					deployment.Status = "Failed"
					if err := db.UpdateDeployment(*deployment); err != nil {
						log.Printf("Error updating deployment status: %v", err)
					}
					h.cmdMux.Unlock()
					// Broadcast failed status
					h.broadcastMessage(map[string]interface{}{
						"type": "status_update",
						"data": deployment,
					})
					// Kill the process if it's still running
					if cmd.Process != nil {
						cmd.Process.Kill()
					}
					h.cmdMux.Lock()
					delete(h.runningCmds, name)
					h.cmdMux.Unlock()
				}
				return
			}
		}

		// If we get here, the process has exited
		if port == "" {
			log.Printf("[%s] Function exited without detecting port. Error output: %s", name, errorBuffer.String())
			h.cmdMux.Lock()
			deployment.Status = "Failed"
			if err := db.UpdateDeployment(*deployment); err != nil {
				log.Printf("Error updating deployment status: %v", err)
			}
			h.cmdMux.Unlock()
			// Broadcast failed status
			h.broadcastMessage(map[string]interface{}{
				"type": "status_update",
				"data": deployment,
			})
		}
	}()
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

	// Check if the function is running
	if deployment.Status != "Running" {
		http.Error(w, "Function is not running", http.StatusBadRequest)
		return
	}

	h.cmdMux.Lock()
	cmd, exists := h.runningCmds[name]
	h.cmdMux.Unlock()

	if !exists {
		// If command not found but status is Running, update status to Stopped
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
		return
	}

	// Send SIGINT (Ctrl+C) to the process and its children
	if cmd.Process != nil {
		// First try to send SIGINT to the process group
		if err := exec.Command("pkill", "-INT", "-P", fmt.Sprintf("%d", cmd.Process.Pid)).Run(); err != nil {
			log.Printf("Error sending SIGINT to process group: %v", err)
		}
		// Then send SIGINT to the main process
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			log.Printf("Error sending SIGINT to process: %v", err)
		}

		// Wait for the process to exit with a timeout
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case err := <-done:
			if err != nil && err.Error() != "signal: interrupt" {
				log.Printf("Error waiting for process: %v", err)
			}
		case <-time.After(10 * time.Second):
			// If process doesn't exit within 10 seconds, force kill it
			log.Printf("Process did not exit after SIGINT, forcing kill")
			if err := cmd.Process.Kill(); err != nil {
				log.Printf("Error killing process: %v", err)
			}
		}
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

func (h *Handlers) deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/delete/")
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

	// If the function is running, stop it first
	if deployment.Status == "Running" {
		h.cmdMux.Lock()
		cmd, exists := h.runningCmds[name]
		h.cmdMux.Unlock()

		if exists && cmd.Process != nil {
			if err := cmd.Process.Kill(); err != nil {
				log.Printf("Error killing process: %v", err)
			}
			h.cmdMux.Lock()
			delete(h.runningCmds, name)
			h.cmdMux.Unlock()
		}
	}

	// Delete the deployment from the database
	if err := db.DeleteDeployment(name); err != nil {
		http.Error(w, fmt.Sprintf("Error deleting deployment: %v", err), http.StatusInternalServerError)
		return
	}

	// Delete the function directory
	functionDir := filepath.Join(h.config.Function.DataDir, name)
	if err := os.RemoveAll(functionDir); err != nil {
		log.Printf("Error deleting function directory: %v", err)
	}

	// Broadcast deletion
	h.broadcastMessage(map[string]interface{}{
		"type": "deployment_deleted",
		"data": map[string]string{
			"name": name,
		},
	})

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Deployment %s deleted successfully", name)
}
