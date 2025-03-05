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
	"strings"
	"sync"
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

// In-memory list of deployments for demo
var deployments []Deployment
var deploymentCounter = 1

// runningCmds will track the active "func run" processes by function name
var runningCmds = make(map[string]*exec.Cmd)

// To avoid race conditions on runningCmds, wrap in a mutex
var cmdMutex sync.Mutex

func main() {
	http.HandleFunc("/create/", createHandler)
	http.HandleFunc("/upload/", uploadHandler)
	http.HandleFunc("/build/", buildHandler)

	// Adding the start and stop handlers
	http.HandleFunc("/start/", startHandler)
	http.HandleFunc("/stop/", stopHandler)

	// Instead of directly using deploymentsHandler on "/deployments/",
	// we now use handleDeployments to multiplex between a list and a single deployment.
	http.HandleFunc("/deployments/", handleDeployments)

	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
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
		http.Error(w, "Function with that name already exists", http.StatusBadRequest)
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

	deploymentID := fmt.Sprintf("%d", deploymentCounter)
	deployment := Deployment{
		ID:        deploymentID,
		Name:      name,
		Language:  language,
		Status:    "Stopped",
		CreatedAt: "TODO",
	}
	deployments = append(deployments, deployment)
	deploymentCounter++

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

	var dep *Deployment
	for i, d := range deployments {
		if d.Name == name {
			dep = &deployments[i]
			break
		}
	}
	if dep == nil {
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	codeFileName, packageFileName := getLanguageSpecificFiles(dep.Language)
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

// buildHandler - handles building and deploying via CLI func build and func deploy
func buildHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/build/")

	buildCmd := exec.Command("func", "build", name)
	buildCmd.Dir = "./data"
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error building function: %s", err), http.StatusInternalServerError)
		return
	}

	deployCmd := exec.Command("func", "deploy", name)
	deployCmd.Dir = "./data"
	deployOutput, err := deployCmd.CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deploying function: %s", err), http.StatusInternalServerError)
		return
	}

	// Update the Status to Running if found
	for i, d := range deployments {
		if d.Name == name {
			deployments[i].Status = "Running"
			break
		}
	}

	fmt.Fprintf(w, "Build and deploy successful.\nBuild output: %s\nDeploy output: %s", buildOutput, deployOutput)
}

// startHandler - starts the function using "func run" in a separate goroutine
func startHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/start/")
	name = strings.TrimSuffix(name, "/")

	// Find the deployment
	var dep *Deployment
	for i, d := range deployments {
		if d.Name == name {
			dep = &deployments[i]
			break
		}
	}
	if dep == nil {
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	// If it's already running, return early (or choose your own policy)
	if dep.Status == "Running" {
		http.Error(w, "Function is already started", http.StatusBadRequest)
		return
	}


	// Prepare the command: e.g. func run <name> --port <port>
	cmd := exec.Command("func", "run")
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

	go func() {
		reader := io.MultiReader(stdoutPipe, stderrPipe)
		buf := make([]byte, 1024)
		for {
			n, err := reader.Read(buf)
			if n > 0 {
				outputChunk := string(buf[:n])
				log.Printf("[func run output for %s]: %s", name, outputChunk)

				if strings.Contains(outputChunk, "Running on host port ") {
					parts := strings.Split(outputChunk, "Running on host port ")
					if len(parts) > 1 {
						fields := strings.Fields(parts[1])
						if len(fields) > 0 {
							foundPort := fields[0]
							// strip any trailing punctuation
							foundPort = strings.TrimRightFunc(foundPort, func(r rune) bool {
								return r < '0' || r > '9'
							})
							cmdMutex.Lock()
							dep.Port = foundPort
							cmdMutex.Unlock()
							log.Printf("[%s] Detected port: %s", name, foundPort)
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
		}
	}()

	// Start the command asynchronously in its own goroutine
	if err := cmd.Start(); err != nil {
		http.Error(w, fmt.Sprintf("Error starting function: %v", err), http.StatusInternalServerError)
		return
	}

	// Store the cmd in our runningCmds map so we can stop it later
	cmdMutex.Lock()
	runningCmds[name] = cmd
	cmdMutex.Unlock()

	// Update the deployment status and port
	dep.Status = "Running"
	dep.Port = port

	// Wait for it in a separate goroutine so it doesn't block this request
	go func() {
		// When the process ends, we can clean up or do other tasks
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

		dep.Status = "Stopped"
		dep.Port = ""
	}()

	// Return a JSON with the port or just a text message
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"url": "http://localhost:" + dep.Port,
	})
}

// stopHandler - stops the function if it's running
func stopHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/stop/")
	name = strings.TrimSuffix(name, "/")

	// Find the deployment
	var dep *Deployment
	for i, d := range deployments {
		if d.Name == name {
			dep = &deployments[i]
			break
		}
	}
	if dep == nil {
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
	dep.Status = "Stopped"
	dep.Port = ""

	fmt.Fprintf(w, "Function %s stopped.", name)
}

// deploymentsHandler - returns the list of all deployments
func deploymentsHandler(w http.ResponseWriter, r *http.Request) {
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

// deploymentDetailHandler - returns deployment details (metadata + code + package content) for /deployments/<name>/
func deploymentDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Strip the "/deployments/" prefix to get the actual name
	name := strings.TrimPrefix(r.URL.Path, "/deployments/")
	// Also remove any trailing slash if present
	name = strings.TrimSuffix(name, "/")

	// Find the deployment by name
	var dep *Deployment
	for i, d := range deployments {
		if d.Name == name {
			dep = &deployments[i]
			break
		}
	}
	if dep == nil {
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	// Attempt to read language-specific code/package files
	codeFile, pkgFile := getLanguageSpecificFiles(dep.Language)
	codePath := filepath.Join("data", dep.Name, codeFile)
	pkgPath := filepath.Join("data", dep.Name, pkgFile)

	codeContent, _ := os.ReadFile(codePath)
	pkgContent, _ := os.ReadFile(pkgPath)

	// We can wrap deployment + file contents into a single response struct
	type deploymentDetail struct {
		Deployment
		Code    string `json:"code,omitempty"`
		Package string `json:"package,omitempty"`
	}

	detail := deploymentDetail{
		Deployment: *dep,
		Code:       string(codeContent),
		Package:    string(pkgContent),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(detail); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}
