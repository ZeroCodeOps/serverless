package main

import (
    "fmt"
    "net/http"
    "log"
	"os/exec"
    "strings"
	
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
}
func buildHandler(w http.ResponseWriter, r *http.Request) {}