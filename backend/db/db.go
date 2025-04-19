package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"main/types"

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
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS deployments (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			language TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL,
			port TEXT,
			built BOOLEAN NOT NULL DEFAULT FALSE
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating deployments table: %v", err)
	}

	return nil
}

// CreateDeployment inserts a new deployment into the database
func CreateDeployment(d types.Deployment) error {
	_, err := DB.Exec(`
		INSERT INTO deployments (id, name, language, status, created_at, port, built)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, d.ID, d.Name, d.Language, d.Status, d.CreatedAt, d.Port, d.Built)
	if err != nil {
		return fmt.Errorf("error creating deployment: %v", err)
	}
	return nil
}

// GetDeployment retrieves a deployment by name
func GetDeployment(name string) (*types.Deployment, error) {
	var d types.Deployment
	err := DB.QueryRow(`
		SELECT id, name, language, status, created_at, port, built
		FROM deployments
		WHERE name = ?
	`, name).Scan(&d.ID, &d.Name, &d.Language, &d.Status, &d.CreatedAt, &d.Port, &d.Built)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting deployment: %v", err)
	}
	return &d, nil
}

// UpdateDeployment updates a deployment's status and port
func UpdateDeployment(d types.Deployment) error {
	_, err := DB.Exec(`
		UPDATE deployments
		SET status = ?, port = ?, built = ?
		WHERE name = ?
	`, d.Status, d.Port, d.Built, d.Name)
	if err != nil {
		return fmt.Errorf("error updating deployment: %v", err)
	}
	return nil
}

// GetAllDeployments retrieves all deployments
func GetAllDeployments() ([]types.Deployment, error) {
	rows, err := DB.Query(`
		SELECT id, name, language, status, created_at, port, built
		FROM deployments
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("error querying deployments: %v", err)
	}
	defer rows.Close()

	var deployments []types.Deployment
	for rows.Next() {
		var d types.Deployment
		err := rows.Scan(&d.ID, &d.Name, &d.Language, &d.Status, &d.CreatedAt, &d.Port, &d.Built)
		if err != nil {
			return nil, fmt.Errorf("error scanning deployment: %v", err)
		}
		deployments = append(deployments, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deployments: %v", err)
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
	Built     bool   `json:"built"`
}
