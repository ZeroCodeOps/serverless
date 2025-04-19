#!/bin/bash

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if registry container exists
if ! docker ps -a | grep -q "local-registry"; then
    echo "Creating local registry container..."
    docker run -d \
        -p 5000:5000 \
        --name local-registry \
        --restart=always \
        registry:2
else
    # Check if registry container is running
    if ! docker ps | grep -q "local-registry"; then
        echo "Starting local registry container..."
        docker start local-registry
    else
        echo "Local registry is already running."
    fi
fi

echo "Local registry is ready at localhost:5000" 