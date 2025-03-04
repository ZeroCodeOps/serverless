package main

import (
    "fmt"
    "net/http"
    "log"
	"os/exec"
    "strings"
	"io"
	"os"
	"path/filepath"
	
)
func main() {
	http.HandleFunc("/", createHandler)
    http.HandleFunc("/upload/", uploadHandler)
    http.HandleFunc("/build/", buildHandler)

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

    language := parts[1] // Extract language
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

    

    cmd := exec.Command("func", "create", "-l", language, name)
    output, err := cmd.CombinedOutput()
	log.Printf("Command Output: %s", output) // Log output for debugging

    if err != nil {
		log.Printf("Error executing command: %v", err)

        http.Error(w, fmt.Sprintf("Error creating function: %s\nOutput: %s", err, output), http.StatusInternalServerError)
        return
    }

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
    buildOutput, err := buildCmd.CombinedOutput()
    if err != nil {
        http.Error(w, fmt.Sprintf("Error building function: %s", err), http.StatusInternalServerError)
        return
    }

    // Run func deploy
    deployCmd := exec.Command("func", "deploy", name)
    deployOutput, err := deployCmd.CombinedOutput()
    if err != nil {
        http.Error(w, fmt.Sprintf("Error deploying function: %s", err), http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Build and deploy successful.\nBuild output: %s\nDeploy output: %s", buildOutput, deployOutput)
}