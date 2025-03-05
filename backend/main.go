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
)

type Deployment struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Language  string `json:"language"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

var deployments []Deployment
var deploymentCounter = 1

func main() {
	http.HandleFunc("/create/", createHandler)
	http.HandleFunc("/upload/", uploadHandler)
	http.HandleFunc("/build/", buildHandler)
	http.HandleFunc("/deployments/", deploymentsHandler)

	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received Request at:", r.URL.Path)

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	language := parts[2] // Extract language
	if language == "" {
		http.Error(w, "Language is required", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// fmt.Fprintf(w, "Function created for language: %s, name: %s", language, name)

	dataDir := "./data"

	// Create the data directory if it doesn't exist
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		err := os.Mkdir(dataDir, 0755) // Corrected permissions
		if err != nil {
			log.Printf("Error creating directory: %v", err)
			http.Error(w, fmt.Sprintf("Error creating directory: %s", err), http.StatusInternalServerError)
			return
		}
	}

	// Set the full path for the function directory
	functionDir := filepath.Join(dataDir, name)

	//Check if the function already exists
	if _, err := os.Stat(functionDir); !os.IsNotExist(err) {
		http.Error(w, "Function with that name already exists", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("func", "create", "-l", language, name)
	cmd.Dir = dataDir // Set the working directory
	output, err := cmd.CombinedOutput()
	log.Printf("Command Output: %s", output) // Log output for debugging

	if err != nil {
		log.Printf("Error executing command: %v", err)

		http.Error(w, fmt.Sprintf("Error creating function: %s\nOutput: %s", err, output), http.StatusInternalServerError)
		return
	}

	//Create Deployment object
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

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/upload/")
	name = strings.TrimSuffix(name, "/")

	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Handle code file
	codeFile, _, err := r.FormFile("code")
	if err != nil {
		http.Error(w, "Error retrieving code file", http.StatusBadRequest)
		return
	}
	defer codeFile.Close()

	// Handle package file
	packageFile, _, err := r.FormFile("package")
	if err != nil {
		http.Error(w, "Error retrieving package file", http.StatusBadRequest)
		return
	}
	defer packageFile.Close()

	// Save files
	if err := saveFile(codeFile, filepath.Join(name, "main.go")); err != nil {
		http.Error(w, "Error saving code file", http.StatusInternalServerError)
		return
	}

	if err := saveFile(packageFile, filepath.Join(name, "package.json")); err != nil {
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

func buildHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/build/")

	// Run func build
	buildCmd := exec.Command("func", "build", name)
	buildCmd.Dir = "./data"
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error building function: %s", err), http.StatusInternalServerError)
		return
	}

	// Run func deploy
	deployCmd := exec.Command("func", "deploy", name)
	deployCmd.Dir = "./data"
	deployOutput, err := deployCmd.CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deploying function: %s", err), http.StatusInternalServerError)
		return
	}
	//Find the deployment
	var deployment *Deployment
	for i, d := range deployments {
		if d.Name == name {
			deployment = &deployments[i]
			break
		}
	}
	if deployment != nil {
		deployment.Status = "Running"
	}

	fmt.Fprintf(w, "Build and deploy successful.\nBuild output: %s\nDeploy output: %s", buildOutput, deployOutput)
}
func deploymentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployments)
}
