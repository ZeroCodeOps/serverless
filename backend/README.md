# Serverless Backend

This is the backend service for the serverless function deployment system.

## Prerequisites

- Docker installed and running
- Go 1.21 or later
- SQLite3

## Setup

1. Start the local Docker registry:
```bash
./start-registry.sh
```

2. Build and run the backend:
```bash
go mod tidy
go run main.go
```

The backend will start on port 8080.

## Local Registry

The backend uses a local Docker registry (localhost:5000) to store function images. This is required for building and running functions.

The registry is managed by the `start-registry.sh` script, which:
- Creates a local registry container if it doesn't exist
- Starts the registry if it's not running
- Uses the official Docker registry image (registry:2)

## API Endpoints

- `POST /create/{language}` - Create a new function
- `POST /upload/{name}` - Upload function code and package files
- `POST /build/{name}` - Build a function
- `POST /start/{name}` - Start a function
- `POST /stop/{name}` - Stop a function
- `GET /deployments/` - List all deployments
- `GET /deployments/{name}` - Get deployment details 