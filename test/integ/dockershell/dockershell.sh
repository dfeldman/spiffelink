#!/bin/sh

# Pull the Alpine image
docker pull alpine:latest

# Run a new Docker container in detached mode and capture its ID
CONTAINER_ID=$(docker run -d alpine:latest tail -f /dev/null)

# If the container didn't start, exit
if [ -z "$CONTAINER_ID" ]; then
    echo "Failed to start the Docker container."
    exit 1
fi

echo "Started Docker container with ID: $CONTAINER_ID"

# Build and run the integration test
go build -o integration_test dockershell.go
./integration_test --container-id=$CONTAINER_ID

# Cleanup: Stop and remove the container
docker rm -f $CONTAINER_ID
