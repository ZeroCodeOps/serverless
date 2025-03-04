package main

import (
    "fmt"
    "net/http"
    "log"
	
)
func main() {
	http.HandleFunc("/", createHandler)
    http.HandleFunc("/upload/", uploadHandler)
    http.HandleFunc("/build/", buildHandler)

    fmt.Println("Server starting on port 8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
func createHandler(w http.ResponseWriter, r *http.Request) {}
func uploadHandler(w http.ResponseWriter, r *http.Request) {
}
func buildHandler(w http.ResponseWriter, r *http.Request) {}