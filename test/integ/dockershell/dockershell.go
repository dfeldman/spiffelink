package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/dfeldman/spiffelink/pkg/dockershell"
	"github.com/sirupsen/logrus"
)

func main() {
	// Command-line flags
	containerID := flag.String("container-id", "", "Docker container ID to run command in")
	flag.Parse()

	if *containerID == "" {
		log.Fatalf("Container ID must be provided")
	}

	// Create the DockerContext
	dc, err := dockershell.NewDockerContext(
		*containerID,
		logrus.New(), // or customize as per your requirements
	)
	if err != nil {
		log.Fatalf("Failed to create docker client: %v", err)
	}

	// Run a sample command
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	output, err := dc.RunCmd(ctx, "echo", []string{"Hello from Docker!"}, nil, 5*time.Second)
	if err != nil {
		log.Fatalf("Failed to run command in container: %v", err)
	}

	fmt.Println("Output from container:", output)
}
