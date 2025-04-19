package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB initializes the SQLite database
func InitDB() error {
	// Create data directory if it doesn't exist
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("error creating data directory: %v", err)
	}

	// Open database connection
	dbPath := filepath.Join(dataDir, "deployments.db")
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	// Create deployments table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS deployments (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		language TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at TEXT NOT NULL,
		port TEXT
	);`

	_, err = DB.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error creating table: %v", err)
	}

	return nil
}

// CreateDeployment inserts a new deployment into the database
func CreateDeployment(deployment Deployment) error {
	query := `
	INSERT INTO deployments (id, name, language, status, created_at, port)
	VALUES (?, ?, ?, ?, ?, ?)`

	_, err := DB.Exec(query, deployment.ID, deployment.Name, deployment.Language, 
		deployment.Status, deployment.CreatedAt, deployment.Port)
	return err
}

// GetDeployment retrieves a deployment by name
func GetDeployment(name string) (*Deployment, error) {
	query := `
	SELECT id, name, language, status, created_at, port
	FROM deployments
	WHERE name = ?`

	row := DB.QueryRow(query, name)
	deployment := &Deployment{}
	err := row.Scan(&deployment.ID, &deployment.Name, &deployment.Language,
		&deployment.Status, &deployment.CreatedAt, &deployment.Port)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

// UpdateDeployment updates a deployment's status and port
func UpdateDeployment(deployment Deployment) error {
	query := `
	UPDATE deployments
	SET status = ?, port = ?
	WHERE name = ?`

	_, err := DB.Exec(query, deployment.Status, deployment.Port, deployment.Name)
	return err
}

// GetAllDeployments retrieves all deployments
func GetAllDeployments() ([]Deployment, error) {
	query := `
	SELECT id, name, language, status, created_at, port
	FROM deployments`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []Deployment
	for rows.Next() {
		var d Deployment
		err := rows.Scan(&d.ID, &d.Name, &d.Language, &d.Status, &d.CreatedAt, &d.Port)
		if err != nil {
			return nil, err
		}
		deployments = append(deployments, d)
	}

	return deployments, nil
}

// Deployment represents a function deployment
type Deployment struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Language  string `json:"language"`
	Status    string `json:"status"` // Can be: "Stopped", "Running", "Failed", "Building", "Built"
	CreatedAt string `json:"createdAt"`
	Port      string `json:"port,omitempty"`
} 